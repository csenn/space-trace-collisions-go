package main

import "math"

func createJulianDate(year, mon, day, hr, min, sec int) float64 {
	return (367.0*float64(year) - math.Floor((7*(float64(year)+math.Floor((float64(mon)+9)/12.0)))*0.25) + math.Floor(275*float64(mon)/9.0) + float64(day) + 1721013.5 + ((float64(sec)/60.0+float64(min))/60.0+float64(hr))/24.0)
}

func julianDateAddSeconds(julianDate float64, seconds float64) float64 {
	deltaDays := seconds / 86400.0
	return julianDate + deltaDays
}

func differenceInSeconds(julianDate1, julianDate2 float64) float64 {
	// Calculate the difference in days
	differenceInDays := julianDate2 - julianDate1
	// 86400 seconds in a day
	return differenceInDays * 86400
}

func julianDateToUTC50(julianDate float64) float64 {
	return julianDate - 2433281.5
}
