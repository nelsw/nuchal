package statistic

import "nchl/pkg/util"

type Pattern struct {
	Id     string `json:"id"`
	Gain   string `json:"gain"`
	Loss   string `json:"loss"`
	Delta  string `json:"delta"`
	Size   string `json:"size"`
	Enable bool   `json:"enable,omitempty"`
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
