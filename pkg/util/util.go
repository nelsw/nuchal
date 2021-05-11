package util

import (
	"strconv"
)

func Float64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	} else {
		return f
	}
}
