package main

import (
	"fmt"
	"sync"
	"time"
)

// Global variables to hold the allocated memory and control burst status
var (
	memory       []byte
	burstRunning bool
	burstMutex   sync.Mutex
	maxMemory    int // Memory limit that can be dynamically set via environment variables
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
	increment := 1 * 1024 * 1024 // 1 MiB increment (1048576 bytes)

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
		time.Sleep(100 * time.Millisecond) // Reduced sleep time for faster memory allocation
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

func main() {
	// Set max memory to 128Mi
	maxMemory = 128 * 1024 * 1024 // 128Mi = 134217728 bytes

	// Simulate a burst of memory with a target size of 100 MiB and a hold duration of 2 minutes
	burstMemoryInLoop(100*1024*1024, 5*time.Minute)
}
