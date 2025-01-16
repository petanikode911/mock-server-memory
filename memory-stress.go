package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Global variables to hold the allocated memory and control burst status
var (
	memory       []byte
	burstRunning bool
	burstMutex   sync.Mutex
)

// Function to allocate or adjust memory size dynamically
func stressMemory(memorySize int) {
	// Clear the previous memory allocation if necessary
	memory = nil

	// Allocate new memory based on the requested size
	memory = make([]byte, memorySize)
	for i := 0; i < len(memory); i++ {
		memory[i] = 1
	}
}

// Function to continuously burst memory over time in a loop
func burstMemoryInLoop(targetMemorySize int, duration time.Duration) {
	burstMutex.Lock()
	burstRunning = true
	burstMutex.Unlock()

	currentMemorySize := len(memory)
	increment := 1024 * 1024 * 5 // 5MB increment (adjust as needed)

	// Gradually allocate memory over the given duration
	startTime := time.Now()

	for {
		// Stop the burst if requested
		burstMutex.Lock()
		if !burstRunning {
			burstMutex.Unlock()
			break
		}
		burstMutex.Unlock()

		if time.Since(startTime) > duration {
			break
		}

		if currentMemorySize < targetMemorySize {
			currentMemorySize += increment
			stressMemory(currentMemorySize)
			fmt.Printf("Current Memory Size: %d bytes\n", currentMemorySize)
		}

		// Optionally, you can add a sleep time between memory allocation increments
		time.Sleep(time.Second)
	}

	// Ensure we reach the target memory size
	stressMemory(targetMemorySize)
}

// Echo handler: Prints the current memory size
func echoHandler(w http.ResponseWriter, r *http.Request) {
	// Output system memory stats (for monitoring)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Alloc: %v, TotalAlloc: %v, Sys: %v, NumGC: %v\n", memStats.Alloc, memStats.TotalAlloc, memStats.Sys, memStats.NumGC)

	// Respond with the current memory size
	fmt.Fprintf(w, "Current Memory Allocated: %d bytes\n", len(memory))
}

// Health check handler: Responds with "OK" if the service is healthy
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Respond with a simple "OK" status to indicate health
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// Liveness check handler: A specific endpoint for liveness probe
func livenessProbeHandler(w http.ResponseWriter, r *http.Request) {
	// Basic liveness check, returning 200 OK
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Service is alive")
}

// Burst handler: Responds to trigger memory burst in a loop
func burstHandler(w http.ResponseWriter, r *http.Request) {
	// Get the target memory size from query parameters
	memorySizeStr := r.URL.Query().Get("memory_size")
	if memorySizeStr == "" {
		http.Error(w, "memory_size query parameter is required", http.StatusBadRequest)
		return
	}

	// Convert the memory size to an integer
	memorySize, err := strconv.Atoi(memorySizeStr)
	if err != nil {
		http.Error(w, "Invalid memory_size value", http.StatusBadRequest)
		return
	}

	// Trigger memory burst in a loop (duration is 10 minutes here)
	duration := 10 * time.Minute
	go burstMemoryInLoop(memorySize, duration)

	// Respond to indicate burst initiation
	fmt.Fprintf(w, "Memory burst started. Target Size: %d bytes over 10 minutes.\n", memorySize)
}

// Stop burst handler: Stops the memory burst loop
func stopBurstHandler(w http.ResponseWriter, r *http.Request) {
	burstMutex.Lock()
	burstRunning = false
	burstMutex.Unlock()

	// Output that the memory burst has been stopped
	fmt.Fprintf(w, "Memory burst has been stopped")
}

func main() {
	// Set up HTTP handlers
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/healthz", healthCheckHandler)
	http.HandleFunc("/application/health", livenessProbeHandler) // Liveness probe handler
	http.HandleFunc("/burst", burstHandler)                      // Memory burst handler
	http.HandleFunc("/stop", stopBurstHandler)                   // Stop memory burst handler

	// Start the server
	port := "8888"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
