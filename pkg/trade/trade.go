package trade

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"nchl/pkg/coinbase"
	"nchl/pkg/model/order"
	"nchl/pkg/model/product"
	"nchl/pkg/model/rate"
	"nchl/pkg/util"
	"time"
)

func CreateTrades(username, productId string) {

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

		start := time.Now()
		end := start.Add(time.Minute)
		this := buildRate(wsConn, productId, end)

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

// todo
func climb(marketPrice float64, username, productId, size string) {

	util.Log(username, productId, "climb started")

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

	var stopPrice float64
	for {
		stopPrice = coinbase.GetPrice(wsConn, productId)
		if stopPrice >= product.PricePlusStopGain(productId, marketPrice) {
			util.Log(username, productId, "default stop price found")
			id := coinbase.CreateOrder(username, order.NewStopLossOrder(productId, size, stopPrice))
			if id != nil {
				continue
			}
			for {

				start := time.Now()
				end := start.Add(time.Minute)

				util.Log(username, productId, "analyzing rakes")

				r := buildRate(wsConn, productId, end)
				if r.Low <= stopPrice {
					util.Log(username, productId, "stop loss executed")
					break
				}

				if r.Close > stopPrice {
					util.Log(username, productId, "better stop price found")
					stopPrice = r.Close
					coinbase.CancelOrder(username, productId, *id)
					_ = coinbase.CreateOrder(username, order.NewStopLossOrder(productId, size, stopPrice))
				}

				start = end
				end = start.Add(time.Minute)
			}

			util.Log(username, productId, "climb completed")
		}
	}
}

func buildRate(wsConn *ws.Conn, productId string, end time.Time) rate.Candlestick {

	fmt.Println("building rate")

	var low, high, open, vol float64
	for {

		price := coinbase.GetPrice(wsConn, productId)
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
			}
		}
	}
}
