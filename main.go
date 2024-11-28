package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	url         string
	requests    int
	concurrency int
	method      string
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

func init() {
	flag.StringVar(&url, "url", "", "URL of the service to be tested")
	flag.IntVar(&requests, "requests", 0, "Total number of requests")
	flag.IntVar(&concurrency, "concurrency", 1, "Number of concurrent requests")
	flag.StringVar(&method, "method", "GET", "HTTP method to use for requests")
}

func main() {
	flag.Parse()

	if url == "" || requests <= 0 || concurrency <= 0 {
		flag.Usage()
		return
	}

	log.Printf("%sStarting stress test with the following parameters:%s\n", Blue, Reset)
	log.Printf("%s  URL: %s%s\n", Blue, url, Reset)
	log.Printf("%s  Total number of requests: %d%s\n", Blue, requests, Reset)
	log.Printf("%s  Concurrency level: %d%s\n", Blue, concurrency, Reset)
	log.Printf("%s  HTTP Method: %s%s\n", Blue, method, Reset)

	var wg sync.WaitGroup
	requestsChan := make(chan struct{}, requests)
	resultsChan := make(chan int, requests)

	var completedRequests int32

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i+1, &wg, requestsChan, resultsChan, &completedRequests)
	}

	for i := 0; i < requests; i++ {
		requestsChan <- struct{}{}
	}
	close(requestsChan)

	wg.Wait()
	close(resultsChan)

	totalTime := time.Since(startTime)
	report(resultsChan, totalTime)
}

func worker(id int, wg *sync.WaitGroup, requestsChan <-chan struct{}, resultsChan chan<- int, completedRequests *int32) {
	defer wg.Done()
	for range requestsChan {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			log.Printf("%sWorker %d: Failed to create request: %v%s\n", Red, id, err, Reset)
			resultsChan <- 0
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("%sWorker %d: Request failed: %v%s\n", Red, id, err, Reset)
			resultsChan <- 0
			continue
		}
		resultsChan <- resp.StatusCode
		resp.Body.Close()

		current := atomic.AddInt32(completedRequests, 1)
		log.Printf("%sWorker %d: Completed request %d/%d with status: %d%s\n", Green, id, current, requests, resp.StatusCode, Reset)
	}
}

func report(resultsChan <-chan int, totalTime time.Duration) {
	totalRequests := 0
	status200 := 0
	statusCodes := make(map[int]int)

	for status := range resultsChan {
		totalRequests++
		if status == 200 {
			status200++
		} else {
			statusCodes[status]++
		}
	}

	successRate := float64(status200) / float64(totalRequests) * 100
	errorRate := float64(totalRequests-status200) / float64(totalRequests) * 100
	averageTimePerRequest := totalTime / time.Duration(totalRequests)

	reportContent := fmt.Sprintf(`
Stress Test Report
================================================
Execution Parameters:
  URL: %s
  HTTP Method: %s
  Total number of requests desired: %d
  Concurrency level: %d
------------------------------------------------
Results:
  Total time taken: %v
  Total number of requests made: %d
  Number of requests with HTTP 200 status: %d
  Success rate: %.2f%%
  Error rate: %.2f%%
  Average time per request: %v

  Distribution of other HTTP status codes:
`, url, method, requests, concurrency, totalTime, totalRequests, status200, successRate, errorRate, averageTimePerRequest)

	for code, count := range statusCodes {
		reportContent += fmt.Sprintf("    Status %d: %d\n", code, count)
	}
	reportContent += "================================================\n"

	// Print report to console
	fmt.Println(reportContent)

	// Write report to file
	file, err := os.Create("stress_test_report.txt")
	if err != nil {
		log.Fatalf("Failed to create report file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(reportContent)
	if err != nil {
		log.Fatalf("Failed to write report to file: %v", err)
	}
}
