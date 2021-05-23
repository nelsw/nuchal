package model

import (
	"github.com/rs/zerolog/log"
	"math"
	"nuchal/pkg/util"
)

type Pattern struct {
	Id     string `json:"id"`
	Gain   string `json:"gain"`
	Loss   string `json:"loss"`
	Delta  string `json:"delta"`
	Size   string `json:"size"`
	Enable bool   `json:"enable,omitempty"`
}

func (p *Pattern) GainPrice(price float64) float64 {
	return price + (price * util.Float64(p.Gain))
}

func (p *Pattern) LossPrice(price float64) float64 {
	return price - (price * util.Float64(p.Loss))
}

func (p *Pattern) DeltaFloat() float64 {
	return util.Float64(p.Delta)
}

func (p *Pattern) GainFloat() float64 {
	return util.Float64(p.Gain)
}

func (p *Pattern) LossFloat() float64 {
	return util.Float64(p.Loss)
}

func (p *Pattern) SizeFloat() float64 {
	return util.Float64(p.Size)
}

func IsTweezer(t, u, v Rate, d float64) bool {
	return isTweezerPattern(t, u, v) && isTweezerValue(u, v, d)
}

func IsTweezerTop(u, v Rate, d float64) bool {
	return isTweezerValue(u, v, d)
}

func isTweezerValue(u, v Rate, d float64) bool {
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= d
	if b {
		log.Info().
			Str("productId", v.ProductId).
			Float64("tweezer", d-f)
	}
	return b
}

func isTweezerPattern(t, u, v Rate) bool {
	return t.IsInit() && u.IsInit() && t.IsDown() && u.IsDown() && v.IsUp()
}