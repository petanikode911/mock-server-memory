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

// Memory limit (64Mi in bytes)
const memoryLimit = 64 * 1024 * 1024

// Function to allocate or adjust memory size dynamically
func stressMemory(memorySize int) {
	// If memorySize exceeds the limit, set it to memoryLimit
	if memorySize > memoryLimit {
		memorySize = memoryLimit
	}

	// Clear the previous memory allocation if necessary
	memory = nil

	// Allocate new memory based on the requested size
	memory = make([]byte, memorySize)
	for i := 0; i < len(memory); i++ {
		memory[i] = 1
	}
}

// Function to simulate a memory burst to trigger HPA scaling
func burstMemory() {
	// Simulate a burst of memory allocation (e.g., 64Mi at a time, respecting memory limits)
	for i := 0; i < 5; i++ { // Repeat the burst 5 times to increase the load
		stressMemory(memoryLimit)          // Allocate 64Mi at once
		time.Sleep(200 * time.Millisecond) // Small delay to simulate burst
	}
}

// Function to reset memory size
func resetMemory() {
	// Clear the allocated memory
	memory = nil
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

// Burst handler: Triggers memory burst to simulate a high memory load
func burstHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate a burst of memory allocation
	burstMemory()

	// Output that a memory burst has been triggered
	fmt.Fprintf(w, "Memory burst triggered to increase load")
}

// Reset handler: Clears the allocated memory and resets to zero
func resetHandler(w http.ResponseWriter, r *http.Request) {
	// Call reset function to clear memory
	resetMemory()

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

func main() {
	// Set up HTTP handlers
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/burst", burstHandler) // Trigger memory burst
	http.HandleFunc("/reset", resetHandler) // Reset memory
	http.HandleFunc("/healthz", healthCheckHandler)
	http.HandleFunc("/application/health", livenessProbeHandler) // Liveness probe handler

	// Start the server
	port := "8888"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
