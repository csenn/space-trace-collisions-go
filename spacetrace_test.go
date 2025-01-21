package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const SatOneLineOne = "1 84232U 79104    25011.29418726 +.00010894 +00000+0 +10327-1 0  9993"
const SatOneLineTwo = "2 84232  20.2440 103.5465 6615434  81.8936 342.8433  3.09154996 94164"

// These two satellites have a close collision distance
const SatTwoLineOne = "1 56700U 23067N   25011.12006866 -.00000852  00000-0 -48043-4 0  9994"
const SatTwoLineTwo = "2 56700  43.0052  50.6716 0001256 262.8432  97.2268 15.02525502 92091"

const SatThreeLineOne = "1 58247U 23171T   25011.52048310  .00003171  00000-0  24954-3 0  9991"
const SatThreeLineTwo = "2 58247  43.0041  59.6580 0001638 274.8194  85.2461 15.02562597 66220"

const CloseCollisionTime = 2460688.299389648
const CloseCollisionDistance = 0.21177285731942194

func round4Decimals(num float64) float64 {
	return math.Round(num*10000) / 10000.0
}

var satOne = NewSgp4Satellite(SatOneLineOne, SatOneLineTwo)
var satTwo = NewSgp4Satellite(SatTwoLineOne, SatTwoLineTwo)
var satThree = NewSgp4Satellite(SatThreeLineOne, SatThreeLineTwo)

func TestMatchPythonSgp4(t *testing.T) {

	julianDate := createJulianDate(2025, 1, 12, 0, 0, 0)
	pos, _ := satOne.propagateAtTime(julianDate)

	assert.Equal(t, round4Decimals(pos.X), round4Decimals(12705.022786314943))
	assert.Equal(t, round4Decimals(pos.Y), round4Decimals(-13783.223025751715))
	assert.Equal(t, round4Decimals(pos.Z), round4Decimals(-3409.846262052983))
}

func TestCloseCollisionDistanceAtSpecificTime(t *testing.T) {

	sat1Pos, _ := satTwo.propagateAtTime(CloseCollisionTime)
	sat2Pos, _ := satThree.propagateAtTime(CloseCollisionTime)

	distance := distanceBetweenPositions(sat1Pos, sat2Pos)

	assert.Equal(t, round4Decimals(CloseCollisionDistance), round4Decimals(distance))
}

func TestFindCollisionTimeWithBinarySearch(t *testing.T) {
	startTime := julianDateAddSeconds(CloseCollisionTime, -2*60)

	timeLeft := julianDateAddSeconds(startTime, -10*60)
	timeRight := julianDateAddSeconds(startTime, 10*60)

	collisionTime, _ := binarySearch(satTwo, satThree, timeLeft, timeRight)
	collisionDistance, _ := distanceBetweenSatellites(satTwo, satThree, collisionTime)

	assert.Equal(t, round4Decimals(CloseCollisionTime), round4Decimals(collisionTime))
	assert.Equal(t, round4Decimals(CloseCollisionDistance), round4Decimals(collisionDistance))
}
