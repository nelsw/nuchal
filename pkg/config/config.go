package config

import (
	"github.com/tkanos/gonfig"
	"math"
)

type Configuration struct {
	DSN      string  `json:"dsn"`
	StopGain float64 `json:"stop_gain"`
	StopLoss float64 `json:"stop_loss"`
	Tweezer  float64 `json:"tweezer"`
	MakerFee float64 `json:"maker_fee"`
	TakerFee float64 `json:"taker_fee"`
}

var config Configuration

func init() {
	if err := gonfig.GetConf("./.app/config.json", &config); err != nil {
		panic(err)
	}
}

func IsTweezer(thatLow, thatClose, thisLow, thisOpen float64) bool {
	return math.Abs(math.Min(thatLow, thatClose)-math.Min(thisLow, thisOpen)) <= config.Tweezer
}

func PricePlusStopGain(price float64) float64 {
	price += config.StopGain
	return price
}

func PriceMinusStopLoss(price float64) float64 {
	price -= config.StopLoss
	return price
}

func DatabaseUrl() string {
	return config.DSN
}

func Fee() float64 {
	return 0.005
}
