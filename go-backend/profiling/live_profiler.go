package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go-backend/internals/api"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type EndpointStats struct {
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	TotalReqs   int       `json:"totalRequests"`
	SuccessReqs int       `json:"successRequests"`
	FailedReqs  int       `json:"failedRequests"`
	MinMs       float64   `json:"minMs"`
	P50Ms       float64   `json:"p50Ms"`
	P90Ms       float64   `json:"p90Ms"`
	P95Ms       float64   `json:"p95Ms"`
	P99Ms       float64   `json:"p99Ms"`
	MaxMs       float64   `json:"maxMs"`
	AvgMs       float64   `json:"avgMs"`
	RPS         float64   `json:"rps"`
	Durations   []float64 `json:"-"`
}

type ConcurrencyBenchmark struct {
	Concurrency int     `json:"concurrency"`
	TotalReqs   int     `json:"totalRequests"`
	DurationSec float64 `json:"durationSec"`
	Throughput  float64 `json:"throughput"`
	AvgLatency  float64 `json:"avgLatencyMs"`
	P95Latency  float64 `json:"p95LatencyMs"`
}

type MemoryProfileStats struct {
	AllocMB        float64 `json:"allocMB"`
	TotalAllocMB   float64 `json:"totalAllocMB"`
	SysMB          float64 `json:"sysMB"`
	HeapAllocMB    float64 `json:"heapAllocMB"`
	HeapInuseMB    float64 `json:"heapInuseMB"`
	HeapSysMB      float64 `json:"heapSysMB"`
	HeapObjects    uint64  `json:"heapObjects"`
	NumGC          uint32  `json:"numGC"`
	PauseTotalNs   uint64  `json:"pauseTotalNs"`
	NumGoroutines  int     `json:"numGoroutines"`
	AllocRateMBPer float64 `json:"allocRateMBPerSec"`
}

type BenchmarkReport struct {
	Timestamp      string                 `json:"timestamp"`
	GoVersion      string                 `json:"goVersion"`
	NumCPU         int                    `json:"numCPU"`
	Endpoints      []EndpointStats        `json:"endpoints"`
	Concurrency    []ConcurrencyBenchmark `json:"concurrency"`
	MemBefore      MemoryProfileStats     `json:"memBefore"`
	MemAfter       MemoryProfileStats     `json:"memAfter"`
	TotalRequests  int                    `json:"totalRequests"`
	OverallRPS     float64                `json:"overallRPS"`
	CPUProfileFile string                 `json:"cpuProfileFile"`
	HeapFile       string                 `json:"heapProfileFile"`
}

func main() {
	fmt.Println("=============================================================")
	fmt.Println("    ITAM Go Backend: Live Runtime Profiling & Benchmarking   ")
	fmt.Println("=============================================================")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 1. Mount net/http/pprof live profiling endpoints
	api.RegisterPprof(r)

	// 2. Add representative endpoints for realistic application profiling
	hasher := utils.NewPasswordHasher()
	samplePassword := "P@ssword123!"
	sampleHash, _ := hasher.HashPassword(samplePassword)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Service ready")
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "uptime": time.Now().Format(time.RFC3339)})
	})

	r.POST("/api/auth/benchmark-login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
		valid := hasher.VerifyPassword(sampleHash, req.Password)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		token, _ := utils.GenerateJWT("admin", "System Administrator", "BenchmarkSecretKey", 60)
		c.JSON(http.StatusOK, gin.H{"token": token, "valid": true})
	})

	r.GET("/api/assets/generate-unit-code", func(c *gin.Context) {
		// Simulate prefix parsing and counter generation
		prefix := "YLVNSDPC" + time.Now().Format("0601")
		unitCode := fmt.Sprintf("%s%03d", prefix, 42)
		c.JSON(http.StatusOK, gin.H{"unitCode": unitCode})
	})

	r.GET("/api/assets/mock-search", func(c *gin.Context) {
		// Simulate in-memory search and JSON serialization
		type MockItem struct {
			UnitCode string `json:"unitCode"`
			Type     string `json:"type"`
			Brand    string `json:"brand"`
			Model    string `json:"model"`
			Status   string `json:"status"`
		}
		items := make([]MockItem, 50)
		for i := 0; i < 50; i++ {
			items[i] = MockItem{
				UnitCode: fmt.Sprintf("YLVNSDPC2601%03d", i+1),
				Type:     "DPC",
				Brand:    "Dell",
				Model:    "OptiPlex 7090",
				Status:   "In-use",
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "total": 50})
	})

	server := httptest.NewServer(r)
	defer server.Close()

	outputDir := "C:/Stuffs/CCNLTHD/go-backend/profiling"
	if envDir := os.Getenv("PROFILING_DIR"); envDir != "" {
		outputDir = envDir
	}
	_ = os.MkdirAll(outputDir, 0755)

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)
	startStats := toMemStats(memStart)

	fmt.Printf("[INFO] Live test server listening at: %s\n", server.URL)
	fmt.Printf("[INFO] Initial Memory: Alloc=%.2fMB, Sys=%.2fMB, Goroutines=%d\n",
		startStats.AllocMB, startStats.SysMB, startStats.NumGoroutines)

	// 3. Start Live CPU Profiling via net/http/pprof endpoint
	cpuProfilePath := filepath.Join(outputDir, "cpu.pprof")
	var cpuBuf bytes.Buffer
	cpuDone := make(chan error, 1)
	go func() {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/debug/pprof/profile?seconds=2", nil)
		r.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			cpuBuf.Write(rec.Body.Bytes())
			_ = os.WriteFile(cpuProfilePath, rec.Body.Bytes(), 0666)
			cpuDone <- nil
		} else {
			cpuDone <- fmt.Errorf("cpu profile returned status %d", rec.Code)
		}
	}()

	// 4. Run Concurrent Benchmarks across Endpoints
	type requestPlan struct {
		name   string
		method string
		path   string
		body   []byte
	}

	plans := []requestPlan{
		{
			name:   "HealthCheck",
			method: "GET",
			path:   "/",
		},
		{
			name:   "API_Status",
			method: "GET",
			path:   "/api/health",
		},
		{
			name:   "GenerateUnitCode",
			method: "GET",
			path:   "/api/assets/generate-unit-code",
		},
		{
			name:   "AssetSearch_JSON",
			method: "GET",
			path:   "/api/assets/mock-search",
		},
		{
			name:   "Auth_Bcrypt_Verify",
			method: "POST",
			path:   "/api/auth/benchmark-login",
			body:   []byte(`{"username":"admin","password":"P@ssword123!"}`),
		},
		{
			name:   "Pprof_Heap_Endpoint",
			method: "GET",
			path:   "/debug/pprof/heap",
		},
	}

	endpointStatsList := make([]EndpointStats, len(plans))
	var totalAllReqs int64
	startTime := time.Now()

	for idx, plan := range plans {
		fmt.Printf("[INFO] Benchmarking Endpoint: %-20s (%s %s)...\n", plan.name, plan.method, plan.path)
		stat := runEndpointBenchmarkWithRouter(r, plan.method, plan.path, plan.body, 15, 120)
		stat.Endpoint = plan.name
		stat.Method = plan.method
		endpointStatsList[idx] = stat
		atomic.AddInt64(&totalAllReqs, int64(stat.TotalReqs))
	}

	// 5. Concurrency Scaling Benchmark
	concurrencyLevels := []int{1, 5, 10, 25, 50}
	var concurrencyResults []ConcurrencyBenchmark
	for _, cLevel := range concurrencyLevels {
		fmt.Printf("[INFO] Benchmarking Concurrency Level: %d workers...\n", cLevel)
		cBench := runConcurrencyBenchmarkWithRouter(r, cLevel, 200)
		concurrencyResults = append(concurrencyResults, cBench)
	}

	// Wait for CPU profile collection
	if err := <-cpuDone; err != nil {
		fmt.Printf("[WARN] CPU profile collection returned: %v\n", err)
	} else {
		fmt.Println("[INFO] Successfully saved live CPU profile ->", cpuProfilePath)
	}

	// 6. Capture live Heap Profile from pprof endpoint
	heapProfilePath := filepath.Join(outputDir, "heap.pprof")
	recHeap := httptest.NewRecorder()
	reqHeap, _ := http.NewRequest("GET", "/debug/pprof/heap", nil)
	r.ServeHTTP(recHeap, reqHeap)
	if recHeap.Code == http.StatusOK {
		_ = os.WriteFile(heapProfilePath, recHeap.Body.Bytes(), 0644)
		fmt.Println("[INFO] Successfully saved live Heap profile ->", heapProfilePath)
	}

	totalDuration := time.Since(startTime).Seconds()
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	endStats := toMemStats(memEnd)
	endStats.AllocRateMBPer = (endStats.TotalAllocMB - startStats.TotalAllocMB) / math.Max(totalDuration, 0.001)

	overallRPS := float64(totalAllReqs) / math.Max(totalDuration, 0.001)

	report := BenchmarkReport{
		Timestamp:      time.Now().Format(time.RFC3339),
		GoVersion:      runtime.Version(),
		NumCPU:         runtime.NumCPU(),
		Endpoints:      endpointStatsList,
		Concurrency:    concurrencyResults,
		MemBefore:      startStats,
		MemAfter:       endStats,
		TotalRequests:  int(totalAllReqs),
		OverallRPS:     overallRPS,
		CPUProfileFile: "cpu.pprof",
		HeapFile:       "heap.pprof",
	}

	resultsJSONPath := filepath.Join(outputDir, "benchmark_results.json")
	reportData, _ := json.MarshalIndent(report, "", "  ")
	_ = os.WriteFile(resultsJSONPath, reportData, 0666)

	fmt.Println("###BENCHMARK_JSON_START###")
	fmt.Println(string(reportData))
	fmt.Println("###BENCHMARK_JSON_END###")

	fmt.Println("###CPU_BASE64_START###")
	fmt.Println(base64.StdEncoding.EncodeToString(cpuBuf.Bytes()))
	fmt.Println("###CPU_BASE64_END###")

	fmt.Println("###HEAP_BASE64_START###")
	fmt.Println(base64.StdEncoding.EncodeToString(recHeap.Body.Bytes()))
	fmt.Println("###HEAP_BASE64_END###")

	fmt.Println("\n=================== LIVE PROFILING SUMMARY ===================")
	fmt.Printf("Total Requests Sent : %d\n", totalAllReqs)
	fmt.Printf("Overall Throughput  : %.2f requests/sec\n", overallRPS)
	fmt.Printf("Memory In-Use (Heap): %.2f MB -> %.2f MB\n", startStats.HeapAllocMB, endStats.HeapAllocMB)
	fmt.Printf("Total Allocations   : %.2f MB (Rate: %.2f MB/sec)\n", endStats.TotalAllocMB, endStats.AllocRateMBPer)
	fmt.Printf("Garbage Collections : %d runs\n", endStats.NumGC-startStats.NumGC)
	fmt.Printf("Active Goroutines   : %d\n", runtime.NumGoroutine())
	fmt.Println("=============================================================")
}

func runEndpointBenchmarkWithRouter(r *gin.Engine, method, path string, body []byte, workers, reqsPerWorker int) EndpointStats {
	var wg sync.WaitGroup
	durationsChan := make(chan float64, workers*reqsPerWorker)
	var successCount, failCount int64

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < reqsPerWorker; i++ {
				reqStart := time.Now()
				var bodyReader io.Reader
				if len(body) > 0 {
					bodyReader = bytes.NewReader(body)
				}
				req, err := http.NewRequest(method, path, bodyReader)
				if err != nil {
					atomic.AddInt64(&failCount, 1)
					continue
				}
				if len(body) > 0 {
					req.Header.Set("Content-Type", "application/json")
				}

				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				durMs := float64(time.Since(reqStart).Microseconds()) / 1000.0

				if rec.Code < 400 {
					atomic.AddInt64(&successCount, 1)
					durationsChan <- durMs
				} else {
					atomic.AddInt64(&failCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	close(durationsChan)
	elapsed := time.Since(start).Seconds()

	var durations []float64
	var sum float64
	for d := range durationsChan {
		durations = append(durations, d)
		sum += d
	}

	sort.Float64s(durations)
	n := len(durations)
	if n == 0 {
		return EndpointStats{TotalReqs: workers * reqsPerWorker}
	}

	p50 := durations[int(float64(n)*0.50)]
	p90 := durations[int(float64(n)*0.90)]
	p95 := durations[int(float64(n)*0.95)]
	p99 := durations[int(float64(n)*0.99)]

	return EndpointStats{
		TotalReqs:   workers * reqsPerWorker,
		SuccessReqs: int(successCount),
		FailedReqs:  int(failCount),
		MinMs:       durations[0],
		P50Ms:       p50,
		P90Ms:       p90,
		P95Ms:       p95,
		P99Ms:       p99,
		MaxMs:       durations[n-1],
		AvgMs:       sum / float64(n),
		RPS:         float64(n) / math.Max(elapsed, 0.001),
	}
}

func runConcurrencyBenchmarkWithRouter(r *gin.Engine, concurrency, reqsPerWorker int) ConcurrencyBenchmark {
	var wg sync.WaitGroup
	var durations []float64
	var mu sync.Mutex

	start := time.Now()
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localDur := make([]float64, 0, reqsPerWorker)
			for i := 0; i < reqsPerWorker; i++ {
				reqStart := time.Now()
				req, _ := http.NewRequest("GET", "/api/health", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				dur := float64(time.Since(reqStart).Microseconds()) / 1000.0
				if rec.Code == http.StatusOK {
					localDur = append(localDur, dur)
				}
			}
			mu.Lock()
			durations = append(durations, localDur...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	sort.Float64s(durations)
	n := len(durations)
	if n == 0 {
		return ConcurrencyBenchmark{Concurrency: concurrency}
	}

	var sum float64
	for _, d := range durations {
		sum += d
	}

	p95 := durations[int(float64(n)*0.95)]
	return ConcurrencyBenchmark{
		Concurrency: concurrency,
		TotalReqs:   n,
		DurationSec: elapsed,
		Throughput:  float64(n) / math.Max(elapsed, 0.001),
		AvgLatency:  sum / float64(n),
		P95Latency:  p95,
	}
}

func toMemStats(m runtime.MemStats) MemoryProfileStats {
	return MemoryProfileStats{
		AllocMB:       float64(m.Alloc) / (1024 * 1024),
		TotalAllocMB:  float64(m.TotalAlloc) / (1024 * 1024),
		SysMB:         float64(m.Sys) / (1024 * 1024),
		HeapAllocMB:   float64(m.HeapAlloc) / (1024 * 1024),
		HeapInuseMB:   float64(m.HeapInuse) / (1024 * 1024),
		HeapSysMB:     float64(m.HeapSys) / (1024 * 1024),
		HeapObjects:   m.HeapObjects,
		NumGC:         m.NumGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGoroutines: runtime.NumGoroutine(),
	}
}
