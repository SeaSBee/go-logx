package stress

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
)

// MemoryProfiler provides memory monitoring capabilities
type MemoryProfiler struct {
	mu              sync.Mutex
	startStats      runtime.MemStats
	endStats        runtime.MemStats
	peakStats       runtime.MemStats
	samples         []MemorySample
	samplingEnabled bool
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// MemorySample represents a memory usage snapshot
type MemorySample struct {
	Timestamp   time.Time
	Alloc       uint64
	TotalAlloc  uint64
	Sys         uint64
	NumGC       uint32
	HeapAlloc   uint64
	HeapSys     uint64
	HeapIdle    uint64
	HeapInuse   uint64
	HeapObjects uint64
}

// MemoryReport contains comprehensive memory analysis
type MemoryReport struct {
	Duration        time.Duration
	StartMemory     MemorySample
	EndMemory       MemorySample
	PeakMemory      MemorySample
	TotalAllocated  uint64
	TotalFreed      uint64
	MemoryGrowth    uint64
	AverageAlloc    uint64
	MaxAlloc        uint64
	GCStats         GCStats
	LeakIndicators  LeakIndicators
	Recommendations []string
}

// GCStats contains garbage collection statistics
type GCStats struct {
	NumGC      uint32
	PauseTotal time.Duration
	PauseNs    []uint64
	LastGC     time.Time
	NextGC     uint64
	GCPercent  int
}

// LeakIndicators contains memory leak detection metrics
type LeakIndicators struct {
	MemoryGrowthRate  float64
	ObjectGrowthRate  float64
	GCFrequency       float64
	HeapFragmentation float64
	PotentialLeak     bool
	LeakConfidence    float64
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		samples:         make([]MemorySample, 0),
		samplingEnabled: false,
		stopChan:        make(chan struct{}),
	}
}

// Start begins memory profiling
func (mp *MemoryProfiler) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	runtime.ReadMemStats(&mp.startStats)
	mp.samples = make([]MemorySample, 0)
	mp.samplingEnabled = true

	// Start background sampling
	mp.wg.Add(1)
	go mp.sampleMemory()
}

// Stop ends memory profiling and returns a report
func (mp *MemoryProfiler) Stop() *MemoryReport {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.samplingEnabled = false
	close(mp.stopChan)
	mp.wg.Wait()

	runtime.ReadMemStats(&mp.endStats)

	// Find peak memory usage
	mp.findPeakMemory()

	return mp.generateReport()
}

// sampleMemory continuously samples memory usage
func (mp *MemoryProfiler) sampleMemory() {
	defer mp.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-mp.stopChan:
			return
		case <-ticker.C:
			mp.takeSample()
		}
	}
}

// takeSample captures a memory snapshot
func (mp *MemoryProfiler) takeSample() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.samplingEnabled {
		return
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	sample := MemorySample{
		Timestamp:   time.Now(),
		Alloc:       stats.Alloc,
		TotalAlloc:  stats.TotalAlloc,
		Sys:         stats.Sys,
		NumGC:       stats.NumGC,
		HeapAlloc:   stats.HeapAlloc,
		HeapSys:     stats.HeapSys,
		HeapIdle:    stats.HeapIdle,
		HeapInuse:   stats.HeapInuse,
		HeapObjects: stats.HeapObjects,
	}

	mp.samples = append(mp.samples, sample)

	// Update peak if this sample has higher memory usage
	if stats.Alloc > mp.peakStats.Alloc {
		mp.peakStats = stats
	}
}

// findPeakMemory identifies the peak memory usage from samples
func (mp *MemoryProfiler) findPeakMemory() {
	if len(mp.samples) == 0 {
		return
	}

	var peakSample MemorySample
	maxAlloc := uint64(0)

	for _, sample := range mp.samples {
		if sample.Alloc > maxAlloc {
			maxAlloc = sample.Alloc
			peakSample = sample
		}
	}

	mp.peakStats = runtime.MemStats{
		Alloc:       peakSample.Alloc,
		TotalAlloc:  peakSample.TotalAlloc,
		Sys:         peakSample.Sys,
		NumGC:       peakSample.NumGC,
		HeapAlloc:   peakSample.HeapAlloc,
		HeapSys:     peakSample.HeapSys,
		HeapIdle:    peakSample.HeapIdle,
		HeapInuse:   peakSample.HeapInuse,
		HeapObjects: peakSample.HeapObjects,
	}
}

// generateReport creates a comprehensive memory analysis
func (mp *MemoryProfiler) generateReport() *MemoryReport {
	if len(mp.samples) == 0 {
		return &MemoryReport{}
	}

	duration := mp.samples[len(mp.samples)-1].Timestamp.Sub(mp.samples[0].Timestamp)

	startSample := mp.samples[0]
	endSample := mp.samples[len(mp.samples)-1]

	// Calculate memory growth
	memoryGrowth := endSample.Alloc - startSample.Alloc
	totalAllocated := mp.endStats.TotalAlloc - mp.startStats.TotalAlloc
	totalFreed := totalAllocated - memoryGrowth

	// Calculate averages
	var totalAlloc uint64
	var maxAlloc uint64
	for _, sample := range mp.samples {
		totalAlloc += sample.Alloc
		if sample.Alloc > maxAlloc {
			maxAlloc = sample.Alloc
		}
	}
	averageAlloc := totalAlloc / uint64(len(mp.samples))

	// Analyze GC behavior
	gcStats := mp.analyzeGC()

	// Detect potential leaks
	leakIndicators := mp.detectLeaks(duration)

	// Generate recommendations
	recommendations := mp.generateRecommendations(memoryGrowth, leakIndicators, gcStats)

	return &MemoryReport{
		Duration:        duration,
		StartMemory:     startSample,
		EndMemory:       endSample,
		PeakMemory:      mp.samples[len(mp.samples)-1], // Will be updated with actual peak
		TotalAllocated:  totalAllocated,
		TotalFreed:      totalFreed,
		MemoryGrowth:    memoryGrowth,
		AverageAlloc:    averageAlloc,
		MaxAlloc:        maxAlloc,
		GCStats:         gcStats,
		LeakIndicators:  leakIndicators,
		Recommendations: recommendations,
	}
}

// analyzeGC analyzes garbage collection behavior
func (mp *MemoryProfiler) analyzeGC() GCStats {
	gcCount := mp.endStats.NumGC - mp.startStats.NumGC

	var totalPause time.Duration
	if len(mp.endStats.PauseNs) > 0 {
		for _, pause := range mp.endStats.PauseNs {
			totalPause += time.Duration(pause)
		}
	}

	// Convert PauseNs array to slice
	pauseNs := make([]uint64, len(mp.endStats.PauseNs))
	copy(pauseNs, mp.endStats.PauseNs[:])

	return GCStats{
		NumGC:      gcCount,
		PauseTotal: totalPause,
		PauseNs:    pauseNs,
		LastGC:     time.Unix(0, int64(mp.endStats.LastGC)),
		NextGC:     mp.endStats.NextGC,
		GCPercent:  100, // Default GC percent
	}
}

// detectLeaks analyzes memory patterns for potential leaks
func (mp *MemoryProfiler) detectLeaks(duration time.Duration) LeakIndicators {
	if len(mp.samples) < 2 {
		return LeakIndicators{}
	}

	startSample := mp.samples[0]
	endSample := mp.samples[len(mp.samples)-1]

	// Calculate growth rates
	memoryGrowthRate := 0.0
	objectGrowthRate := 0.0
	if duration.Seconds() > 0 {
		memoryGrowthRate = float64(endSample.Alloc-startSample.Alloc) / duration.Seconds()
		objectGrowthRate = float64(endSample.HeapObjects-startSample.HeapObjects) / duration.Seconds()
	}

	// Calculate GC frequency
	gcFrequency := 0.0
	if duration.Seconds() > 0 {
		gcFrequency = float64(mp.endStats.NumGC-mp.startStats.NumGC) / duration.Seconds()
	}

	// Calculate heap fragmentation
	heapFragmentation := 0.0
	if endSample.HeapSys > 0 {
		heapFragmentation = float64(endSample.HeapIdle) / float64(endSample.HeapSys) * 100
	}

	// Determine potential leak
	potentialLeak := memoryGrowthRate > 1024*1024 // 1MB/s growth threshold
	leakConfidence := 0.0

	if potentialLeak {
		leakConfidence = 0.8
		if objectGrowthRate > 1000 { // 1000 objects/s growth
			leakConfidence = 0.9
		}
		if gcFrequency < 0.1 { // Low GC frequency
			leakConfidence = 0.95
		}
	}

	return LeakIndicators{
		MemoryGrowthRate:  memoryGrowthRate,
		ObjectGrowthRate:  objectGrowthRate,
		GCFrequency:       gcFrequency,
		HeapFragmentation: heapFragmentation,
		PotentialLeak:     potentialLeak,
		LeakConfidence:    leakConfidence,
	}
}

// generateRecommendations creates memory optimization suggestions
func (mp *MemoryProfiler) generateRecommendations(memoryGrowth uint64, leaks LeakIndicators, gc GCStats) []string {
	var recommendations []string

	if memoryGrowth > 100*1024*1024 { // 100MB growth
		recommendations = append(recommendations, "High memory growth detected - consider object pooling")
	}

	if leaks.PotentialLeak {
		recommendations = append(recommendations, "Potential memory leak detected - review object lifecycle")
	}

	if gc.NumGC < 5 {
		recommendations = append(recommendations, "Low GC frequency - consider manual GC calls")
	}

	if leaks.HeapFragmentation > 50 {
		recommendations = append(recommendations, "High heap fragmentation - consider memory compaction")
	}

	if gc.PauseTotal > 100*time.Millisecond {
		recommendations = append(recommendations, "Long GC pauses - consider tuning GC parameters")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Memory usage is healthy")
	}

	return recommendations
}

// PrintReport prints a formatted memory report
func (mp *MemoryProfiler) PrintReport(report *MemoryReport) {
	fmt.Printf("\n=== MEMORY PROFILING REPORT ===\n")
	fmt.Printf("Duration: %v\n", report.Duration)
	fmt.Printf("Memory Growth: %s\n", formatBytes(report.MemoryGrowth))
	fmt.Printf("Total Allocated: %s\n", formatBytes(report.TotalAllocated))
	fmt.Printf("Total Freed: %s\n", formatBytes(report.TotalFreed))
	fmt.Printf("Average Allocation: %s\n", formatBytes(report.AverageAlloc))
	fmt.Printf("Peak Allocation: %s\n", formatBytes(report.MaxAlloc))

	fmt.Printf("\n--- Garbage Collection ---\n")
	fmt.Printf("GC Count: %d\n", report.GCStats.NumGC)
	fmt.Printf("Total GC Pause: %v\n", report.GCStats.PauseTotal)
	fmt.Printf("GC Frequency: %.2f GC/s\n", report.LeakIndicators.GCFrequency)

	fmt.Printf("\n--- Leak Analysis ---\n")
	fmt.Printf("Memory Growth Rate: %.2f MB/s\n", report.LeakIndicators.MemoryGrowthRate/1024/1024)
	fmt.Printf("Object Growth Rate: %.2f objects/s\n", report.LeakIndicators.ObjectGrowthRate)
	fmt.Printf("Heap Fragmentation: %.1f%%\n", report.LeakIndicators.HeapFragmentation)
	fmt.Printf("Potential Leak: %t (confidence: %.1f%%)\n",
		report.LeakIndicators.PotentialLeak, report.LeakIndicators.LeakConfidence*100)

	fmt.Printf("\n--- Recommendations ---\n")
	for i, rec := range report.Recommendations {
		fmt.Printf("%d. %s\n", i+1, rec)
	}
	fmt.Printf("\n")
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// StartCPUProfile starts CPU profiling
func StartCPUProfile(filename string) (*os.File, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	pprof.StartCPUProfile(f)
	return f, nil
}

// StopCPUProfile stops CPU profiling
func StopCPUProfile(f *os.File) {
	pprof.StopCPUProfile()
	f.Close()
}

// StartMemoryProfile starts memory profiling
func StartMemoryProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.WriteHeapProfile(f)
}

// StartTrace starts execution tracing
func StartTrace(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	return trace.Start(f)
}

// StopTrace stops execution tracing
func StopTrace() {
	trace.Stop()
}
