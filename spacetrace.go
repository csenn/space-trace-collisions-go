package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SatelliteApiData struct {
	ObjectID string `json:"OBJECT_ID"`
	TLE_1    string `json:"TLE_LINE1"`
	TLE_2    string `json:"TLE_LINE2"`
}

// const INTERVALS = 20
const INTERVALS = 360
const TIME_STEP_MINUTES = 4

var START_JULIAN_DATE = createJulianDate(2025, 1, 12, 0, 0, 0)

func main() {

	startTime := time.Now()

	satellitesData, err := loadSatellitesData()
	if err != nil {
		fmt.Println("Error loading satellites data:", err)
		return
	}

	totalSatellites := len(satellitesData)

	spg4Satellites := make([]Spg4Satellite, totalSatellites)
	for i, satApiData := range satellitesData {
		spg4Satellites[i] = NewSgp4Satellite(satApiData.TLE_1, satApiData.TLE_2)
	}

	times := []float64{}
	for i := 0; i < INTERVALS; i++ {
		seconds := float64(i) * 60.0 * TIME_STEP_MINUTES
		times = append(times, julianDateAddSeconds(START_JULIAN_DATE, seconds))
	}

	fmt.Println("Computing satellite locations")
	currentTime := time.Now()
	satLocations := buildSatLocations(spg4Satellites, times)
	fmt.Println("Time to precompute satellite locations:", time.Since(currentTime).Seconds())

	currentTime = time.Now()
	results := tierOneCollisionsWithWorkerPool(len(times), totalSatellites, satLocations)
	fmt.Println(len(results))
	fmt.Println("Time to build clusters:", time.Since(currentTime).Seconds())

	currentTime = time.Now()
	// minDistancePairs := processCollisionsTierTwo(results, times, spg4Satellites)
	minDistancePairs := tierTwoCollisionsWithWorkerPool(results, times, spg4Satellites)
	fmt.Println("Time to process collisions tier two:", time.Since(currentTime).Seconds())

	// print first 100 results
	for _, pair := range minDistancePairs.getTopPairs(100) {
		oneId := satellitesData[pair.Sat1ID].ObjectID
		twoId := satellitesData[pair.Sat2ID].ObjectID
		fmt.Println(oneId, twoId, pair.JulianTime, pair.Distance)
	}

	fmt.Println("Total time:", time.Since(startTime).Seconds())
}

func loadSatellitesData() ([]SatelliteApiData, error) {
	file, err := os.Open("satellites-api.json")
	if err != nil {
		return []SatelliteApiData{}, err
	}
	defer file.Close()

	var satellitesData []SatelliteApiData
	if err := json.NewDecoder(file).Decode(&satellitesData); err != nil {
		return []SatelliteApiData{}, err
	}

	return satellitesData, nil
}
