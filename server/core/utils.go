package core

import (
	"math"
)

type StrSlice []string

func (s StrSlice) Pos(value string) int {
	for p, v := range s {
		if v == value {
			return p
		}
	}
	return -1
}

func MinMaxfloat64(v float64, min float64, max float64) float64 {
	return float64(math.Min(math.Max(float64(v), float64(min)), float64(max)))
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
