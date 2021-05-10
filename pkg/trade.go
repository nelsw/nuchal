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

		log(username, productId, "analyzing rates")

		start := time.Now()
		end := start.Add(time.Minute)
		this := rate(wsConn, productId, end)

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {

			log(username, productId, "pattern recognized")
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)

			if math.Abs(thatFloor-thisFloor) <= 0.01 {
				log(username, productId, "tweezer in range")
				price, size := CreateMarketBuyOrder(username, productId, size(this.Close))
				go climb(price, username, productId, size)
			} else {
				log(username, productId, "tweezer out of range")
			}
		}

		then = that
		that = this
		start = end
		end = start.Add(time.Minute)
	}
}

func climb(marketPrice float64, username, productId, size string) {

	log(username, productId, "climb started")

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
		stopPrice = GetPrice(wsConn, productId)
		if stopPrice >= marketPrice+(marketPrice*stopGain) {
			log(username, productId, "default stop price found")
			orderId = CreateStopLossOrder(username, productId, size, stopPrice)
			break
		}
	}

	for {

		start := time.Now()
		end := start.Add(time.Minute)

		log(username, productId, "analyzing rakes")

		rate := rate(wsConn, productId, end)
		if rate.Low <= stopPrice {
			log(username, productId, "stop loss executed")
			break
		}

		if rate.Close > stopPrice {
			log(username, productId, "better stop price found")
			stopPrice = rate.Close
			CancelOrder(username, productId, orderId)
			orderId = CreateStopLossOrder(username, productId, size, stopPrice)
		}

		start = end
		end = start.Add(time.Minute)
	}

	log(username, productId, "climb completed")
}

func rate(wsConn *ws.Conn, productId string, end time.Time) Rate {
	rate := Rate{}
	for {
		price := GetPrice(wsConn, productId)
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
