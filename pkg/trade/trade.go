package trade

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"nchl/pkg/coinbase"
	"nchl/pkg/model/order"
	"nchl/pkg/model/product"
	"nchl/pkg/model/rate"
	"time"
)

func CreateTrades(username, productId string) error {

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
		this, err := buildRate(wsConn, productId)
		if err != nil {
			return err
		}

		if rate.IsTweezer(this, then, that) {

			id := coinbase.CreateOrder(username, order.NewMarketBuyOrder(productId, product.Size(this.Close)))
			if id == nil { // error occurred and was logged, likely out of funds ... todo
				continue
			}
			size, price := coinbase.GetOrderSizeAndPrice(username, productId, *id)
			_ = coinbase.CreateOrder(username, order.NewStopEntryOrder(productId, size, product.PricePlusStopGain(productId, price)))
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
