package model

import "nuchal/pkg/util"

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
