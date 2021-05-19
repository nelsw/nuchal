package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"time"
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
	rounded := fmt.Sprintf("%.2f", x)

	chunks := strings.Split(rounded, `.`)
	dollars := chunks[0]
	cents := chunks[1]

	places := len(dollars)

	if places < 4 {
		return fmt.Sprintf("$%s.%s", dollars, cents)
	}

	pivot := places - 3
	var newFields []string
	for i, oldField := range dollars {
		if i == pivot {
			newFields = append(newFields, ",")
		}
		newFields = append(newFields, string(oldField))
	}
	rounded = strings.Join(newFields, ``)
	return fmt.Sprintf("$%s.%s", rounded, cents)
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

func Sleep(d time.Duration) {
	exit := time.Now().Add(d)
	for {
		log.Info().Msg("...")
		time.Sleep(d)
		if time.Now().After(exit) {
			break
		}
		d = time.Duration(d.Nanoseconds() / 2)
	}
}

func sleep(n int64) {
	time.Sleep(time.Nanosecond * time.Duration(n))
}
