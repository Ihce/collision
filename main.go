package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func hashFunction(input []byte) []byte {
	hasher := sha256.New()
	hasher.Write(input)
	return hasher.Sum(nil)
}

func bruteForce(targetHash []byte, numWorkers int) string {
	var wg sync.WaitGroup
	resultChan := make(chan string, 1)
	var attempts uint64

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				randomInput := generateRandomString()
				randomHash := hashFunction(randomInput)
				atomic.AddUint64(&attempts, 1)
				if comparePartialHash(randomHash, targetHash) {
					select {
					case resultChan <- string(randomInput):
					default:
					}
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		for {
			fmt.Printf("\rAttempts: %d", atomic.LoadUint64(&attempts))
		}
	}()

	return <-resultChan
}

func generateRandomString() []byte {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func comparePartialHash(hash1, hash2 []byte) bool {
	if len(hash1) < 8 || len(hash2) < 8 {
		return false
	}
	for i := 0; i < 8; i++ {
		if hash1[i] != hash2[i] {
			return false
		}
	}
	return true
}

func main() {
	input := "dbs0025@auburn.edu"
	emailHash := hashFunction([]byte(input))
	fmt.Printf("EmailHash: %x\n", emailHash)

	numWorkers := 4

	start := time.Now()
	matchedString := bruteForce(emailHash, numWorkers)
	elapsed := time.Since(start)

	fmt.Printf("\nMatched String: %s\n", matchedString)
	fmt.Printf("Time taken: %s\n", elapsed)
}
