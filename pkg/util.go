package pkg

import (
	"encoding/json"
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	stopGain = 0.0295
	stopLoss = 0.395
)

func size(price float64) string {
	if price < 1 {
		return "1"
	} else if price < 2 {
		return "1"
	} else {
		return "1"
	}
}

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

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func formatPrice(f float64) string {
	return fmt.Sprintf("%.3f", f)
}

func float6(f float64) string {
	return fmt.Sprintf("%.6f", f)
}

func getClient(username string) *cb.Client {
	key, pass, secret := GetUserConfig(username)
	return &cb.Client{
		"https://api.pro.coinbase.com",
		*secret,
		*key,
		*pass,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}
