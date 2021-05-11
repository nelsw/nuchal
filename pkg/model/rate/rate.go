package rate

import (
	"time"
)

type Candlestick struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductId string  `json:"product" gorm:"primaryKey"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (v *Candlestick) IsDown() bool {
	return v.Open > v.Close
}

func (v *Candlestick) IsUp() bool {
	return !v.IsDown()
}

func (v *Candlestick) IsInit() bool {
	return v != nil && v != (&Candlestick{})
}

func (v *Candlestick) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Candlestick) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}
