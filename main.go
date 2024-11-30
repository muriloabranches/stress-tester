package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	url         string
	requests    int
	concurrency int
	method      string
	body        string
	headers     headerFlags
	timeout     time.Duration
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

type headerFlags map[string]string

func (h *headerFlags) String() string {
	return fmt.Sprintf("%v", *h)
}

func (h *headerFlags) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format")
	}
	(*h)[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	return nil
}

func main() {
	headers = make(headerFlags)
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&url, "url", "", "URL of the service to be tested")
	fs.IntVar(&requests, "requests", 0, "Total number of requests")
	fs.IntVar(&concurrency, "concurrency", 1, "Number of concurrent requests")
	fs.StringVar(&method, "method", "GET", "HTTP method to use for requests")
	fs.StringVar(&body, "body", "", "Body of the request")
	fs.Var(&headers, "header", "HTTP headers to include in the request (can be used multiple times)")
	fs.DurationVar(&timeout, "timeout", 30*time.Second, "Timeout for each request")

	fs.Parse(os.Args[1:])

	if url == "" || requests <= 0 || concurrency <= 0 {
		fs.Usage()
		return
	}

	log.Printf("%sStarting stress test with the following parameters:%s\n", Blue, Reset)
	log.Printf("%s  URL: %s%s\n", Blue, url, Reset)
	log.Printf("%s  Total number of requests: %d%s\n", Blue, requests, Reset)
	log.Printf("%s  Concurrency level: %d%s\n", Blue, concurrency, Reset)
	log.Printf("%s  HTTP Method: %s%s\n", Blue, method, Reset)
	log.Printf("%s  Body: %s%s\n", Blue, body, Reset)
	log.Printf("%s  Timeout: %v%s\n", Blue, timeout, Reset)

	if len(headers) > 0 {
		log.Printf("%s  Headers:%s\n", Blue, Reset)
		for key, value := range headers {
			log.Printf("%s    %s: %s%s\n", Blue, key, value, Reset)
		}
	}

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
	client := &http.Client{Timeout: timeout}

	for range requestsChan {
		req, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			log.Printf("%sWorker %d: Failed to create request: %v%s\n", Red, id, err, Reset)
			resultsChan <- 0
			continue
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%sWorker %d: Request failed: %v%s\n", Red, id, err, Reset)
			resultsChan <- 0
			continue
		}
		resultsChan <- resp.StatusCode
		resp.Body.Close()

		current := atomic.AddInt32(completedRequests, 1)
		logColor := Green
		if resp.StatusCode >= 400 {
			logColor = Red
		}
		log.Printf("%sWorker %d: Completed request %d/%d with status: %d%s\n", logColor, id, current, requests, resp.StatusCode, Reset)
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

	headersString := ""
	for key, value := range headers {
		headersString += fmt.Sprintf("    %s: %s\n", key, value)
	}

	reportContent := fmt.Sprintf(`
Stress-Tester Report
================================================
Execution Parameters:
  URL: %s
  HTTP Method: %s
  Body: %s
  Headers:
%s
  Timeout: %v
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
`, url, method, body, headersString, timeout, requests, concurrency, totalTime, totalRequests, status200, successRate, errorRate, averageTimePerRequest)

	for code, count := range statusCodes {
		if code == 0 {
			reportContent += fmt.Sprintf("    Status timeout: %d\n", count)
		} else {
			reportContent += fmt.Sprintf("    Status %d: %d\n", code, count)
		}
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
