package statistic

import (
	"github.com/rs/zerolog/log"
	"math"
	"time"
)

type Candlestick struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductId string  `json:"product_id" gorm:"primaryKey"`
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

func IsTweezer(t, u, v Candlestick, d float64) bool {
	return isTweezerPattern(t, u, v) && isTweezerValue(t, u, d)
}

func isTweezerValue(u, v Candlestick, d float64) bool {
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= d
	if b {
		log.Info().
			Str("productId", v.ProductId).
			Float64("tweezer", d-f)
	}
	return b
}

func isTweezerPattern(t, u, v Candlestick) bool {
	return t.IsInit() && u.IsInit() && t.IsDown() && u.IsDown() && v.IsUp()
}
