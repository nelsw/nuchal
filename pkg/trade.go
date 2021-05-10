package pkg

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"math"
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

	var then, that Rate
	for {

		fmt.Println("analyzing rates")

		start := time.Now()
		end := start.Add(time.Minute)
		this := rate(wsConn, productId, end)

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {

			fmt.Println("pattern recognized")

			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)

			if math.Abs(thatFloor-thisFloor) <= 0.01 {
				fmt.Println("tweezer in range")
				price, size := CreateMarketBuyOrder(username, productId, size(this.Close))
				CreateEntryOrder(username, productId, size, price)
			} else {
				fmt.Println("tweezer out of range")
			}
		}

		then = that
		that = this
		start = end
		end = start.Add(time.Minute)
	}
}

// todo
func rake(marketPrice float64, username, productId, size string) {

	fmt.Println("rake started")

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
	var orderId string
	for {
		stopPrice = GetTickerPrice(wsConn, productId)
		if stopPrice >= marketPrice+(marketPrice*0.0195) {
			fmt.Println("default stop price found")
			orderId = CreateStopLossOrder(username, productId, size, stopPrice)
			break
		}
	}

	for {

		start := time.Now()
		end := start.Add(time.Minute)

		fmt.Println("analyzing rakes")

		rate := rate(wsConn, productId, end)
		if rate.Low <= stopPrice {
			fmt.Println("stop loss executed")
			break
		}

		if rate.Close > stopPrice {
			fmt.Println("better stop price found")
			stopPrice = rate.Close
			CancelOrder(username, orderId)
			orderId = CreateStopLossOrder(username, productId, size, stopPrice)
		}

		start = end
		end = start.Add(time.Minute)
	}

	fmt.Println("rake completed")
}

func rate(wsConn *ws.Conn, productId string, end time.Time) Rate {

	rate := Rate{}
	for {
		price := GetTickerPrice(wsConn, productId)
		if rate.Low == 0 {
			rate.Low = price
			rate.High = price
			rate.Open = price
		} else if rate.High < price {
			rate.High = price
		} else if rate.Low > price {
			rate.Low = price
		}

		if time.Now().After(end) {
			rate.Close = price
			break
		}
	}
	return rate
}

func size(price float64) string {
	if price < 1 {
		return "100"
	} else if price < 2 {
		return "10"
	} else {
		return "1"
	}
}
