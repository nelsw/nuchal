package rate

import (
	"github.com/rs/zerolog/log"
	"math"
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

func IsTweezer(t, u, v Candlestick, d float64) bool {
	b := isTweezerPattern(t, u, v) && isTweezerValue(t, u, d)
	log.Info().
		Str("productId", v.ProductId).
		Msgf("tweezer found? [%v]", b)
	return b
}

func isTweezerValue(u, v Candlestick, d float64) bool {
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= d
	s := ">"
	if b {
		s = "<="
	}
	log.Info().
		Str("productId", v.ProductId).
		Msgf("tweezer delta? [%v] [%f] %s [%f]", b, f, s, d)
	return b
}

func isTweezerPattern(t, u, v Candlestick) bool {
	b := t.IsInit() && u.IsInit() && t.IsDown() && u.IsDown() && v.IsUp()
	log.Info().
		Str("productId", v.ProductId).
		Msgf("tweezer trend? [%v]", b)
	return b
}
