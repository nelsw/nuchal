package pkg

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"math"
	"time"
)

func CreateTrades() {

	SetupWebsocket()

	start := time.Now().Round(time.Minute)
	stop := start.Add(time.Minute)

	var then, that Rate

	for  {

		this := buildRate(start, stop)

		if then == (Rate{}) {
			then = this
			continue
		}

		if that == (Rate{}) {
			that = this
			continue
		}

		if then.IsDown() && that.IsDown() && this.IsUp() {
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

func buildRate(from, to time.Time) Rate {

	rate := Rate{
		Unix:   from.Unix(),
		Volume: 0,
	}

	for {
		p := GetPrice()
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

		size := order.Size
		price := Float(order.Price)

		stopGain := Price(price + (price * target.Gain))
		if gainExit, err := createEntryOrder(size, stopGain); err != nil {
			fmt.Println("error creating gain exit order",err)
		} else  {
			fmt.Println("successfully created gain exit order", gainExit)
		}

		stopLoss := Price(price - (price * target.Loss))
		if lossExit, err := createExitOrder(size, stopLoss); err != nil {
			fmt.Println("error creating loss exit order",err)
		} else  {
			fmt.Println("successfully created loss exit order", lossExit)
		}
	}
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
		Type:      "entry",
	}, 0)
}

func createExitOrder(size, price string) (*cb.Order, error) {
	return CreateOrder(&cb.Order{
		Price:     price,
		ProductID: target.ProductId,
		Side:      "sell",
		Size:      size,
		StopPrice: price,
		Type:      "loss",
	}, 0)
}


