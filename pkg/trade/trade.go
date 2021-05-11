package trade

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nchl/pkg/coinbase"
	"nchl/pkg/conf"
	"nchl/pkg/rate"
	"time"
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
		Price:     Price(price),
		ProductID: productId,
		Side:      sell.String(),
		Size:      size,
		Type:      limit.String(),
		StopPrice: Price(price),
		Stop:      entry.String(),
	}
}

func NewStopLossOrder(productId, size string, price float64) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: productId,
		Side:      sell.String(),
		Size:      size,
		Type:      limit.String(),
		StopPrice: Price(price),
		Stop:      loss.String(),
	}
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f) // todo - get increment units dynamically from cb api
}

func CreateTrades(user conf.User, product conf.Product) error {

	fmt.Println("creating trades")

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		panic(err)
	}
	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			fmt.Println(err)
		}
	}(wsConn)

	var then, that rate.Candlestick
	for {
		this, err := buildRate(wsConn, product.Id)
		if err != nil {
			return err
		}

		if rate.IsTweezer(this, then, that) {

			id := coinbase.CreateOrder(user, NewMarketBuyOrder(product.Id, conf.Size(this.Close)))
			if id == nil { // error occurred and was logged, likely out of funds ... todo
				continue
			}
			size, price := coinbase.GetOrderSizeAndPrice(user, product.Id, *id)
			_ = coinbase.CreateOrder(user, NewStopEntryOrder(product.Id, size, product.EntryPrice(price)))
		}

		then = that
		that = this
	}
}

func buildRate(wsConn *ws.Conn, productId string) (rate.Candlestick, error) {

	fmt.Println("building rate")

	end := time.Now().Add(time.Minute)

	var low, high, open, vol float64
	for {

		price, err := coinbase.GetPrice(wsConn, productId)
		if err != nil {
			return rate.Candlestick{}, err
		}

		vol++

		if low == 0 {
			low = price
			high = price
			open = price
		} else if high < price {
			high = price
		} else if low > price {
			low = price
		}

		if time.Now().After(end) {
			fmt.Println("built rate")
			return rate.Candlestick{
				time.Now().UnixNano(),
				productId,
				low,
				high,
				open,
				price,
				vol,
			}, nil
		}
	}
}
