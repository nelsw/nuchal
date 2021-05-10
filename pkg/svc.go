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

const (
	buy           = "buy"
	market        = "market"
	sell          = "sell"
	entry         = "entry"
	granularity   = 60
	ticker        = "ticker"
	subscribe     = "subscribe"
	subscriptions = "subscriptions"
	BaseUrl       = "https://api.pro.coinbase.com"
)

func GetTickerPrice(wsConn *ws.Conn, productId string) float64 {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     subscribe,
		Channels: []cb.MessageChannel{{ticker, []string{productId}}},
	}); err != nil {
		panic(err)
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			panic(err)
		}
		if receivedMessage.Type != subscriptions {
			break
		}
	}

	if receivedMessage.Type != ticker {
		panic(fmt.Sprintf("message type != ticker, %v", receivedMessage))
	}

	return float(receivedMessage.Price)
}

func GetOrder(username, id string, attempt ...int) cb.Order {

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if order, err := getOrder(username, id); err != nil {

		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(5 * time.Second)
		return GetOrder(username, id, i)

	} else if !order.Settled {
		return GetOrder(username, id, 0)
	} else {
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

func CreateMarketOrder(username, productId, size string) (float64, string) {
	order, err := createOrder(username, &cb.Order{
		ProductID: productId,
		Side:      buy,
		Size:      size,
		Type:      market,
	})
	if err != nil {
		panic(err)
	} else {
		settledOrder := GetOrder(username, order.ID)
		return float(settledOrder.ExecutedValue) / float(settledOrder.Size), settledOrder.Size
	}
}

func CreateEntryOrder(username, productId, size string, price float64) {
	_, _ = createOrder(username, &cb.Order{
		Price:     fmt.Sprintf("%.3f", price),
		ProductID: productId,
		Side:      sell,
		Size:      size,
		StopPrice: fmt.Sprintf("%.3f", price),
		Stop:      entry,
	})
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

func NewRates(username, productId string, from time.Time) []Rate {

	to := time.Now()
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
			return rates
		}
		from = to
		to = to.Add(time.Hour * 4)
	}
}

func GetHistoricRates(username, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if rates, err := getClient(username).GetHistoricRates(productId, cb.GetHistoricRatesParams{
		from,
		to,
		granularity,
	}); err != nil {
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Second * 5)
		return GetHistoricRates(username, productId, from, to, i)
	} else {
		return rates
	}
}

func getClient(username string) *cb.Client {
	key, pass, secret := GetUserConfig(username)
	return &cb.Client{
		BaseUrl,
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
