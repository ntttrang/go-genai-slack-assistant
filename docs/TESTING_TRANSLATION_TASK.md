# Performance Testing Plan

## Overview

Comprehensive performance testing strategy for the Slack Translation Bot. This plan covers unit benchmarks, load testing, profiling, and detailed testing tasks to ensure high throughput, low latency, and efficient resource utilization.

---

## Performance Goals

- **Throughput**: 500+ concurrent Slack events/second
- **Latency**: P95 < 2s, P99 < 5s
- **Memory**: < 300MB under load
- **Cache Hit Rate**: > 80% for repeated translations
- **Error Rate**: < 0.5%

---

## Task 1: Emoji Handling Performance Benchmarks

**Goal**: Ensure emoji extraction and restoration doesn't become a bottleneck

**Sub-tasks**:
- [ ] Create `tests/benchmarks/emoji_bench_test.go`
- [ ] Benchmark emoji extraction from message (1000+ runs)
- [ ] Benchmark emoji restoration in translation (1000+ runs)
- [ ] Test with messages containing 5, 10, 20+ emojis
- [ ] Track memory allocations
- [ ] Target: < 1ms per operation
- [ ] Document results in `docs/perf-emoji-results.md`

**Expected Output**:
```
BenchmarkEmojiExtraction-8        50000      22156 ns/op       8192 B/op       10 allocs/op
BenchmarkEmojiRestoration-8       50000      18924 ns/op       4096 B/op        5 allocs/op
```

**How to Run**:
```bash
go test -bench=BenchmarkEmoji -benchmem -v ./tests/benchmarks/
```

---

## Task 2: Language Detection Performance Benchmarks

**Goal**: Ensure lingua-go doesn't impact throughput

**Sub-tasks**:
- [ ] Create `tests/benchmarks/language_detection_bench_test.go`
- [ ] Benchmark detection on short messages (< 50 chars)
- [ ] Benchmark detection on medium messages (50-500 chars)
- [ ] Benchmark detection on long messages (500+ chars)
- [ ] Test mixed language messages (EN + VI + emoji)
- [ ] Target: < 5ms per detection
- [ ] Compare cached vs uncached detection
- [ ] Track allocation patterns

**Test Data**:
- Short: "Hello world"
- Medium: "This is a longer English message with multiple words and punctuation"
- Long: Full paragraph (500+ chars)
- Mixed: "Hello 👋 Xin chào 🇻🇳"

**How to Run**:
```bash
go test -bench=BenchmarkLanguageDetection -benchmem -v ./tests/benchmarks/
```

---

## Task 3: Cache Operations Benchmarks

**Goal**: Verify Redis cache performance

**Sub-tasks**:
- [ ] Create `tests/benchmarks/cache_bench_test.go`
- [ ] Benchmark cache GET (hit case)
- [ ] Benchmark cache SET
- [ ] Benchmark cache DELETE
- [ ] Test with various value sizes (1KB, 10KB, 100KB)
- [ ] Test with TTL operations (24h default)
- [ ] Target: < 2ms for cache hits, < 5ms for sets
- [ ] Track network latency impact

**How to Run**:
```bash
go test -bench=BenchmarkCache -benchmem -v ./tests/benchmarks/
```

---

## Task 4: Database Query Benchmarks

**Goal**: Ensure DB queries don't bottleneck under load

**Sub-tasks**:
- [ ] Create `tests/benchmarks/database_bench_test.go`
- [ ] Benchmark indexed SELECT queries
- [ ] Benchmark INSERT operations
- [ ] Benchmark bulk INSERT (100, 1000, 10000 records)
- [ ] Benchmark UPDATE by ID
- [ ] Benchmark connection pool efficiency
- [ ] Target: < 10ms for indexed queries
- [ ] Identify missing indexes

**How to Run**:
```bash
go test -bench=BenchmarkDatabase -benchmem -v ./tests/benchmarks/
```

---

## Task 5: Baseline Load Test (HTTP Endpoint)

**Goal**: Establish baseline performance under concurrent load

**Sub-tasks**:
- [ ] Create `scripts/load-test-baseline.js` (k6)
- [ ] Test 50 concurrent users for 60 seconds
- [ ] Measure: latency (p50, p95, p99), throughput, error rate
- [ ] Verify no memory leaks over duration
- [ ] Generate baseline report
- [ ] Expected results: 500+ req/s, p95 < 2s, < 0.5% errors

**Prerequisites**:
- [ ] k6 installed: `brew install k6`
- [ ] App running on `http://localhost:8080`

**How to Run**:
```bash
# Start app first
./api

# In another terminal
k6 run scripts/load-test-baseline.js
```

---

## Task 6: Spike Load Test

**Goal**: Test system behavior under sudden traffic spikes

**Sub-tasks**:
- [ ] Create `scripts/load-test-spike.js` (k6)
- [ ] Normal load: 50 VUs for 20s
- [ ] Spike: Jump to 200 VUs for 30s
- [ ] Cool down: Return to 50 VUs for 20s
- [ ] Monitor: Recovery time, queue depth, error rate
- [ ] Expected: System recovers within 5s after spike ends

**How to Run**:
```bash
k6 run scripts/load-test-spike.js
```

---

## Task 7: Sustained Load Test (Soak Test)

**Goal**: Identify memory leaks and degradation over time

**Sub-tasks**:
- [ ] Create `scripts/load-test-soak.js` (k6)
- [ ] Steady load: 100 VUs for 5-10 minutes
- [ ] Monitor memory usage every 30s
- [ ] Track: GC pauses, goroutine count, connection count
- [ ] Expected: No memory degradation, stable metrics
- [ ] Generate time-series report

**How to Run**:
```bash
k6 run scripts/load-test-soak.js
```

---

## Task 8: Mixed Event Type Load Test

**Goal**: Test realistic Slack event mix

**Sub-tasks**:
- [ ] Create `scripts/load-test-mixed.js` (k6)
- [ ] 70% message events
- [ ] 20% reaction_added events
- [ ] 10% other events (errors, edge cases)
- [ ] Verify latency breakdown per event type
- [ ] Expected: All types complete within SLA

**How to Run**:
```bash
k6 run scripts/load-test-mixed.js
```

---

## Task 9: Cache Hit Ratio Validation

**Goal**: Ensure cache effectiveness (target: > 80% hit rate)

**Sub-tasks**:
- [ ] Create `tests/performance/cache_ratio_test.go`
- [ ] Replay same 100 messages repeatedly
- [ ] Track cache hits vs misses
- [ ] Measure: Hit ratio %, time saved by cache
- [ ] Calculate: Cache effectiveness in percentage
- [ ] Expected: 80%+ hit rate, 90%+ time saved on hits
- [ ] Test TTL expiration behavior

**How to Run**:
```bash
go test -v ./tests/performance/cache_ratio_test.go
```

---

## Task 10: End-to-End Latency Breakdown

**Goal**: Identify which stage takes the most time

**Sub-tasks**:
- [ ] Create `tests/performance/e2e_latency_test.go`
- [ ] Instrument every stage with timing:
  - Event parsing: ___ms
  - Language detection: ___ms
  - Cache lookup: ___ms
  - Gemini API call: ___ms
  - DB save: ___ms
  - Slack reply: ___ms
- [ ] Run 100 sample messages
- [ ] Generate latency distribution chart
- [ ] Identify optimization opportunities

**How to Run**:
```bash
go test -v ./tests/performance/e2e_latency_test.go
```

---

## Task 11: Connection Pool Tuning

**Goal**: Optimize database connection pool

**Sub-tasks**:
- [ ] Create `tests/performance/connection_pool_test.go`
- [ ] Test with different pool sizes (5, 10, 20, 50)
- [ ] Measure: Query latency, connection wait time
- [ ] Monitor: Active connections, idle connections
- [ ] Find optimal pool size for throughput
- [ ] Expected: No waiting at optimal size under target load

**How to Run**:
```bash
go test -v ./tests/performance/connection_pool_test.go
```

---

## Task 12: Memory Profiling & Analysis

**Goal**: Identify and fix memory leaks

**Sub-tasks**:
- [ ] Run baseline app with `GODEBUG=gctrace=1`
- [ ] Generate heap profile: `go tool pprof http://localhost:6060/debug/pprof/heap`
- [ ] Analyze top memory consumers
- [ ] Run load test and generate new heap profile
- [ ] Compare before/after profiles
- [ ] Identify any growing allocations
- [ ] Fix memory leaks found
- [ ] Document findings in `docs/memory-analysis.md`

**How to Run**:
```bash
# Start app with GC tracing
GODEBUG=gctrace=1 ./api

# In another terminal, capture heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# In pprof interactive shell:
(pprof) top10
(pprof) list main.ProcessEvent
```

---

## Task 13: CPU Profiling & Hot Path Analysis

**Goal**: Identify CPU bottlenecks

**Sub-tasks**:
- [ ] Run load test with CPU profiling enabled
- [ ] Generate CPU profile: `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`
- [ ] Analyze top CPU consumers
- [ ] Create flame graph: `go tool pprof -http=:8081 cpu.prof`
- [ ] Identify optimization opportunities
- [ ] Expected: message processing, translation, cache ops should be top functions
- [ ] Document hottest paths

**How to Run**:
```bash
# Start app (ensure pprof is enabled)
./api

# Run load test in another terminal
k6 run scripts/load-test-baseline.js

# Capture CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# View as flame graph
go tool pprof -http=:8081 cpu.prof
```

---

## Task 14: Goroutine Leak Detection

**Goal**: Ensure no goroutine leaks under load

**Sub-tasks**:
- [ ] Create `tests/performance/goroutine_leak_test.go`
- [ ] Capture baseline goroutine count
- [ ] Run 1000 messages through system
- [ ] Compare final vs baseline goroutine count
- [ ] Expected: No net increase (±5 goroutines acceptable)
- [ ] Use pprof goroutine profile to identify unfinished goroutines

**How to Run**:
```bash
go test -v ./tests/performance/goroutine_leak_test.go
```

---

## Task 15: Error Rate & Failure Mode Testing

**Goal**: Verify system gracefully handles failures

**Sub-tasks**:
- [ ] Create `tests/performance/failure_modes_test.go`
- [ ] Gemini API timeout: Verify fallback behavior
- [ ] Redis connection failure: Verify graceful degradation
- [ ] Database connection failure: Verify circuit breaker
- [ ] Invalid messages: Verify error handling
- [ ] Measure: Error rate, recovery time, data loss
- [ ] Expected: < 0.5% errors, recovery within 10s

**How to Run**:
```bash
go test -v ./tests/performance/failure_modes_test.go
```

---

## Task 16: Rate Limiting Validation

**Goal**: Ensure rate limiting works under load

**Sub-tasks**:
- [ ] Create `tests/performance/rate_limit_test.go`
- [ ] Test per-user rate limit (100 req/s)
- [ ] Test global rate limit (1000 req/s)
- [ ] Verify 429 response when exceeded
- [ ] Measure: Token refill rate, burst capacity
- [ ] Expected: Clean rejection without server errors

**How to Run**:
```bash
go test -v ./tests/performance/rate_limit_test.go
```

---

## Task 17: Database Index Optimization

**Goal**: Ensure optimal query performance

**Sub-tasks**:
- [ ] Analyze current queries with slow log
- [ ] Create `tests/performance/query_optimization_test.go`
- [ ] Test before/after adding missing indexes
- [ ] Expected indexes:
  - message.channel_id (filtering by channel)
  - message.user_id (filtering by user)
  - message.created_at (date range queries)
  - translation.message_id (lookups)
- [ ] Benchmark queries with index vs without
- [ ] Expected: 10x speedup with indexes

**How to Run**:
```bash
go test -v ./tests/performance/query_optimization_test.go
```

---

## Task 18: Stress Test (Breaking Point)

**Goal**: Find system breaking point

**Sub-tasks**:
- [ ] Create `scripts/load-test-stress.js` (k6)
- [ ] Start: 100 VUs, gradually increase
- [ ] Increment: +50 VUs every 30s until system fails
- [ ] Record: Breaking point VU count, error type
- [ ] Monitor: CPU, memory, goroutines at breaking point
- [ ] Expected: Should handle at least 200 concurrent VUs
- [ ] Document findings: When/how system fails

**How to Run**:
```bash
k6 run scripts/load-test-stress.js
```

---

## Task 19: Real-World Scenario Simulation

**Goal**: Test with realistic Slack usage patterns

**Sub-tasks**:
- [ ] Create `scripts/load-test-realistic.js` (k6)
- [ ] Simulate office hours (9am-6pm): High traffic
- [ ] Simulate off-hours: Low traffic
- [ ] Include burst events (announcements, meetings)
- [ ] Message length distribution: 20% short, 60% medium, 20% long
- [ ] Language distribution: 50% EN, 30% VI, 20% mixed
- [ ] Track: Latency, cache effectiveness, cost metrics

**How to Run**:
```bash
k6 run scripts/load-test-realistic.js
```

---

## Task 20: Performance Regression Detection

**Goal**: Catch performance regressions in CI/CD

**Sub-tasks**:
- [ ] Create benchmark baseline: `make perf-baseline`
- [ ] Add to CI pipeline: Compare against baseline
- [ ] Create performance gates:
  - Latency p95: Must be ≤ +10% from baseline
  - Throughput: Must be ≥ -5% from baseline
  - Memory: Must be ≤ +15% from baseline
- [ ] Generate performance reports on each build
- [ ] Expected: CI fails if regression detected

**How to Run**:
```bash
# Create baseline (first time)
make perf-baseline

# Run tests (will compare against baseline)
make perf-test
```

---

## Quick Start Guide

### 1. Setup Prerequisites
```bash
# Install k6
brew install k6

# Install Go benchmarking tools
go install golang.org/x/perf/cmd/benchstat@latest
```

### 2. Create Benchmark Directory
```bash
mkdir -p tests/benchmarks
mkdir -p tests/performance
mkdir -p scripts
```

### 3. Run Basic Benchmarks
```bash
# After creating benchmark files
go test -bench=. -benchmem ./tests/benchmarks/...
```

### 4. Run Load Tests
```bash
# Start your app
./api &

# Run k6 load test
k6 run scripts/load-test-baseline.js

# View results in terminal and summary.json
```

### 5. Profile Application
```bash
# Enable pprof in main app (if not already done)
# Then access profiles
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## Performance Success Criteria

### Phase 1: Baseline (Week 1)
- ✓ Throughput: 500+ req/s
- ✓ P95 Latency: < 2s
- ✓ P99 Latency: < 5s
- ✓ Memory: < 300MB
- ✓ Error rate: < 0.5%

### Phase 2: Optimization (Week 2)
- ✓ Cache hit ratio: > 80%
- ✓ Recovery from spike: < 5s
- ✓ No goroutine leaks
- ✓ No memory leaks

### Phase 3: Advanced (Week 3+)
- ✓ CPU efficiency: < 30% average
- ✓ Connection efficiency: < 20 active
- ✓ Query performance: < 10ms

---

## Metrics Dashboard

| Metric | Target | Status |
|--------|--------|--------|
| Throughput (req/s) | 500+ | TBD |
| P95 Latency (ms) | <2000 | TBD |
| P99 Latency (ms) | <5000 | TBD |
| Memory Usage (MB) | <300 | TBD |
| Error Rate (%) | <0.5 | TBD |
| Cache Hit Ratio (%) | >80 | TBD |
| Emoji op latency (ms) | <1 | TBD |
| Lang detection (ms) | <5 | TBD |
| DB query latency (ms) | <10 | TBD |

---

## Documentation to Create

- [ ] `docs/perf-emoji-results.md` - Emoji benchmark results
- [ ] `docs/perf-language-detection-results.md` - Language detection results
- [ ] `docs/perf-cache-analysis.md` - Cache effectiveness analysis
- [ ] `docs/memory-analysis.md` - Memory profiling analysis
- [ ] `docs/cpu-hotspots.md` - CPU profiling hot spots
- [ ] `docs/performance-baseline.md` - Baseline metrics
- [ ] `docs/performance-regression-gates.md` - Regression detection thresholds

---

## References

- [Go Benchmark Best Practices](https://golang.org/pkg/testing/)
- [pprof Documentation](https://github.com/google/pprof/tree/master/doc)
- [k6 Documentation](https://k6.io/docs/)
- [Redis Performance Tuning](https://redis.io/topics/benchmarks)
