package main

import (
	"fmt"
	"runtime"
	"sync"
)

func buildSatLocations(satellitesData []Spg4Satellite, times []float64) [][]SatPosition {
	satLocations := make([][]SatPosition, len(satellitesData))

	for i, spg4Sat := range satellitesData {
		satLocations[i] = make([]SatPosition, len(times))

		// errorCount := 0
		for t, julianDate := range times {
			position, err := spg4Sat.propagateAtTime(julianDate)
			if err != nil {
				// fmt.Println("Error running sgp4", err)
				// errorCount++
				continue
			}
			satLocations[i][t] = position
		}

		// if errorCount > 0 {
		// 	fmt.Println("Satellite", i, "had", errorCount, "errors")
		// }
	}

	return satLocations
}

func processTimesWithWorkerPool(numTimes int, numSatellites int, satLocations [][]SatPosition) [][]SatPair {

	numWorkers := runtime.NumCPU() // Number of worker goroutines
	tasks := make(chan int, numTimes)
	results := make([][]SatPair, numTimes) // Preallocate result slice
	var wg sync.WaitGroup

	// Worker function
	worker := func() {
		for i := range tasks {
			// time := times[i]
			timeCluster := NewTimeCluster(i, numSatellites, satLocations)
			results[i] = timeCluster.getAtRiskPairs()
			fmt.Println("At risk pairs", len(results[i]), "for time", i)
		}
		wg.Done()
	}

	// Launch worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker()
	}

	// Add tasks to the channel
	for i := 0; i < numTimes; i++ {
		tasks <- i
	}
	close(tasks) // Close the channel to signal workers no more tasks

	// Wait for all workers to finish
	wg.Wait()
	return results
}
