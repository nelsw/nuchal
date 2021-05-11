package product

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	All      []Product
	Simulate string
	Trades   []string
}

var config Configuration

func init() {
	if file, err := os.Open("./.app/product/config.json"); err != nil {
		panic(err)
	} else if err = json.NewDecoder(file).Decode(&config); err != nil {
		panic(err)
	}
}

func PricePlusStopGain(productId string, price float64) float64 {
	price += price * findProduct(productId).stopGain()
	return price
}

func PriceMinusStopLoss(productId string, price float64) float64 {
	price -= price * findProduct(productId).stopLoss()
	return price
}

func Size(price float64) string {
	if price < 1 {
		return "10"
	} else if price < 2 {
		return "5"
	} else {
		return "1"
	}
}

func findProduct(id string) Product {
	for _, product := range config.All {
		if product.Id == id {
			return product
		}
	}
	panic(fmt.Sprintf("no product found for id [%s]", id))
}

func IdToSimulateTrade() string {
	return config.Simulate
}

func IdsToTrade() []string {
	return config.Trades
}
