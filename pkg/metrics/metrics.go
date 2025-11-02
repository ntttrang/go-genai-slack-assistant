package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	TranslationRequests map[string]int64
	UserRequests        map[string]int64
	ChannelRequests     map[string]int64

	APILatencies []time.Duration
	SuccessCount int64
	FailureCount int64

	CacheHits   int64
	CacheMisses int64

	GeminiTokensUsed int64

	ErrorsByType map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		TranslationRequests: make(map[string]int64),
		UserRequests:        make(map[string]int64),
		ChannelRequests:     make(map[string]int64),
		APILatencies:        make([]time.Duration, 0),
		ErrorsByType:        make(map[string]int64),
	}
}

func (m *Metrics) RecordTranslationRequest(userID, channelID string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TranslationRequests[userID]++
	m.UserRequests[userID]++
	m.ChannelRequests[channelID]++
	m.APILatencies = append(m.APILatencies, duration)

	if success {
		m.SuccessCount++
	} else {
		m.FailureCount++
	}
}

func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheHits++
}

func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMisses++
}

func (m *Metrics) RecordGeminiTokens(tokens int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GeminiTokensUsed += tokens
}

func (m *Metrics) RecordError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsByType[errorType]++
}

func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_requests"] = m.getTotalRequests()
	stats["success_count"] = m.SuccessCount
	stats["failure_count"] = m.FailureCount
	stats["success_rate"] = m.getSuccessRate()
	stats["average_latency_ms"] = m.getAverageLatency()
	stats["cache_hit_rate"] = m.getCacheHitRate()
	stats["total_gemini_tokens"] = m.GeminiTokensUsed
	stats["errors_by_type"] = m.ErrorsByType
	stats["top_users"] = m.getTopUsers()
	stats["top_channels"] = m.getTopChannels()

	return stats
}

func (m *Metrics) getTotalRequests() int64 {
	return m.SuccessCount + m.FailureCount
}

func (m *Metrics) getSuccessRate() float64 {
	total := m.getTotalRequests()
	if total == 0 {
		return 0.0
	}
	return float64(m.SuccessCount) / float64(total) * 100
}

func (m *Metrics) getAverageLatency() float64 {
	if len(m.APILatencies) == 0 {
		return 0.0
	}

	var totalDuration time.Duration
	for _, d := range m.APILatencies {
		totalDuration += d
	}
	return float64(totalDuration.Milliseconds()) / float64(len(m.APILatencies))
}

func (m *Metrics) getCacheHitRate() float64 {
	total := m.CacheHits + m.CacheMisses
	if total == 0 {
		return 0.0
	}
	return float64(m.CacheHits) / float64(total) * 100
}

func (m *Metrics) getTopUsers() map[string]int64 {
	top := make(map[string]int64)
	for user, count := range m.UserRequests {
		if count > 0 {
			top[user] = count
		}
	}
	return top
}

func (m *Metrics) getTopChannels() map[string]int64 {
	top := make(map[string]int64)
	for channel, count := range m.ChannelRequests {
		if count > 0 {
			top[channel] = count
		}
	}
	return top
}
