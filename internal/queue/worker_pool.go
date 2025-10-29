package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service/slack"
	"go.uber.org/zap"
)

// WorkerPool manages message queues and workers for ordered message processing.
// Each unique channel gets its own queue and worker goroutine.
type WorkerPool struct {
	queues       sync.Map              // map[string]chan *model.MessageEvent
	seenEvents   sync.Map              // map[string]bool for deduplication by event_id
	processor    slack.EventProcessor  // processes events synchronously
	bufferSize   int                   // buffer size for each queue channel
	idleTimeout  time.Duration         // time after which idle workers are cleaned up
	shutdown     chan struct{}         // signal for graceful shutdown
	wg           sync.WaitGroup        // wait for all workers to finish
	logger       *zap.Logger
}

// NewWorkerPool creates a new worker pool for processing message events.
func NewWorkerPool(
	processor slack.EventProcessor,
	bufferSize int,
	idleTimeout time.Duration,
	logger *zap.Logger,
) *WorkerPool {
	return &WorkerPool{
		queues:      sync.Map{},
		processor:   processor,
		bufferSize:  bufferSize,
		idleTimeout: idleTimeout,
		shutdown:    make(chan struct{}),
		logger:      logger,
	}
}

// Enqueue adds a message event to the appropriate queue based on channel.
// If no queue exists for this channel, a new one is created and a worker is spawned.
// Duplicate events (same event_id) are silently dropped to prevent processing duplicates from Slack retries.
func (wp *WorkerPool) Enqueue(event *model.MessageEvent) {
	// Deduplicate by event_id
	if event.EventID != "" {
		if _, exists := wp.seenEvents.LoadOrStore(event.EventID, true); exists {
			wp.logger.Warn("Duplicate event detected, dropping (SKIPPED)",
				zap.String("event_id", event.EventID),
				zap.String("channel_id", event.ChannelID),
				zap.String("message_ts", event.MessageTS),
				zap.String("user_id", event.UserID))
			return
		}
		wp.logger.Debug("New event_id (ACCEPTED)",
			zap.String("event_id", event.EventID),
			zap.Uint64("sequence", event.Sequence))
	} else {
		wp.logger.Warn("Event with empty event_id detected",
			zap.String("channel_id", event.ChannelID),
			zap.String("message_ts", event.MessageTS),
			zap.Uint64("sequence", event.Sequence))
	}

	queueKey := event.GetQueueKey()

	// Get existing queue or create new one
	queueInterface, loaded := wp.queues.LoadOrStore(queueKey, make(chan *model.MessageEvent, wp.bufferSize))
	eventChan := queueInterface.(chan *model.MessageEvent)

	// If this is a new queue, spawn a worker goroutine
	if !loaded {
		wp.wg.Add(1)
		go wp.worker(queueKey, eventChan)
		wp.logger.Info("Started new worker for channel queue",
			zap.String("channel_id", event.ChannelID))
	}

	// Send message to channel
	select {
	case eventChan <- event:
		wp.logger.Debug("Message enqueued",
			zap.String("queue_key", queueKey),
			zap.String("message_ts", event.MessageTS),
			zap.String("event_id", event.EventID))
	case <-wp.shutdown:
		wp.logger.Warn("Dropping message, shutdown in progress",
			zap.String("queue_key", queueKey))
	default:
		// Buffer full - block until space available
		wp.logger.Warn("Queue buffer full, blocking until space available",
			zap.String("queue_key", queueKey),
			zap.Int("buffer_size", wp.bufferSize))
		eventChan <- event
	}
}

// worker processes messages from a single queue sequentially.
// It exits when idle timeout is reached or shutdown is signaled.
func (wp *WorkerPool) worker(queueKey string, eventChan chan *model.MessageEvent) {
	defer wp.wg.Done()
	defer wp.cleanup(queueKey, eventChan)

	idleTimer := time.NewTimer(wp.idleTimeout)
	defer idleTimer.Stop()

	wp.logger.Info("Worker started", zap.String("queue_key", queueKey))

	for {
		select {
		case event := <-eventChan:
			// Reset idle timer - we have work to do
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(wp.idleTimeout)

			// Process event synchronously (ensures ordering)
			wp.logger.Info("Processing event (SEQUENTIAL)",
				zap.String("queue_key", queueKey),
				zap.String("message_ts", event.MessageTS),
				zap.Uint64("sequence", event.Sequence),
				zap.String("user_id", event.UserID),
				zap.Time("received_at", event.ReceivedAt))

			ctx := context.Background()
			wp.processor.ProcessEvent(ctx, event.Payload)

			wp.logger.Info("Event processed (COMPLETE)",
				zap.String("queue_key", queueKey),
				zap.String("message_ts", event.MessageTS),
				zap.Uint64("sequence", event.Sequence))

		case <-idleTimer.C:
			// No messages for idleTimeout duration, exit worker
			wp.logger.Info("Worker idle timeout reached, exiting",
				zap.String("queue_key", queueKey),
				zap.Duration("idle_timeout", wp.idleTimeout))
			return

		case <-wp.shutdown:
			// Graceful shutdown: drain remaining messages
			wp.logger.Info("Worker received shutdown signal, draining queue",
				zap.String("queue_key", queueKey))
			wp.drainQueue(queueKey, eventChan)
			return
		}
	}
}

// drainQueue processes all remaining messages in the queue during shutdown.
func (wp *WorkerPool) drainQueue(queueKey string, eventChan chan *model.MessageEvent) {
	drained := 0
	for {
		select {
		case event := <-eventChan:
			wp.logger.Debug("Draining event",
				zap.String("queue_key", queueKey),
				zap.String("message_ts", event.MessageTS))
			ctx := context.Background()
			wp.processor.ProcessEvent(ctx, event.Payload)
			drained++
		default:
			// Queue is empty
			if drained > 0 {
				wp.logger.Info("Queue drained",
					zap.String("queue_key", queueKey),
					zap.Int("messages_processed", drained))
			}
			return
		}
	}
}

// cleanup closes the channel and removes it from the map.
func (wp *WorkerPool) cleanup(queueKey string, eventChan chan *model.MessageEvent) {
	close(eventChan)
	wp.queues.Delete(queueKey)
	
	// Note: We don't clean up seenEvents here since they need to persist across worker lifecycle
	// to handle Slack's retry window. Memory impact is minimal as events only accumulate briefly.
	
	wp.logger.Info("Worker cleaned up",
		zap.String("queue_key", queueKey))
}

// Shutdown gracefully stops all workers and waits for them to finish.
// It waits up to the specified timeout for all workers to drain their queues.
func (wp *WorkerPool) Shutdown(timeout time.Duration) error {
	wp.logger.Info("Starting WorkerPool shutdown",
		zap.Duration("timeout", timeout))

	// Signal all workers to stop
	close(wp.shutdown)

	// Wait for all workers to drain and finish
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("All workers stopped gracefully")
		return nil
	case <-time.After(timeout):
		wp.logger.Warn("Shutdown timeout reached, some messages may be lost",
			zap.Duration("timeout", timeout))
		return fmt.Errorf("shutdown timeout after %v", timeout)
	}
}

// GetQueueCount returns the current number of active queues (for monitoring/testing).
func (wp *WorkerPool) GetQueueCount() int {
	count := 0
	wp.queues.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
