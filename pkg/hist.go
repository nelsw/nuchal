package pkg

import (
	"fmt"
	"time"
)

type Rate struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductId string  `json:"product" gorm:"primaryKey"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (v *Rate) IsDown() bool {
	return v.Open > v.Close
}

func (v *Rate) IsUp() bool {
	return !v.IsDown()
}

func (v *Rate) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Rate) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}

const (
	query = "product_id = ?"
	desc  = "unix desc"
	asc   = "unix asc"
	year  = 2021
	month = 5
	day   = 1
	hour  = 0
	min   = 0
	sec   = 0
	nsc   = 0
)

var (
	from, to time.Time
	rates    []Rate
)

func init() {
	if err := db.AutoMigrate(&Rate{}); err != nil {
		panic(err)
	}
}

func SetupRates() {
	fmt.Println("setting up rates for", target.ProductId)
	db.Where(query, target.ProductId).Order(asc).Find(&rates)
	if len(rates) == 0 {
		fmt.Println("no rates found for", target.ProductId)
		setupTimes()
		rates = BuildRates()
		fmt.Println("built rates", len(rates))
		for _, rate := range rates {
			db.Save(rate)
		}
		fmt.Println("saved rates")
	}
	fmt.Println()
}

func setupTimes() {

	fmt.Println("setting up times to build rates")

	from = time.Date(year, month, day, hour, min, sec, nsc, time.UTC)
	to = from.Add(time.Hour * 4)

	var rate Rate
	db.Where(query, target.ProductId).Order(desc).First(&rate)
	if rate == (Rate{}) {
		fmt.Println("no rate found for", target.ProductId)
	} else {
		fmt.Println("rate found", Print(rate))
		from = rate.Time()
	}

	fmt.Println("setup times to build rates")
	fmt.Println()
}
