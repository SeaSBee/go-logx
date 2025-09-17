package stress

import (
	"runtime"
	"sync"
	"testing"
	"time"

	logx "github.com/seasbee/go-logx"
)

// TestMemoryProfiling runs a focused memory profiling test
func TestMemoryProfiling(t *testing.T) {
	// Initialize memory profiler
	profiler := NewMemoryProfiler()
	profiler.Start()

	// Create logger with minimal I/O
	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	config.Development = false // Reduce I/O overhead

	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test parameters
	numGoroutines := 1000
	messagesPerGoroutine := 20
	totalMessages := numGoroutines * messagesPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start memory-intensive logging
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				// Create structured logs with various field types
				logger.Info("Memory profiling test",
					logx.Int("worker_id", id),
					logx.Int("message_id", j),
					logx.String("test_type", "memory_profiling"),
					logx.Int64("timestamp", time.Now().UnixNano()),
					logx.Float64("random_value", float64(id*j)/1000.0),
					logx.Bool("success", (id+j)%2 == 0),
				)
			}
		}(i)
	}

	wg.Wait()

	// Stop profiling and generate report
	report := profiler.Stop()
	profiler.PrintReport(report)

	// Validate results
	if report.MemoryGrowth > 50*1024*1024 { // 50MB threshold
		t.Logf("⚠️ High memory growth: %s", formatBytes(report.MemoryGrowth))
	}

	if report.LeakIndicators.PotentialLeak {
		t.Logf("⚠️ Potential memory leak detected (confidence: %.1f%%)",
			report.LeakIndicators.LeakConfidence*100)
	}

	t.Logf("✅ Memory profiling test completed: %d messages, %s memory growth",
		totalMessages, formatBytes(report.MemoryGrowth))
}

// TestMemoryLeakDetection tests specific memory leak scenarios
func TestMemoryLeakDetection(t *testing.T) {
	profiler := NewMemoryProfiler()
	profiler.Start()

	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	config.Development = false

	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Simulate potential memory leak scenario
	var wg sync.WaitGroup
	numIterations := 10
	goroutinesPerIteration := 100

	for iteration := 0; iteration < numIterations; iteration++ {
		wg.Add(goroutinesPerIteration)

		for i := 0; i < goroutinesPerIteration; i++ {
			go func(id, iter int) {
				defer wg.Done()

				// Create large structured logs
				for j := 0; j < 5; j++ {
					logger.Info("Leak detection test",
						logx.Int("iteration", iter),
						logx.Int("worker_id", id),
						logx.Int("message_id", j),
						logx.String("large_field", generateLargeString(500)),
						logx.Int64("timestamp", time.Now().UnixNano()),
					)
				}
			}(i, iteration)
		}

		wg.Wait()

		// Force GC between iterations to detect leaks
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	report := profiler.Stop()
	profiler.PrintReport(report)

	// Analyze leak indicators
	if report.LeakIndicators.PotentialLeak {
		t.Logf("🔍 Memory leak analysis:")
		t.Logf("  - Memory growth rate: %.2f MB/s", report.LeakIndicators.MemoryGrowthRate/1024/1024)
		t.Logf("  - Object growth rate: %.2f objects/s", report.LeakIndicators.ObjectGrowthRate)
		t.Logf("  - GC frequency: %.2f GC/s", report.LeakIndicators.GCFrequency)
		t.Logf("  - Heap fragmentation: %.1f%%", report.LeakIndicators.HeapFragmentation)
		t.Logf("  - Leak confidence: %.1f%%", report.LeakIndicators.LeakConfidence*100)
	}

	t.Logf("✅ Memory leak detection test completed")
}

// TestGarbageCollectionBehavior tests GC behavior under load
func TestGarbageCollectionBehavior(t *testing.T) {
	profiler := NewMemoryProfiler()
	profiler.Start()

	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	config.Development = false

	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create memory pressure to trigger GC
	var wg sync.WaitGroup
	numWorkers := 500
	messagesPerWorker := 30

	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < messagesPerWorker; j++ {
				// Create logs with varying memory usage
				logger.Info("GC behavior test",
					logx.Int("worker_id", id),
					logx.Int("message_id", j),
					logx.String("payload", generateLargeString(100+id%200)),
					logx.Int64("timestamp", time.Now().UnixNano()),
				)
			}
		}(i)
	}

	wg.Wait()

	report := profiler.Stop()
	profiler.PrintReport(report)

	// Analyze GC behavior
	t.Logf("🔍 Garbage Collection Analysis:")
	t.Logf("  - GC Count: %d", report.GCStats.NumGC)
	t.Logf("  - Total GC Pause: %v", report.GCStats.PauseTotal)

	// Handle case where no GC occurred
	if report.GCStats.NumGC > 0 {
		t.Logf("  - GC Frequency: %.2f GC/s", report.LeakIndicators.GCFrequency)
		avgPause := report.GCStats.PauseTotal / time.Duration(report.GCStats.NumGC)
		t.Logf("  - Average GC Pause: %v", avgPause)

		if avgPause > 10*time.Millisecond {
			t.Logf("⚠️ High average GC pause: %v", avgPause)
		}
	} else {
		t.Logf("  - GC Frequency: 0.00 GC/s (no GC occurred)")
		t.Logf("  - Average GC Pause: N/A (no GC occurred)")
	}

	t.Logf("✅ GC behavior test completed")
}
