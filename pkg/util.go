package pkg

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func log(username, productId, message string, v ...interface{}) {
	if v == nil || len(v) == 0 {
		fmt.Println(fmt.Sprintf("%s - %s - %s", username, productId, message))
		return
	}

	fmt.Println(fmt.Sprintf("%s - %s - %s [%s]", username, productId, message, pretty(v)))
}

func pretty(v interface{}) string {
	b, _ := json.MarshalIndent(&v, "", "  ")
	return string(b)
}

func float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	} else {
		return f
	}
}

func formatPrice(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
