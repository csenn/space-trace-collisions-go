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

const INTERVALS = 360

// const INTERVALS = 20
const TIME_STEP_MINUTES = 4

var START_JULIAN_DATE = createJulianDate(2025, 1, 12, 0, 0, 0)

func main() {
	// fmt.Println("jul date", satellite.JDay(2025, 1, 12, 0, 0, 0))
	// fmt.Println("jul date back", satellite.JDay(2025, 1, 12, -1000, 0, 0))
	// fmt.Println("jul date forw", satellite.JDay(2025, 1, 12, 1000, 0, 0))

	// line1 := "1 84232U 79104    25011.29418726 +.00010894 +00000+0 +10327-1 0  9993"
	// line2 := "2 84232  20.2440 103.5465 6615434  81.8936 342.8433  3.09154996 94164"

	// satTemp := satellite.ParseTLE(line1, line2, satellite.GravityWGS84)
	// posTemp, _ := satellite.Propagate(satTemp, 2025, 1, 12, 0, 0, 0)

	// fmt.Println("woops", posTemp)
	// fmt.Println("jul date", satellite.JDay(2025, 1, 12, 0, 0, 0))

	// 2.4606875e+06
	// 2460687.5
	// woops {826.777049060983 -37.07150745757041 -275.86613536280976}
	// DIMS:  12705.022785775955 -13783.223025773186 -3409.8462618576045
	//       [12705.022786314943 -13783.223025751715 -3409.846262052983]
	// FINAL-Res [-1051.9096188795945 -9649.986178906509 1217.9123088455192]
	// -2393.5129261     9965.0899195       -0.0003231

	// Read the file content

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
	results := processTimesWithWorkerPool(len(times), totalSatellites, satLocations)
	fmt.Println(len(results))
	fmt.Println("Time to build clusters:", time.Since(currentTime).Seconds())

	currentTime = time.Now()
	// minDistancePairs := processCollisionsTierTwo(results, times, spg4Satellites)
	minDistancePairs := processCollisionTwoWithWorkerPool(results, times, spg4Satellites)
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
