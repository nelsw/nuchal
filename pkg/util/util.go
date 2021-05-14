package util

import (
	"fmt"
	"strconv"
)

func Int(s string) int {
	if i, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return i
	}
}

func Float64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	} else {
		return f
	}
}

func Usd(f float64) string {
	x := (f * 100) + 0.5
	x = x / 100
	return fmt.Sprintf("$%.2f", x)
}
