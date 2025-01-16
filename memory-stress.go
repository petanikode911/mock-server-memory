package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

// Global variable to hold the allocated memory
var memory []byte

// Memory stress function: Allocate memory dynamically based on size
func stressMemory(memorySize int) {
	// Allocate memory based on the requested size
	memory = make([]byte, memorySize)
	for i := 0; i < len(memory); i++ {
		memory[i] = 1
	}
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

func main() {
	// Set up HTTP server with echo handler
	http.HandleFunc("/echo", echoHandler)

	// Start the server
	port := "8080"
	fmt.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
