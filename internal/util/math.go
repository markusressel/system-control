package util

import (
	"cmp"
	"math"
)

// Clamp restricts a value to be within the range [min, max].
func Clamp[T cmp.Ordered](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// RoundTo2DP rounds a float64 to 2 decimal places.
func RoundTo2DP(value float64) float64 {
	ratio := math.Pow(10, 2)
	return math.Round(value*ratio) / ratio
}
