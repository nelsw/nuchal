package util

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Log(username, productId, message string, v ...interface{}) {
	if v == nil || len(v) == 0 {
		fmt.Println(fmt.Sprintf("%s - %s - %s", username, productId, message))
		return
	}

	fmt.Println(fmt.Sprintf("%s - %s - %s [%s]", username, productId, message, Pretty(v)))
}

func Pretty(v interface{}) string {
	b, _ := json.MarshalIndent(&v, "", "  ")
	return string(b)
}

func Float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	} else {
		return f
	}
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f) // todo - get increment units dynamically from cb api
}
