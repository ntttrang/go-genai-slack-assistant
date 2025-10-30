package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"go.uber.org/zap"
)

// mockEventProcessor implements slack.EventProcessor for testing
type mockEventProcessor struct {
	processedEvents []string // stores message timestamps in order
	mu              sync.Mutex
	processDelay    time.Duration // simulate processing time
	callCount       int32
}

func newMockEventProcessor(delay time.Duration) *mockEventProcessor {
	return &mockEventProcessor{
		processedEvents: []string{},
		processDelay:    delay,
	}
}

func (m *mockEventProcessor) ProcessEvent(ctx context.Context, payload map[string]interface{}) {
	atomic.AddInt32(&m.callCount, 1)
	
	// Simulate processing time
	if m.processDelay > 0 {
		time.Sleep(m.processDelay)
	}

	// Extract timestamp from payload for tracking order
	if event, ok := payload["event"].(map[string]interface{}); ok {
		if ts, ok := event["ts"].(string); ok {
			m.mu.Lock()
			m.processedEvents = append(m.processedEvents, ts)
			m.mu.Unlock()
		}
	}
}

func (m *mockEventProcessor) getProcessedEvents() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.processedEvents))
	copy(result, m.processedEvents)
	return result
}

func (m *mockEventProcessor) getCallCount() int32 {
	return atomic.LoadInt32(&m.callCount)
}

func TestWorkerPool_MessageOrdering(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(10 * time.Millisecond)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Create 3 messages from same user in same channel
	events := []*model.MessageEvent{
		{
			EventID:    "evt1",
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  "1000.001",
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
			ReceivedAt: time.Now(),
		},
		{
			EventID:    "evt2",
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  "1000.002",
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.002"}},
			ReceivedAt: time.Now(),
		},
		{
			EventID:    "evt3",
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  "1000.003",
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.003"}},
			ReceivedAt: time.Now(),
		},
	}

	// Enqueue all messages
	for _, event := range events {
		workerPool.Enqueue(event)
	}

	// Wait for processing to complete
	time.Sleep(200 * time.Millisecond)

	// Verify order
	processed := processor.getProcessedEvents()
	if len(processed) != 3 {
		t.Fatalf("Expected 3 processed events, got %d", len(processed))
	}

	expected := []string{"1000.001", "1000.002", "1000.003"}
	for i, ts := range expected {
		if processed[i] != ts {
			t.Errorf("Expected event %d to have timestamp %s, got %s", i, ts, processed[i])
		}
	}
}

func TestWorkerPool_ParallelProcessing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(50 * time.Millisecond)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Create messages from two different users in the same channel
	// With channel-level ordering, they will be processed sequentially in a single queue
	event1 := &model.MessageEvent{
		ChannelID:  "C123",
		UserID:     "U111",
		MessageTS:  "1000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
		ReceivedAt: time.Now(),
	}

	event2 := &model.MessageEvent{
		ChannelID:  "C123",
		UserID:     "U222",
		MessageTS:  "2000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "2000.001"}},
		ReceivedAt: time.Now(),
	}

	// Enqueue both
	start := time.Now()
	workerPool.Enqueue(event1)
	workerPool.Enqueue(event2)

	// Wait for both to complete (sequential processing: ~100ms)
	time.Sleep(150 * time.Millisecond)
	elapsed := time.Since(start)

	// Both should have been processed
	if processor.getCallCount() != 2 {
		t.Fatalf("Expected 2 calls, got %d", processor.getCallCount())
	}

	t.Logf("Sequential processing (same channel) completed in %v", elapsed)

	// Verify we have 1 active queue (same channel, so single queue)
	if workerPool.GetQueueCount() != 1 {
		t.Errorf("Expected 1 active queue for same channel, got %d", workerPool.GetQueueCount())
	}
}

func TestWorkerPool_WorkerSpawning(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(0)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Initially no queues
	if workerPool.GetQueueCount() != 0 {
		t.Errorf("Expected 0 queues initially, got %d", workerPool.GetQueueCount())
	}

	// Enqueue first message - should create queue and spawn worker
	event1 := &model.MessageEvent{
		ChannelID:  "C123",
		UserID:     "U456",
		MessageTS:  "1000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
		ReceivedAt: time.Now(),
	}
	workerPool.Enqueue(event1)

	// Give worker time to spawn
	time.Sleep(10 * time.Millisecond)

	// Should have 1 queue
	if workerPool.GetQueueCount() != 1 {
		t.Errorf("Expected 1 queue after first message, got %d", workerPool.GetQueueCount())
	}

	// Enqueue second message from same user - should reuse queue
	event2 := &model.MessageEvent{
		ChannelID:  "C123",
		UserID:     "U456",
		MessageTS:  "1000.002",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.002"}},
		ReceivedAt: time.Now(),
	}
	workerPool.Enqueue(event2)

	time.Sleep(10 * time.Millisecond)

	// Should still have 1 queue
	if workerPool.GetQueueCount() != 1 {
		t.Errorf("Expected 1 queue after second message from same user, got %d", workerPool.GetQueueCount())
	}
}

func TestWorkerPool_IdleCleanup(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(0)
	// Use very short idle timeout for testing
	workerPool := NewWorkerPool(processor, 10, 100*time.Millisecond, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Enqueue a message
	event := &model.MessageEvent{
		ChannelID:  "C123",
		UserID:     "U456",
		MessageTS:  "1000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
		ReceivedAt: time.Now(),
	}
	workerPool.Enqueue(event)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)

	// Should have 1 queue
	if workerPool.GetQueueCount() != 1 {
		t.Errorf("Expected 1 queue, got %d", workerPool.GetQueueCount())
	}

	// Wait for idle timeout (100ms + buffer)
	time.Sleep(150 * time.Millisecond)

	// Queue should be cleaned up
	if workerPool.GetQueueCount() != 0 {
		t.Errorf("Expected 0 queues after idle timeout, got %d", workerPool.GetQueueCount())
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(20 * time.Millisecond)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)

	// Enqueue 5 messages
	for i := 0; i < 5; i++ {
		event := &model.MessageEvent{
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  time.Now().Format("1000.00") + string(rune('0'+i)),
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": time.Now().Format("1000.00") + string(rune('0'+i))}},
			ReceivedAt: time.Now(),
		}
		workerPool.Enqueue(event)
	}

	// Give a moment for messages to be enqueued
	time.Sleep(10 * time.Millisecond)

	// Shutdown - should drain all messages
	err := workerPool.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	// All 5 messages should be processed
	if processor.getCallCount() != 5 {
		t.Errorf("Expected 5 messages processed after shutdown, got %d", processor.getCallCount())
	}
}

func TestWorkerPool_ShutdownTimeout(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Very slow processor
	processor := newMockEventProcessor(1 * time.Second)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)

	// Enqueue 3 messages
	for i := 0; i < 3; i++ {
		event := &model.MessageEvent{
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  "1000.00" + string(rune('0'+i)),
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.00" + string(rune('0'+i))}},
			ReceivedAt: time.Now(),
		}
		workerPool.Enqueue(event)
	}

	time.Sleep(10 * time.Millisecond)

	// Shutdown with very short timeout - should timeout
	err := workerPool.Shutdown(50 * time.Millisecond)
	if err == nil {
		t.Error("Expected shutdown to timeout, but it didn't")
	}
}

func TestWorkerPool_BufferFull(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Slow processor
	processor := newMockEventProcessor(100 * time.Millisecond)
	// Small buffer
	workerPool := NewWorkerPool(processor, 2, 1*time.Minute, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Enqueue 3 messages rapidly (buffer is 2)
	for i := 0; i < 3; i++ {
		event := &model.MessageEvent{
			ChannelID:  "C123",
			UserID:     "U456",
			MessageTS:  "1000.00" + string(rune('0'+i)),
			Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.00" + string(rune('0'+i))}},
			ReceivedAt: time.Now(),
		}
		workerPool.Enqueue(event)
	}

	// Wait for all to be processed
	time.Sleep(400 * time.Millisecond)

	// All 3 should eventually be processed
	if processor.getCallCount() != 3 {
		t.Errorf("Expected 3 messages processed, got %d", processor.getCallCount())
	}
}

func TestWorkerPool_GetQueueKey(t *testing.T) {
	event := &model.MessageEvent{
		ChannelID: "C123",
		UserID:    "U456",
	}

	key := event.GetQueueKey()
	expected := "C123"

	if key != expected {
		t.Errorf("Expected queue key %s, got %s", expected, key)
	}
}

func TestWorkerPool_EventDeduplication(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	processor := newMockEventProcessor(0)
	workerPool := NewWorkerPool(processor, 10, 1*time.Minute, logger)
	defer func() {
		_ = workerPool.Shutdown(5 * time.Second)
	}()

	// Create two events with the same event_id (simulating Slack retry)
	event1 := &model.MessageEvent{
		EventID:    "evt123",
		ChannelID:  "C123",
		UserID:     "U111",
		MessageTS:  "1000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
		ReceivedAt: time.Now(),
	}

	event2 := &model.MessageEvent{
		EventID:    "evt123", // Same event ID - should be deduplicated
		ChannelID:  "C123",
		UserID:     "U111",
		MessageTS:  "1000.001",
		Payload:    map[string]interface{}{"event": map[string]interface{}{"ts": "1000.001"}},
		ReceivedAt: time.Now(),
	}

	// Enqueue both events
	workerPool.Enqueue(event1)
	workerPool.Enqueue(event2)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Only first event should have been processed (second was deduplicated)
	if processor.getCallCount() != 1 {
		t.Errorf("Expected 1 processed event (second was deduplicated), got %d", processor.getCallCount())
	}
}
