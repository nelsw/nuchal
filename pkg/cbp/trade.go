package cbp

import (
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
)

type Trade struct {
	cb.Fill
}

func NewTrade(cbFill cb.Fill) *Trade {
	trade := new(Trade)
	trade.Fill = cbFill
	return trade
}

func (t Trade) Price() float64 {
	return util.Float64(t.Fill.Price)
}

func (t Trade) Size() float64 {
	return util.Float64(t.Fill.Size)
}

func (t Trade) Total() float64 {
	return t.Price() * t.Size()
}
