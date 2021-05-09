package pkg

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"math"
	"time"
)

func CreateTrades() {

	fmt.Println("setting up websocket for user", user)

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

	fmt.Println("set up websocket for user", user)
	fmt.Println()

	start := time.Now().Round(time.Minute)
	stop := start.Add(time.Minute)

	var then, that Rate

	for {

		this := buildRate(wsConn, start, stop)
		fmt.Println("processing rate", this)

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Max(thatFloor, thisFloor)-math.Min(thatFloor, thisFloor) <= target.Tweezer {
				createOrders()
			}
		}

		then = that
		that = this
		start = stop
		stop = start.Add(time.Minute)
	}
}

func buildRate(wsConn *ws.Conn, from, to time.Time) Rate {

	rate := Rate{
		Unix:   from.Unix(),
		Volume: 0,
	}

	for {
		p := GetPrice(wsConn)
		if rate.Low == 0 {
			rate.Low = p
			rate.High = p
			rate.Open = p
		} else if rate.High < p {
			rate.High = p
		} else if rate.Low > p {
			rate.Low = p
		}
		rate.Volume++

		if time.Now().After(to) {
			rate.Close = p
			return rate
		}
	}
}

func createOrders() {

	if order, err := createMarketOrder(); err != nil {
		panic(err)
	} else {

		settledOrder := getOrder(order.ID)

		size := settledOrder.Size
		price := Float(settledOrder.ExecutedValue)

		stopGain := Price(price + (price * target.Gain))
		if gainExit, err := createEntryOrder(size, stopGain); err != nil {
			fmt.Println("error creating gain exit order", err)
		} else {
			fmt.Println("successfully created gain exit order", gainExit)
		}
	}
}

func getOrder(id string) cb.Order {
	return attemptGetOrder(id, 0)
}

func cancelOrder(id string) {
	err := client.CancelOrder(id)
	if err != nil {
		panic(err)
	}
}

func attemptGetOrder(id string, attempt int) cb.Order {
	order, err := GetOrder(id)
	if err != nil {
		attempt++
		if attempt < 100 {
			return attemptGetOrder(id, attempt)
		}
		panic(err)
	}
	if !order.Settled {
		return attemptGetOrder(id, 0)
	}
	return order
}

func createMarketOrder() (*cb.Order, error) {
	return CreateOrder(&cb.Order{
		ProductID: target.ProductId,
		Side:      "buy",
		Size:      Size(),
		Type:      "market",
	}, 0)
}

func createEntryOrder(size, price string) (*cb.Order, error) {
	return CreateOrder(&cb.Order{
		Price:     price,
		ProductID: target.ProductId,
		Side:      "sell",
		Size:      size,
		StopPrice: price,
		Stop:      "entry",
	}, 10)
}
