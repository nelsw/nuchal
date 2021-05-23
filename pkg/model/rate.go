package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

type Rate struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductId string  `json:"product_id" gorm:"primaryKey"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func NewRate(productId string, historicRate cb.HistoricRate) *Rate {
	rate := new(Rate)
	rate.Unix = historicRate.Time.UnixNano()
	rate.ProductId = productId
	rate.Low = historicRate.Low
	rate.High = historicRate.High
	rate.Open = historicRate.Open
	rate.Close = historicRate.Close
	rate.Volume = historicRate.Volume
	return rate
}

func (v *Rate) IsDown() bool {
	return v.Open > v.Close
}

func (v *Rate) IsUp() bool {
	return !v.IsDown()
}

func (v *Rate) IsInit() bool {
	return v != nil && v != (&Rate{})
}

func (v *Rate) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Rate) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}
