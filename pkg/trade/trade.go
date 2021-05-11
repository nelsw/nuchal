package trade

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"nchl/pkg/coinbase"
	"nchl/pkg/config"
	"nchl/pkg/history"
	"nchl/pkg/order"
	"nchl/pkg/product"
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

	var then, that history.Rate
	for {

		util.Log(username, productId, "analyzing rates")

		start := time.Now()
		end := start.Add(time.Minute)
		this := rate(wsConn, productId, end)

		if then != (history.Rate{}) && that != (history.Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			util.Log(username, productId, "pattern recognized")
			if config.IsTweezer(that.Low, that.Close, this.Low, this.Close) {
				util.Log(username, productId, "tweezer out of range")
				continue
			}
			util.Log(username, productId, "tweezer in range")
			if id, err := coinbase.CreateOrder(username, order.NewMarketBuyOrder(productId, product.Size(this.Close))); err == nil {
				size, price := coinbase.GetOrderSizeAndPrice(username, productId, *id)
				_, _ = coinbase.CreateOrder(username, order.NewStopEntryOrder(productId, size, config.PricePlusStopGain(price)))
			}
		}

		then = that
		that = this
		start = end
		end = start.Add(time.Minute)
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
		if stopPrice >= config.PricePlusStopGain(marketPrice) {
			util.Log(username, productId, "default stop price found")
			orderId, err := coinbase.CreateOrder(username, order.NewStopLossOrder(productId, size, stopPrice))
			if err != nil {
				continue
			}
			for {

				start := time.Now()
				end := start.Add(time.Minute)

				util.Log(username, productId, "analyzing rakes")

				rate := rate(wsConn, productId, end)
				if rate.Low <= stopPrice {
					util.Log(username, productId, "stop loss executed")
					break
				}

				if rate.Close > stopPrice {
					util.Log(username, productId, "better stop price found")
					stopPrice = rate.Close
					coinbase.CancelOrder(username, productId, *orderId)
					_, _ = coinbase.CreateOrder(username, order.NewStopLossOrder(productId, size, stopPrice))
				}

				start = end
				end = start.Add(time.Minute)
			}

			util.Log(username, productId, "climb completed")
		}
	}
}

func rate(wsConn *ws.Conn, productId string, end time.Time) history.Rate {

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
			return history.Rate{
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
