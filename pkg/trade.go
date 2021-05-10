package pkg

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"math"
	"time"
)

const (
	Fee   = 0.005
	Hi    = 0.0295
	Lo    = 0.495
	Twz   = 0.01
	WsUrl = "wss://ws-feed.pro.coinbase.com"
)

func CreateTrades(username, productId string) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial(WsUrl, nil)
	if err != nil {
		panic(err)
	}
	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			fmt.Println(err)
		}
	}(wsConn)

	from := time.Now().Round(time.Minute)
	to := from.Add(time.Minute)

	var then, that Rate

	for {

		this := rate(wsConn, productId, to)

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {

			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)

			if math.Abs(thatFloor-thisFloor) <= Twz {
				price, size := CreateMarketOrder(username, productId, size(this.Close))
				CreateEntryOrder(username, productId, size, price+(price*Hi))
			}
		}

		then = that
		that = this
		from = to
		to = from.Add(time.Minute)
	}
}

func rate(wsConn *ws.Conn, productId string, to time.Time) Rate {
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

		if time.Now().After(to) {
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
