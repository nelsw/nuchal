package model

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
)

// Product is an aggregate of a Coinbase Product and the type of pattern to apply towards trading the product.
type Product struct {
	cb.Product
	Pattern
}

func (p *Product) MarketEntryOrder() *cb.Order {
	return &cb.Order{
		ProductID: p.Id,
		Side:      "buy",
		Size:      p.Size,
		Type:      "market",
	}
}

func (p *Product) StopEntryOrder(price float64, size string) *cb.Order {
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

func (p *Product) StopGainOrder(fill cb.Fill) *cb.Order {
	exit := util.Float64(fill.Price)
	gain := exit + (exit * p.GainFloat())
	return p.StopEntryOrder(gain, fill.Size)
}

func (p *Product) StopLossOrder(price float64, size string) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: p.Id,
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
