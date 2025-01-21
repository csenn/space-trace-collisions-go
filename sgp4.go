package main

// #cgo CFLAGS: -I${SRCDIR}/sgp4/wrappers
// #cgo LDFLAGS: -L${SRCDIR}/sgp4/mac_m1_gfortran -ldllmain -lastrofunc -ltimefunc -lenvconst -ltle -lsgp4prop
// #include "DllMainDll.h"
// #include "AstroFuncDll.h"
// #include "TimeFuncDll.h"
// #include "EnvConstDll.h"
// #include "TleDll.h"
// #include "Sgp4PropDll.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

type SatPosition struct {
	X float64
	Y float64
	Z float64
}

type Spg4Satellite struct {
	TLE1   string
	TLE2   string
	satKey C.long
}

func NewSgp4Satellite(tle1, tle2 string) Spg4Satellite {
	line1 := C.CString(tle1)
	line2 := C.CString(tle2)
	defer C.free(unsafe.Pointer(line1))
	defer C.free(unsafe.Pointer(line2))

	satKey := C.TleAddSatFrLines(line1, line2)

	ErrCode := C.Sgp4InitSat(satKey)
	if ErrCode != 0 {
		fmt.Println("Error initializing SGP4")
		exitErr()
	}

	return Spg4Satellite{TLE1: tle1, TLE2: tle2, satKey: satKey}
}

func (s *Spg4Satellite) propagateAtTime(julianDate float64) (SatPosition, error) {

	// julianDate := satellite.JDay(year, month, day, hours, minutes, seconds)
	// utc50Date := julianDate - 2433281.5

	pos := make([]C.double, 3)
	utc50Date := julianDateToUTC50(julianDate)

	ErrCode := C.Sgp4PropDs50UtcPos(s.satKey, C.double(utc50Date), &pos[0])
	if ErrCode != 0 {
		// Go-managed buffer
		lastErrMsg := make([]byte, 128)
		// Pass slice to C
		C.GetLastErrMsg((*C.char)(unsafe.Pointer(&lastErrMsg[0])))
		// Convert to Go string
		errMsg := string(lastErrMsg)
		return SatPosition{}, fmt.Errorf("propagation error: %s", strings.TrimSpace(errMsg))
	}

	return SatPosition{
		X: float64(pos[0]),
		Y: float64(pos[1]),
		Z: float64(pos[2]),
	}, nil
}

// Destroys the satellite from the underlying SGP4 library
func (s *Spg4Satellite) destroySat() {
	C.Sgp4RemoveSat(s.satKey)
}

func allocstr(length int) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteByte(0)
	}
	str := b.String()
	return str
}

func exitErr() {
	lastErrMsg := C.CString(allocstr(128))
	C.GetLastErrMsg(lastErrMsg)
	fmt.Println(C.GoString(lastErrMsg))
	os.Exit(0)
}

/*


//Test strings

func runSgp4() {
	//Test strings
	infoStr := C.CString(allocstr(128))
	C.DllMainGetInfo(infoStr)
	fmt.Println(strings.TrimSpace(C.GoString(infoStr)))
	C.free(unsafe.Pointer(infoStr)) //Need to free the string after use

	ThetaG := C.ThetaGrnwchFK5(1.520833333333)
	fmt.Println(ThetaG)

	//In case the length matters
	//temp1 := fmt.Sprintf("%-512v", "1 90021U RELEAS14 00051.47568104 +.00000184 +00000+0 +00000-4 0 0814")
	//temp2 := fmt.Sprintf("%-512v", "2 90021   0.0222 182.4923 0000720  45.6036 131.8822  1.00271328 1199")
	// line1 := C.CString("1 90021U RELEAS14 00051.47568104 +.00000184 +00000+0 +00000-4 0 0814")
	// line2 := C.CString("2 90021   0.0222 182.4923 0000720  45.6036 131.8822  1.00271328 1199")
	line1 := C.CString("1 84232U 79104    25011.29418726 +.00010894 +00000+0 +10327-1 0  9993")
	line2 := C.CString("2 84232  20.2440 103.5465 6615434  81.8936 342.8433  3.09154996 94164")
	// "EPOCH": "2025-01-11T07:03:37.779264",

	satKey := C.TleAddSatFrLines(line1, line2)

	C.free(unsafe.Pointer(line1)) //Need to free the string after use
	C.free(unsafe.Pointer(line2)) //Need to free the string after use
	fmt.Println(satKey)

	fmt.Println("Initializing SGP4 - hwllo")
	ErrCode := C.Sgp4InitSat(satKey)
	if ErrCode != 0 {
		exitErr()
	}

	var ds50UTC C.double
	var yr C.int
	var day C.double
	pos := make([]C.double, 3)    //Do I need to convert back to Float64?
	vel := make([]C.double, 3)    //Do I need to convert back to Float64?
	llh := make([]C.double, 3)    //Do I need to convert back to Float64?
	posnew := make([]C.double, 3) //Do I need to convert back to Float64?
	velnew := make([]C.double, 3) //Do I need to convert back to Float64?
	sgp4MeanKep := make([]C.double, 6)
	for mse := 0.0; mse <= 43200.0; mse += 2700.0 {
		ErrCode = C.Sgp4PropMse(satKey, C.double(mse), &ds50UTC, &pos[0], &vel[0], &llh[0])
		if ErrCode != 0 {
			exitErr()
		}
		fmt.Printf(" %17.7f%17.7f%17.7f%17.7f%17.7f%17.7f%17.7f\n", mse, pos[0], pos[1], pos[2], vel[0], vel[1], vel[2])
		C.UTCToYrDays(ds50UTC, &yr, &day)
		fmt.Println("yr, day", yr, day)
		ErrCode = C.Sgp4PosVelToKep(yr, day, &pos[0], &vel[0], &posnew[0], &velnew[0], &sgp4MeanKep[0])
		fmt.Printf(" %17.7f%17.7f%17.7f%17.7f%17.7f%17.7f\n", sgp4MeanKep[0], sgp4MeanKep[1], sgp4MeanKep[2], sgp4MeanKep[3], sgp4MeanKep[4], sgp4MeanKep[5])
	}

	fmt.Println("FINAL")

	// -2433281.5
	// 2433282.5 = const to subtract
	// 2460687.5 = JAN 11 2025 0 0 0 0
	// 27405.0
	// ErrCode = C.Sgp4PropMse(satKey, C.double(0.0), &ds50UTC, &pos[0], &vel[0], &llh[0])
	ErrCode = C.Sgp4PropDs50UtcPos(satKey, C.double(27406.0), &pos[0])

	// fmt.Printf(" f%17.7f%17.7f%17.7f\n", pos[0], pos[1], pos[2])
	fmt.Println("FINAL-Res 3", pos)

}
*/
