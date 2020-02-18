package cryptolib

import "fmt"

// HumanHash returns a human readable hash/second value from a very large number.
func HumanHash(val float64) string {
	unit := "H/s"

	if val >= 1e+21 {
		val = val / 1e+21
		unit = "ZH/s"
	} else if val >= 1e+18 {
		val = val / 1e+18
		unit = "EH/s"
	} else if val >= 1e+15 {
		val = val / 1e+15
		unit = "PH/s"
	} else if val >= 1e+12 {
		val = val / 1e+12
		unit = "TH/s"
	} else if val >= 1e+9 {
		val = val / 1e+9
		unit = "GH/s"
	} else if val >= 1e+6 {
		val = val / 1e+6
		unit = "MH/s"
	} else if val >= 1e+3 {
		val = val / 1e+3
		unit = "kH/s"
	}

	return fmt.Sprintf("%.2f %s", val, unit)
}
