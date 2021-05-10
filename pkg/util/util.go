package util

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 32); err != nil {
		panic(err)
	} else {
		return f
	}
}

func num2(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f)
}

func Print(v interface{}) string {
	b, _ := json.MarshalIndent(&v, "", "  ")
	return string(b)
}
