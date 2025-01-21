package main

import (
	"sort"
)

const BOX_SIZE = 1200
const MAX_DIST = 100

type ClusterKey struct {
	X, Y, Z int
}

type SatCoord struct {
	SatelliteID int
	Dim         float64
}

type TimeCluster struct {
	TimeIndex    int
	SatCount     int
	SatLocations [][]SatPosition
	Clusters     map[ClusterKey][]int
}

func NewTimeCluster(timeIndex, satCount int, satLocations [][]SatPosition) *TimeCluster {
	return &TimeCluster{
		TimeIndex:    timeIndex,
		SatCount:     satCount,
		SatLocations: satLocations,
		Clusters:     make(map[ClusterKey][]int),
	}
}

func (t *TimeCluster) buildClusters() {

	for i := 0; i < t.SatCount; i++ {
		position := t.SatLocations[i][t.TimeIndex]
		clusterKey := createClusterKey(position)

		if _, ok := t.Clusters[clusterKey]; !ok {
			t.Clusters[clusterKey] = []int{i}
		} else {
			t.Clusters[clusterKey] = append(t.Clusters[clusterKey], i)
		}
	}
}

func (t *TimeCluster) getAtRiskPairs() []SatPair {

	if len(t.Clusters) == 0 {
		t.buildClusters()
	}

	atRiskPairSet := make(map[SatPair]struct{})

	for clusterKey := range t.Clusters {
		allSatIndexes := t.getAllSatIdsInNeighborCluster(clusterKey)
		closePairs := t.getClosePairs(allSatIndexes)
		for pair := range closePairs {
			atRiskPairSet[pair] = struct{}{}
		}
	}

	atRiskPairs := []SatPair{}
	for pair := range atRiskPairSet {
		atRiskPairs = append(atRiskPairs, pair)
	}

	return atRiskPairs
}

func (t *TimeCluster) getClosePairs(satIndexes []int) map[SatPair]struct{} {

	xCoords := []SatCoord{}
	yCoords := []SatCoord{}
	zCoords := []SatCoord{}

	for _, satIndex := range satIndexes {
		satPosition := t.SatLocations[satIndex][t.TimeIndex]
		xCoords = append(xCoords, SatCoord{SatelliteID: satIndex, Dim: satPosition.X})
		yCoords = append(yCoords, SatCoord{SatelliteID: satIndex, Dim: satPosition.Y})
		zCoords = append(zCoords, SatCoord{SatelliteID: satIndex, Dim: satPosition.Z})
	}

	xPairs := t.findDimPairs(xCoords)
	yPairs := t.findDimPairs(yCoords)
	zPairs := t.findDimPairs(zCoords)

	intersection := intersectSets(xPairs, yPairs)
	intersection = intersectSets(intersection, zPairs)

	// Filter out pairs that are 0 distance
	closePairs := make(map[SatPair]struct{})
	for pair := range intersection {
		pos1 := t.SatLocations[pair.ID1][t.TimeIndex]
		pos2 := t.SatLocations[pair.ID2][t.TimeIndex]
		dist := distanceBetweenPositions(pos1, pos2)
		if dist != 0 {
			closePairs[pair] = struct{}{}
		}
	}

	return closePairs
}

func (t *TimeCluster) getAllSatIdsInNeighborCluster(clusterKey ClusterKey) []int {

	// Make sure to include the cluster itself
	neighbors := []ClusterKey{
		{X: clusterKey.X, Y: clusterKey.Y, Z: clusterKey.Z},
		{X: clusterKey.X + 1, Y: clusterKey.Y, Z: clusterKey.Z},
		{X: clusterKey.X - 1, Y: clusterKey.Y, Z: clusterKey.Z},
		{X: clusterKey.X, Y: clusterKey.Y + 1, Z: clusterKey.Z},
		{X: clusterKey.X, Y: clusterKey.Y - 1, Z: clusterKey.Z},
		{X: clusterKey.X, Y: clusterKey.Y, Z: clusterKey.Z + 1},
		{X: clusterKey.X, Y: clusterKey.Y, Z: clusterKey.Z - 1},
	}

	allSatIndexes := []int{}
	for _, neighborKey := range neighbors {
		if _, ok := t.Clusters[neighborKey]; ok {
			allSatIndexes = append(allSatIndexes, t.Clusters[neighborKey]...)
		}
	}

	return allSatIndexes
}

func (t *TimeCluster) findDimPairs(satCoords []SatCoord) map[SatPair]struct{} {

	// fmt.Println(satCoords)

	// First sort the coords
	sort.Slice(satCoords, func(i, j int) bool {
		return satCoords[i].Dim < satCoords[j].Dim
	})

	satPairs := make(map[SatPair]struct{})

	// Then find the pairs
	for i := 0; i < len(satCoords); i++ {
		for j := i + 1; j < len(satCoords); j++ {
			if satCoords[j].Dim-satCoords[i].Dim <= MAX_DIST {
				satPairs[NewSatPair(satCoords[i].SatelliteID, satCoords[j].SatelliteID)] = struct{}{}
			} else {
				break
			}
		}
	}

	return satPairs
}

type SatPair struct {
	ID1, ID2 int
}

// Helper function to create a canonical pair
func NewSatPair(id1, id2 int) SatPair {
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	return SatPair{ID1: id1, ID2: id2}
}

// TODO: change this to use struct
func createClusterKey(position SatPosition) ClusterKey {
	xIndex := int(position.X / BOX_SIZE)
	yIndex := int(position.Y / BOX_SIZE)
	zIndex := int(position.Z / BOX_SIZE)
	return ClusterKey{X: xIndex, Y: yIndex, Z: zIndex}
}

func intersectSets(set1, set2 map[SatPair]struct{}) map[SatPair]struct{} {
	intersection := make(map[SatPair]struct{})
	for pair := range set1 {
		if _, exists := set2[pair]; exists {
			intersection[pair] = struct{}{}
		}
	}
	return intersection
}
