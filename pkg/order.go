package pkg

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
)

type orderType string

func (t orderType) String() string {
	return string(t)
}

type orderSide string

func (t orderSide) String() string {
	return string(t)
}

type orderStop string

func (t orderStop) String() string {
	return string(t)
}

const (
	market orderType = "market"
	limit  orderType = "limit"
	buy    orderSide = "buy"
	sell   orderSide = "sell"
	loss   orderStop = "loss"
	entry  orderStop = "entry"
)

func NewMarketBuyOrder(productId, size string) *cb.Order {
	return &cb.Order{
		ProductID: productId,
		Side:      buy.String(),
		Size:      size,
		Type:      market.String(),
	}
}

func NewStopEntryOrder(productId, size string, price float64) *cb.Order {
	return &cb.Order{
		Price:     formatPrice(price),
		ProductID: productId,
		Side:      sell.String(),
		Size:      size,
		Type:      limit.String(),
		StopPrice: formatPrice(price),
		Stop:      entry.String(),
	}
}

func NewStopLossOrder(productId, size string, price float64) *cb.Order {
	return &cb.Order{
		Price:     formatPrice(price),
		ProductID: productId,
		Side:      sell.String(),
		Size:      size,
		Type:      limit.String(),
		StopPrice: formatPrice(price),
		Stop:      loss.String(),
	}
}
