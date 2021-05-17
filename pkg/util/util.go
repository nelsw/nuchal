package util

import (
	"fmt"
	"os"
	"strconv"
)

func Int(s string) int {
	if i, err := strconv.Atoi(s); err != nil {
		return -1
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

func MinInt(a, z int) int {
	if a < z {
		return a
	} else {
		return z
	}
}

func Usd(f float64) string {
	x := (f * 100) + 0.5
	x = x / 100
	return fmt.Sprintf("$%.2f", x)
}

func FirstIntOrZero(arr []int) int {
	if arr != nil && len(arr) > 0 {
		return arr[0]
	}
	return 0
}

func IsInsufficientFunds(err error) bool {
	return err != nil && err.Error() == "Insufficient funds"
}

func DoIndefinitely(fun func()) error {
	exit := make(chan string)
	go fun()
	for {
		select {
		case <-exit:
			return nil
		}
	}
}

func IsTestMode() bool {
	return os.Getenv("MODE") == "test"
}

func IsZero(s string) bool {
	return Float64(s) == 0.0
}
