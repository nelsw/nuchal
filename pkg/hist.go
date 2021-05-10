package pkg

import (
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
	asc     = "unix asc"
	desc    = "unix desc"
	query   = "product_id = ?"
	timeVal = "2021-04-01T00:00:00+00:00"
)

func init() {
	if err := db.AutoMigrate(&Rate{}); err != nil {
		panic(err)
	}
}

func Rates(name, productId string) []Rate {

	var rate Rate
	db.Where(query, productId).Order(desc).First(&rate)

	var from time.Time
	if rate != (Rate{}) {
		from = rate.Time()
	} else {
		from, _ = time.Parse(time.RFC3339, timeVal)
	}

	db.Save(NewRates(name, productId, from))

	var allRates []Rate
	db.Where(query, productId).Order(asc).Find(&allRates)
	return allRates
}
