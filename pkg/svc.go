package pkg

import (
	"encoding/json"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"strconv"
	"time"
)

// GetTickerPrice gets the latest ticker price for the given productId.
// Note that we omit logging from this method to avoid blowing up the logs.
func GetTickerPrice(wsConn *ws.Conn, productId string) float64 {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productId}}},
	}); err != nil {
		panic(err)
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			panic(err)
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		panic(fmt.Sprintf("message type != ticker, %v", receivedMessage))
	}

	return float(receivedMessage.Price)
}

func FindSettledOrder(username, id string, attempt ...int) cb.Order {

	fmt.Println("find settled order")

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if order, err := getOrder(username, id); err != nil {

		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		return FindSettledOrder(username, id, i)

	} else if !order.Settled {

		fmt.Println("found unsettled order")
		time.Sleep(1 * time.Second)
		return FindSettledOrder(username, id, 0)

	} else {
		fmt.Println("found settled order")
		return order
	}
}

func getOrder(username, id string) (cb.Order, error) {
	fmt.Println("finding order")
	order, err := getClient(username).GetOrder(id)
	if err != nil {
		fmt.Println("error finding order", err)
	} else {
		fmt.Println("found order", prettyJson(order))
	}
	return order, err
}

func CreateMarketBuyOrder(username, productId, size string) (float64, string) {
	fmt.Println("creating market buy order")
	if order, err := createOrder(username, &cb.Order{
		ProductID: productId,
		Side:      "buy",
		Size:      size,
		Type:      "market",
	}); err != nil {
		fmt.Println("error creating market buy order", err)
		panic(err)
	} else {
		fmt.Println("created market buy order", prettyJson(order))
		settledOrder := FindSettledOrder(username, order.ID)
		return float(settledOrder.ExecutedValue) / float(settledOrder.Size), settledOrder.Size
	}
}

func CreateMarketSellOrder(username, productId, size string) {
	fmt.Println("creating market sell order")
	if order, err := createOrder(username, &cb.Order{
		ProductID: productId,
		Side:      "sell",
		Size:      size,
		Type:      "market",
	}); err != nil {
		panic(err)
	} else {
		fmt.Println("created market sell order", prettyJson(order))
	}
}

func CreateEntryOrder(username, productId, size string, price float64) {
	fmt.Println("creating entry order")
	_, err := createOrder(username, &cb.Order{
		Price:     formatPrice(price),
		ProductID: productId,
		Side:      "sell",
		Size:      size,
		StopPrice: formatPrice(price),
		Stop:      "entry",
	})
	if err != nil {
		fmt.Println("error creating entry order")
	} else {
		fmt.Println("created entry order")
	}
}

func CreateStopLossOrder(username, productId, size string, price float64, attempt ...int) string {

	fmt.Println("creating stop loss order")

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	stopLossOrder, err := createOrder(username, &cb.Order{
		Price:     formatPrice(price),
		ProductID: productId,
		Side:      "sell",
		Size:      size,
		StopPrice: formatPrice(price),
		Stop:      "loss",
	})
	if err != nil {
		fmt.Println("error creating stop loss order")
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(5 * time.Second)
		return CreateStopLossOrder(username, productId, size, price, i)
	} else {
		fmt.Println("created stop loss order", prettyJson(stopLossOrder))
		return stopLossOrder.ID
	}
}

func createOrder(username string, order *cb.Order, attempt ...int) (*cb.Order, error) {

	fmt.Println("creating order", prettyJson(order))

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if r, err := getClient(username).CreateOrder(order); err != nil {

		i++
		if i > 10 {
			fmt.Println("error creating order", err)
			return nil, err
		}
		time.Sleep(5 * time.Second)
		return createOrder(username, order, i)

	} else {
		fmt.Println("order created", prettyJson(r))
		return &r, nil
	}
}

func CancelOrder(username, orderId string, attempt ...int) {
	fmt.Println("cancelling order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	err := getClient(username).CancelOrder(orderId)
	if err != nil {
		i++
		if i > 10 {
			fmt.Println("error canceling order", err)
			panic(err)
		}
		time.Sleep(5 * time.Second)
		CancelOrder(username, orderId, i)
		fmt.Println("cancelled order")
	}
}

func NewRates(username, productId string, from time.Time) []Rate {

	fmt.Println("getting new rates")

	to := from.Add(time.Hour * 4)
	var rates []Rate
	for {
		for _, rate := range GetHistoricRates(username, productId, from, to) {
			rates = append(rates, Rate{
				rate.Time.UnixNano(),
				productId,
				rate.Low,
				rate.High,
				rate.Open,
				rate.Close,
				rate.Volume,
			})
		}
		if to.After(time.Now()) {
			fmt.Println("got new rates")
			return rates
		}
		from = to
		to = to.Add(time.Hour * 4)
	}
}

func GetHistoricRates(username, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {

	fmt.Println("getting historic rates")

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if rates, err := getClient(username).GetHistoricRates(productId, cb.GetHistoricRatesParams{
		from,
		to,
		60,
	}); err != nil {
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Second * 5)
		return GetHistoricRates(username, productId, from, to, i)
	} else {
		fmt.Println("got historic rates")
		return rates
	}
}

func getClient(username string) *cb.Client {
	key, pass, secret := GetUserConfig(username)
	return &cb.Client{
		"https://api.pro.coinbase.com",
		*key,
		*pass,
		*secret,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}

func float(s string) float64 {
	if f, err := strconv.ParseFloat(s, 32); err != nil {
		panic(err)
	} else {
		return f
	}
}

func prettyJson(v interface{}) string {
	b, _ := json.MarshalIndent(&v, "", "  ")
	return string(b)
}

func formatPrice(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
