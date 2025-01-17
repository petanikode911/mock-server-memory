package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	memory       []byte
	burstRunning bool
	burstMutex   sync.Mutex
	maxMemory    int // Memory limit from environment variable
)

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

func burstMemoryInLoop(targetMemorySize int, holdDuration time.Duration) {
	burstMutex.Lock()
	burstRunning = true
	burstMutex.Unlock()

	currentMemorySize := len(memory)
	increment := 100 * 1024 // 100 KiB increment

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
		time.Sleep(1 * time.Second) // Adjust the sleep duration to allow for gradual increases
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

func livenessProbeHandler(w http.ResponseWriter, r *http.Request) {
	// Basic liveness check, returning 200 OK
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Service is alive")
}

func main() {
	// Read maxMemory from environment variable (55Mi)
	maxMemoryStr := os.Getenv("MAX_MEMORY")
	if maxMemoryStr == "" {
		maxMemoryStr = "57739264" // Default to 55Mi if not set (55 * 1024 * 1024)
	}
	maxMemory, _ = strconv.Atoi(maxMemoryStr)

	// Set up HTTP handlers
	http.HandleFunc("/application/health", livenessProbeHandler) // Liveness probe handler

	// Start memory burst simulation (for demonstration, 50Mi target, hold for 2 minutes)
	go burstMemoryInLoop(50*1024*1024, 2*time.Minute)

	// Start the server
	port := "8888"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
