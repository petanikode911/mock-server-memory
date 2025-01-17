package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Global variables to hold the allocated memory and control burst status
var (
	memory       []byte
	burstRunning bool
	burstMutex   sync.Mutex
	maxMemory    int // Memory limit that can be dynamically set via HTTP
)

// Function to allocate or adjust memory size dynamically with a cap to maxMemory
func stressMemory(memorySize int) {
	// Cap memory to the limit
	if memorySize > maxMemory {
		fmt.Println("Requested memory exceeds the limit, capping to max memory limit")
		memorySize = maxMemory
	}

	// Clear the previous memory allocation if necessary
	memory = nil

	// Allocate new memory based on the requested size
	memory = make([]byte, memorySize)
	for i := 0; i < len(memory); i++ {
		memory[i] = 1
	}
}

// Function to gradually increase memory over time and then hold it for the specified duration
func burstMemoryInLoop(targetMemorySize int, holdDuration time.Duration) {
	burstMutex.Lock()
	burstRunning = true
	burstMutex.Unlock()

	currentMemorySize := len(memory)
	increment := 34 * 1024 // 34 KiB increment (approximately 34816 bytes)

	startTime := time.Now()

	// Gradually increase memory
	for {
		// Stop the burst if requested
		burstMutex.Lock()
		if !burstRunning {
			burstMutex.Unlock()
			break
		}
		burstMutex.Unlock()

		if currentMemorySize < targetMemorySize {
			currentMemorySize += increment
			if currentMemorySize > maxMemory {
				currentMemorySize = maxMemory // Cap to the limit
			}
			stressMemory(currentMemorySize)
			fmt.Printf("Current Memory Size: %d bytes\n", currentMemorySize)
		}

		// Sleep to avoid instant overload, allowing the system to react
		time.Sleep(200 * time.Millisecond) // Reduced sleep time for faster memory allocation
	}

	// Hold memory at target size for the specified duration
	if time.Since(startTime) < holdDuration {
		remainingTime := holdDuration - time.Since(startTime)
		fmt.Printf("Holding memory at %d bytes for %v\n", targetMemorySize, remainingTime)
		time.Sleep(remainingTime)
	}

	// Stop the burst after the target time has passed
	burstMutex.Lock()
	burstRunning = false
	burstMutex.Unlock()

	// Ensure the target memory size is reached, but don't exceed the limit
	stressMemory(targetMemorySize)
}

// Reset handler: Clears the allocated memory and resets to zero
func resetHandler(w http.ResponseWriter, r *http.Request) {
	burstMutex.Lock()
	burstRunning = false
	burstMutex.Unlock()

	// Clear the allocated memory
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

	// Get the hold duration from query parameters (optional, default is 30 minutes)
	holdDurationStr := r.URL.Query().Get("hold_duration")
	holdDuration := 30 * time.Minute // Default to 30 minutes
	if holdDurationStr != "" {
		duration, err := strconv.Atoi(holdDurationStr)
		if err != nil {
			http.Error(w, "Invalid hold_duration value", http.StatusBadRequest)
			return
		}
		holdDuration = time.Duration(duration) * time.Minute
	}

	// Set maxMemory to the memory_size value from the query
	maxMemory = memorySize

	// Trigger memory burst in a loop asynchronously (duration is the time to hold the memory)
	go burstMemoryInLoop(memorySize, holdDuration)

	// Immediately respond to indicate burst initiation without waiting for hold_duration to finish
	fmt.Fprintf(w, "Memory burst started. Target Size: %d bytes for %v.\n", memorySize, holdDuration)
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
	http.HandleFunc("/reset", resetHandler) // Reset memory
	http.HandleFunc("/healthz", healthCheckHandler)
	http.HandleFunc("/application/health", livenessProbeHandler) // Liveness probe handler
	http.HandleFunc("/burst", burstHandler)                      // Memory burst handler
	http.HandleFunc("/stop", stopBurstHandler)                   // Stop memory burst handler

	// Start the server
	port := "8888"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
