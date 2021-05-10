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
	b, _ := json.MarshalIndent(&v, "", "  ")
	fmt.Println(fmt.Sprintf("%s - %s - %s [%s]", username, productId, message, string(b)))
}

func float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 32); err != nil {
		panic(err)
	} else {
		return f
	}
}

func formatPrice(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
