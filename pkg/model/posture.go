package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nuchal/pkg/util"
)

// Posture is the type of position to take on a currency.
type Posture struct {
	cb.Product
	Pattern
}

func (p *Posture) ProductId() string {
	return p.Product.ID
}

func (p *Posture) MarketEntryOrder() *cb.Order {
	return &cb.Order{
		ProductID: p.ProductId(),
		Side:      "buy",
		Size:      fmt.Sprintf("%.3f", util.Float64(p.BaseMinSize)*3),
		Type:      "market",
	}
}

func (p *Posture) StopEntryOrder(price float64, size string) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: p.Id,
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: Price(price),
		Stop:      "entry",
	}
}

func (p *Posture) StopGainOrder(fill cb.Fill) *cb.Order {
	exit := util.Float64(fill.Price)
	gain := exit + (exit * p.GainFloat())
	return p.StopEntryOrder(gain, fill.Size)
}

func (p *Posture) StopLossOrder(price float64, size string) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: p.ProductId(),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: Price(price),
		Stop:      "loss",
	}
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f) // todo - get increment units dynamically from cb api
}
