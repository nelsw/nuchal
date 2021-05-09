package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Fee     = 0.005
	Won     = "🟢"
	Lost    = "🔴"
	Dmnd = "💎"
	Poop    = "💩"
	Twzr  = "🎯"
	Gain    = "🔺"
	Loss    = "🔻"
)

func Product(symbol string) string {
	return strings.ToUpper(symbol) + toUSD
}

func Float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 32); err != nil {
		panic(err)
	} else {
		return f
	}
}

func Int(s string) int {
	if i, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return i
	}
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f)
}