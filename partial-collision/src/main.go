package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Input []byte
	Hash  []byte
}

func hashFunction(input []byte) []byte {
	hasher := sha256.New()
	hasher.Write(input)
	return hasher.Sum(nil)
}

func bruteForce(ctx context.Context, input []byte, rng *rand.Rand, attempts *uint64, resultChan chan<- [2]Result, wg *sync.WaitGroup) {
	defer wg.Done()

	var firstResult Result
	foundFirst := false

	for {
		select {
		case <-ctx.Done():
			// Stop if context is canceled
			return
		default:
			atomic.AddUint64(attempts, 1)
			randomSuffix := generateRandomString(rng)
			randomInput := append(input, randomSuffix...)
			randomHash := hashFunction(randomInput)
			if !foundFirst {
				if checkFirstFourBytes(randomHash) {
					firstResult = Result{Input: randomInput, Hash: randomHash}
					foundFirst = true
				}
			} else {
				if checkFirstFourBytesMatch(randomHash, firstResult.Hash) {
					resultChan <- [2]Result{
						firstResult,
						{Input: randomInput, Hash: randomHash},
					}
					return

				}
			}
		}
	}
}

func generateRandomString(rng *rand.Rand) []byte {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return b
}

func checkFirstFourBytes(hash []byte) bool {
	return hash[0] == hash[1] && hash[0] == hash[2] && hash[0] == hash[3]
}

func checkFirstFourBytesMatch(hash1, hash2 []byte) bool {
	return hash1[0] == hash2[0] && hash1[1] == hash2[1] && hash1[2] == hash2[2] && hash1[3] == hash2[3]
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	input := []byte("dbs0025@auburn.edu")

	// Setup for parallel brute-force
	numWorkers := 1
	var attempts uint64
	resultChan := make(chan [2]Result)
	var wg sync.WaitGroup

	// Create a cancelable context to stop all goroutines when one finds a match
	ctx, cancel := context.WithCancel(context.Background())

	// Start timer
	start := time.Now()

	// Launch worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerRng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i))) // Each worker gets a unique seed
		go bruteForce(ctx, input, workerRng, &attempts, resultChan, &wg)
	}

	var tickerWg sync.WaitGroup
	tickerWg.Add(1)

	// Goroutine to print attempts per second every 10 seconds
	go func() {
		defer tickerWg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start).Seconds()
				attemptsSoFar := atomic.LoadUint64(&attempts)
				attemptsPerSecond := float64(attemptsSoFar) / elapsed
				fmt.Fprintf(os.Stderr, "\rAttempts per second: %.2f", attemptsPerSecond)
			case <-ctx.Done():
				fmt.Fprintln(os.Stderr, "")
				return
			}
		}
	}()
	// Wait for a result from any goroutine
	matchedResults := <-resultChan
	cancel()  // Signal all workers to stop
	wg.Wait() // Wait for all goroutines to finish

	// End timer
	elapsed := time.Since(start)

	// Output result

	for index, result := range matchedResults {
		fmt.Printf("INPUT %d -- %s\n", index+1, base64.StdEncoding.EncodeToString(result.Input))
		// fmt.Printf("Digest %d -- %x\n", index, result.Hash)
	}
	fmt.Fprintf(os.Stderr, "Attempts: %d\n", atomic.LoadUint64(&attempts))
	fmt.Fprintf(os.Stderr, "Time taken: %s\n", elapsed)

	os.Exit(0)
}
