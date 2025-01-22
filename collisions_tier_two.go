package main

import (
	"math"
	"runtime"
	"sort"
	"sync"
)

func tierTwoCollisionsWithWorkerPool(atRiskPairs [][]SatPair, julianTimes []float64, spg4Satellites []Spg4Satellite) *MinDistancePairs {

	// Number of worker goroutines
	numWorkers := runtime.NumCPU()

	numTasks := 0
	for _, pairs := range atRiskPairs {
		numTasks += len(pairs)
	}

	tasks := make(chan int, numTasks)
	var wg sync.WaitGroup

	minDistancePairs := NewMinDistancePairs()

	// Worker function
	worker := func() {
		for i := range tasks {
			julianTime := julianTimes[i]

			timeLeft := julianDateAddSeconds(julianTime, -10*60)
			timeRight := julianDateAddSeconds(julianTime, 10*60)

			for _, pair := range atRiskPairs[i] {
				satOne := spg4Satellites[pair.ID1]
				satTwo := spg4Satellites[pair.ID2]

				minTime, err := binarySearch(satOne, satTwo, timeLeft, timeRight)
				if err != nil {
					// fmt.Println("error propagating sat1", err)
					continue
				}

				minDistance, _ := distanceBetweenSatellites(satOne, satTwo, minTime)
				// fmt.Println("pair", pair.ID1, pair.ID2, "minTime", minTime, "minDistance", minDistance)

				if minDistance == 0 {
					continue
				}

				minDistancePairs.addPair(pair.ID1, pair.ID2, minTime, minDistance)
			}

		}
		wg.Done()
	}

	// Launch worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker()
	}

	for i := range atRiskPairs {
		tasks <- i
	}

	close(tasks) // Close the channel to signal workers no more tasks

	// Wait for all workers to finish
	wg.Wait()

	return minDistancePairs
}

func processCollisionsTierTwo(atRiskPairs [][]SatPair, julianTimes []float64, spg4Satellites []Spg4Satellite) *MinDistancePairs {

	minDistancePairs := NewMinDistancePairs()

	for i, pairs := range atRiskPairs {

		julianTime := julianTimes[i]
		timeLeft := julianDateAddSeconds(julianTime, -10*60)
		timeRight := julianDateAddSeconds(julianTime, 10*60)

		for _, pair := range pairs {

			minTime, err := binarySearch(spg4Satellites[pair.ID1], spg4Satellites[pair.ID2], timeLeft, timeRight)
			if err != nil {
				// fmt.Println("error propagating sat1", err)
				continue
			}

			minDistance, _ := distanceBetweenSatellites(spg4Satellites[pair.ID1], spg4Satellites[pair.ID2], minTime)
			// fmt.Println("pair", pair.ID1, pair.ID2, "minTime", minTime, "minDistance", minDistance)

			if minDistance == 0 {
				continue
			}

			minDistancePairs.addPair(pair.ID1, pair.ID2, minTime, minDistance)

		}
	}

	return minDistancePairs
}

func binarySearch(sat1, sat2 Spg4Satellite, timeLeft, timeRight float64) (atTime float64, err error) {

	timeMid := (timeLeft + timeRight) / 2.0

	// TODO: this should be seconds not minutes
	if math.Abs(differenceInSeconds(timeLeft, timeRight)) < 0.1 {
		return timeMid, nil
	}

	deltaTime := julianDateAddSeconds(timeMid, -0.05)
	distanceLeftDelta, err := distanceBetweenSatellites(sat1, sat2, deltaTime)
	if err != nil {
		return 0, err
	}

	distanceMid, err := distanceBetweenSatellites(sat1, sat2, timeMid)
	if err != nil {
		return 0, err
	}

	// fmt.Println("distanceLeftDelta", timeMid, distanceMid)

	if distanceLeftDelta < distanceMid {
		return binarySearch(sat1, sat2, timeLeft, timeMid)
	} else {
		return binarySearch(sat1, sat2, timeMid, timeRight)
	}

}

func distanceBetweenSatellites(sat1, sat2 Spg4Satellite, atTime float64) (float64, error) {

	sat1Pos, err := sat1.propagateAtTime(atTime)
	if err != nil {
		return 0, err
	}

	sat2Pos, err := sat2.propagateAtTime(atTime)
	if err != nil {
		return 0, err
	}

	return distanceBetweenPositions(sat1Pos, sat2Pos), nil
}

func distanceBetweenPositions(sat1Pos, sat2Pos SatPosition) float64 {
	return math.Sqrt(math.Pow(sat1Pos.X-sat2Pos.X, 2) + math.Pow(sat1Pos.Y-sat2Pos.Y, 2) + math.Pow(sat1Pos.Z-sat2Pos.Z, 2))
}

type MinDistancePairs struct {
	pairs sync.Map
	// map[SatPair]MinDistancePoint
}

type MinDistancePoint struct {
	JulianTime float64
	Distance   float64
}

func NewMinDistancePairs() *MinDistancePairs {
	return &MinDistancePairs{
		pairs: sync.Map{},
	}
}

func (p *MinDistancePairs) addPair(sat1, sat2 int, julianTime float64, distance float64) {
	pair := NewSatPair(sat1, sat2)
	point := MinDistancePoint{
		JulianTime: julianTime,
		Distance:   distance,
	}

	if val, ok := p.pairs.Load(pair); !ok {
		p.pairs.Store(pair, point)
	} else if existingPoint, ok := val.(MinDistancePoint); ok && distance < existingPoint.Distance {
		p.pairs.Store(pair, point)
	}
}

type OutPair struct {
	Sat1ID     int
	Sat2ID     int
	JulianTime float64
	Distance   float64
}

func (p *MinDistancePairs) getTopPairs(n int) []OutPair {

	results := []OutPair{}

	p.pairs.Range(func(key, value interface{}) bool {
		pair := key.(SatPair)
		point := value.(MinDistancePoint)
		results = append(results, OutPair{
			Sat1ID:     pair.ID1,
			Sat2ID:     pair.ID2,
			JulianTime: point.JulianTime,
			Distance:   point.Distance,
		})
		return true
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	return results[:n]

}
