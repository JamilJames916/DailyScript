package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
}

// Job represents a unit of work
type Job struct {
	ID   int
	Data interface{}
}

// Result represents the result of a job
type Result struct {
	JobID int
	Data  interface{}
	Error error
}

// WorkerFunc is the function signature for work to be done
type WorkerFunc func(job Job) Result

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, numWorkers*2), // Buffer for better performance
		results:    make(chan Result, numWorkers*2),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(workerFunc WorkerFunc) {
	// Start workers
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i, workerFunc)
	}
}

// worker is the actual worker goroutine
func (wp *WorkerPool) worker(id int, workerFunc WorkerFunc) {
	defer wp.wg.Done()
	
	for job := range wp.jobs {
		fmt.Printf("Worker %d processing job %d\n", id, job.ID)
		result := workerFunc(job)
		wp.results <- result
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

// Close closes the job channel and waits for all workers to finish
func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// Example worker functions

// CPUIntensiveWork simulates CPU-intensive work
func CPUIntensiveWork(job Job) Result {
	n := job.Data.(int)
	
	// Simulate some CPU-intensive work (calculate factorial)
	result := 1
	for i := 1; i <= n; i++ {
		result *= i
		// Add some delay to simulate work
		time.Sleep(time.Microsecond * 100)
	}
	
	return Result{
		JobID: job.ID,
		Data:  result,
		Error: nil,
	}
}

// IOIntensiveWork simulates I/O-intensive work
func IOIntensiveWork(job Job) Result {
	filename := job.Data.(string)
	
	// Simulate file I/O
	file, err := os.Create(fmt.Sprintf("temp_%s_%d.txt", filename, job.ID))
	if err != nil {
		return Result{
			JobID: job.ID,
			Data:  nil,
			Error: err,
		}
	}
	defer file.Close()
	defer os.Remove(file.Name()) // Clean up
	
	// Write some data
	data := fmt.Sprintf("This is test data for job %d\n", job.ID)
	_, err = file.WriteString(data)
	
	return Result{
		JobID: job.ID,
		Data:  file.Name(),
		Error: err,
	}
}

// WebRequestWork simulates web requests
func WebRequestWork(job Job) Result {
	url := job.Data.(string)
	
	// Simulate network delay
	time.Sleep(time.Millisecond * time.Duration(100+job.ID*10))
	
	return Result{
		JobID: job.ID,
		Data:  fmt.Sprintf("Response from %s", url),
		Error: nil,
	}
}

// Concurrent map-reduce example
func MapReduce() {
	fmt.Println("\n=== Map-Reduce Example ===")
	
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	// Map phase: square each number
	squared := make(chan int, len(numbers))
	var mapWg sync.WaitGroup
	
	for _, num := range numbers {
		mapWg.Add(1)
		go func(n int) {
			defer mapWg.Done()
			squared <- n * n
		}(num)
	}
	
	// Close channel when all mappers are done
	go func() {
		mapWg.Wait()
		close(squared)
	}()
	
	// Reduce phase: sum all squared numbers
	sum := 0
	for sq := range squared {
		sum += sq
	}
	
	fmt.Printf("Sum of squares (1² + 2² + ... + 10²) = %d\n", sum)
}

// Producer-Consumer example
func ProducerConsumer() {
	fmt.Println("\n=== Producer-Consumer Example ===")
	
	buffer := make(chan int, 5) // Buffered channel
	done := make(chan bool)
	
	// Producer
	go func() {
		for i := 1; i <= 10; i++ {
			fmt.Printf("Producing: %d\n", i)
			buffer <- i
			time.Sleep(time.Millisecond * 100)
		}
		close(buffer)
	}()
	
	// Consumer
	go func() {
		for item := range buffer {
			fmt.Printf("Consuming: %d\n", item)
			time.Sleep(time.Millisecond * 150) // Consumer is slower
		}
		done <- true
	}()
	
	<-done
	fmt.Println("Producer-Consumer example completed")
}

// Fan-out Fan-in example
func FanOutFanIn() {
	fmt.Println("\n=== Fan-out Fan-in Example ===")
	
	input := make(chan int)
	
	// Fan-out: distribute work to multiple workers
	numWorkers := 3
	channels := make([]chan int, numWorkers)
	
	for i := 0; i < numWorkers; i++ {
		channels[i] = make(chan int)
		go func(id int, ch chan int) {
			for n := range ch {
				result := n * n
				fmt.Printf("Worker %d: %d² = %d\n", id, n, result)
				time.Sleep(time.Millisecond * 100)
			}
		}(i, channels[i])
	}
	
	// Distribute input to workers
	go func() {
		for i := 1; i <= 9; i++ {
			channels[(i-1)%numWorkers] <- i
		}
		
		// Close all worker channels
		for _, ch := range channels {
			close(ch)
		}
		close(input)
	}()
	
	time.Sleep(time.Second) // Wait for workers to complete
	fmt.Println("Fan-out Fan-in example completed")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run concurrent-processor.go <example>")
		fmt.Println("Examples:")
		fmt.Println("  cpu        - CPU-intensive work with worker pool")
		fmt.Println("  io         - I/O-intensive work with worker pool")
		fmt.Println("  web        - Simulated web requests with worker pool")
		fmt.Println("  mapreduce  - Map-reduce pattern example")
		fmt.Println("  prodcons   - Producer-consumer pattern example")
		fmt.Println("  fanout     - Fan-out fan-in pattern example")
		fmt.Println("  all        - Run all examples")
		os.Exit(1)
	}

	example := os.Args[1]
	numCPU := runtime.NumCPU()
	
	fmt.Printf("Running on %d CPU cores\n", numCPU)
	runtime.GOMAXPROCS(numCPU)

	switch example {
	case "cpu":
		fmt.Println("\n=== CPU-Intensive Work Example ===")
		
		pool := NewWorkerPool(numCPU)
		pool.Start(CPUIntensiveWork)
		
		// Submit jobs
		numJobs := 10
		go func() {
			for i := 1; i <= numJobs; i++ {
				pool.Submit(Job{ID: i, Data: i + 5}) // Calculate factorial of (i+5)
			}
			pool.Close()
		}()
		
		// Collect results
		start := time.Now()
		for i := 0; i < numJobs; i++ {
			result := <-pool.Results()
			if result.Error != nil {
				fmt.Printf("Job %d failed: %v\n", result.JobID, result.Error)
			} else {
				fmt.Printf("Job %d result: %v\n", result.JobID, result.Data)
			}
		}
		fmt.Printf("Completed in: %v\n", time.Since(start))

	case "io":
		fmt.Println("\n=== I/O-Intensive Work Example ===")
		
		pool := NewWorkerPool(numCPU * 2) // More workers for I/O bound tasks
		pool.Start(IOIntensiveWork)
		
		// Submit jobs
		numJobs := 8
		go func() {
			for i := 1; i <= numJobs; i++ {
				pool.Submit(Job{ID: i, Data: fmt.Sprintf("file_%d", i)})
			}
			pool.Close()
		}()
		
		// Collect results
		start := time.Now()
		for i := 0; i < numJobs; i++ {
			result := <-pool.Results()
			if result.Error != nil {
				fmt.Printf("Job %d failed: %v\n", result.JobID, result.Error)
			} else {
				fmt.Printf("Job %d created: %v\n", result.JobID, result.Data)
			}
		}
		fmt.Printf("Completed in: %v\n", time.Since(start))

	case "web":
		fmt.Println("\n=== Web Request Example ===")
		
		pool := NewWorkerPool(5)
		pool.Start(WebRequestWork)
		
		urls := []string{
			"https://api.github.com",
			"https://httpbin.org/get",
			"https://jsonplaceholder.typicode.com/posts/1",
			"https://api.github.com/users/octocat",
			"https://httpbin.org/delay/1",
		}
		
		// Submit jobs
		go func() {
			for i, url := range urls {
				pool.Submit(Job{ID: i + 1, Data: url})
			}
			pool.Close()
		}()
		
		// Collect results
		start := time.Now()
		for i := 0; i < len(urls); i++ {
			result := <-pool.Results()
			if result.Error != nil {
				fmt.Printf("Job %d failed: %v\n", result.JobID, result.Error)
			} else {
				fmt.Printf("Job %d result: %v\n", result.JobID, result.Data)
			}
		}
		fmt.Printf("Completed in: %v\n", time.Since(start))

	case "mapreduce":
		MapReduce()

	case "prodcons":
		ProducerConsumer()

	case "fanout":
		FanOutFanIn()

	case "all":
		// Run a simpler version of each example
		fmt.Println("Running all concurrency examples...")
		
		MapReduce()
		ProducerConsumer()
		FanOutFanIn()
		
		// Quick worker pool example
		fmt.Println("\n=== Quick Worker Pool Example ===")
		pool := NewWorkerPool(2)
		pool.Start(CPUIntensiveWork)
		
		go func() {
			for i := 1; i <= 3; i++ {
				pool.Submit(Job{ID: i, Data: i + 3})
			}
			pool.Close()
		}()
		
		for i := 0; i < 3; i++ {
			result := <-pool.Results()
			fmt.Printf("Quick job %d result: %v\n", result.JobID, result.Data)
		}

	default:
		fmt.Printf("Unknown example: %s\n", example)
		os.Exit(1)
	}
}
