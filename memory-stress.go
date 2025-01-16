package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

// Global variable to hold the allocated memory
var memory []byte

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
	currentMemorySize := len(memory)
	increment := 1024 * 1024 * 5 // 5MB increment (adjust as needed)

	// Gradually allocate memory over the given duration
	startTime := time.Now()

	for {
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

// Echo handler: Responds with an echo of the incoming request
func echoHandler(w http.ResponseWriter, r *http.Request) {
	// Get the memory size from query parameters
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

	// Call the memory stress function with the dynamic memory size
	stressMemory(memorySize)

	// Output system memory stats (for monitoring)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Alloc: %v, TotalAlloc: %v, Sys: %v, NumGC: %v\n", memStats.Alloc, memStats.TotalAlloc, memStats.Sys, memStats.NumGC)

	// Echo the request method, URL, and memory stress details
	fmt.Fprintf(w, "Echo: Method = %s, URL = %s\n", r.Method, r.URL)
	fmt.Fprintf(w, "Memory Size Requested: %d bytes\n", memorySize)
}

// Reset handler: Clears the allocated memory and resets to zero
func resetHandler(w http.ResponseWriter, r *http.Request) {
	// Call reset function to clear memory
	memory = nil

	// Output that memory has been reset
	fmt.Fprintf(w, "Memory has been reset to 0 bytes")
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

func main() {
	// Set up HTTP handlers
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/reset", resetHandler) // Reset memory
	http.HandleFunc("/healthz", healthCheckHandler)
	http.HandleFunc("/application/health", livenessProbeHandler) // Liveness probe handler
	http.HandleFunc("/burst", burstHandler)                      // Memory burst handler

	// Start the server
	port := "8888"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
