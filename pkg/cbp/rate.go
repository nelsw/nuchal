package cbp

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

type Rate struct {
	Unix      int64  `json:"unix" gorm:"primaryKey"`
	ProductId string `json:"product_id" gorm:"primaryKey"`
	cb.HistoricRate
}

func NewRate(productId string, historicRate cb.HistoricRate) *Rate {
	rate := new(Rate)
	rate.Unix = historicRate.Time.UnixNano()
	rate.ProductId = productId
	rate.HistoricRate = historicRate
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

func (v *Rate) Label() string {
	return time.Unix(0, v.Unix).Format(time.Kitchen)
}

func (v *Rate) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}
