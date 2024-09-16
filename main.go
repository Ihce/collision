package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func hashFunction(input []byte) []byte {
	hasher := sha256.New()
	hasher.Write(input)
	return hasher.Sum(nil)
}

func bruteForce(ctx context.Context, targetHash []byte, rng *rand.Rand, attempts *uint64, resultChan chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Stop if context is canceled
			return
		default:
			atomic.AddUint64(attempts, 1)
			randomInput := generateRandomString(rng)
			randomHash := hashFunction(randomInput)
			if comparePartialHash(randomHash, targetHash) {
				resultChan <- randomInput
				return
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

func comparePartialHash(hash1, hash2 []byte) bool {
	if len(hash1) < 8 || len(hash2) < 8 {
		return false
	}
	for i := 0; i < 4; i++ {
		if hash1[i] != hash2[i] {
			return false
		}
	}
	return true
}

func main() {

	input := "dbs0025@auburn.edu"
	emailHash := hashFunction([]byte(input))
	fmt.Printf("Email: %s\n", input)
	fmt.Printf("EmailHash: %x\n", emailHash)

	// Setup for parallel brute-force
	numWorkers := 4
	var attempts uint64
	resultChan := make(chan []byte)
	var wg sync.WaitGroup

	// Create a cancelable context to stop all goroutines when one finds a match
	ctx, cancel := context.WithCancel(context.Background())

	// Start timer
	start := time.Now()

	// Launch worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerRng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i))) // Each worker gets a unique seed
		go bruteForce(ctx, emailHash, workerRng, &attempts, resultChan, &wg)
	}

	// Wait for a result from any goroutine
	matchedInput := <-resultChan
	cancel()  // Signal all workers to stop
	wg.Wait() // Wait for all goroutines to finish

	// End timer
	elapsed := time.Since(start)

	// Compute the hash of the matched input
	matchedHash := hashFunction(matchedInput)

	// Output result
	fmt.Printf("\nMatched Input: %s\n", matchedInput)
	fmt.Printf("Matched Hash: %x\n", matchedHash)
	fmt.Printf("Attempts: %d\n", atomic.LoadUint64(&attempts))
	fmt.Printf("Time taken: %s\n", elapsed)

	os.Exit(0)
}
