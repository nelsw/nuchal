package pkg

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

const (
	granularity = 60
	ticker = "ticker"
	subscribe = "subscribe"
	subscriptions = "subscriptions"
	BaseUrl = "https://api.pro.worker.com"
	WsUrl   = "wss://ws-feed.pro.worker.com"
)

var (
	client = cb.NewClient()
	wsDialer ws.Dialer
	wsConn *ws.Conn
)

func SetupClientConfig() {
	fmt.Println("setting up client config for user", user)
	client.UpdateConfig(&cb.ClientConfig{
		BaseUrl,
		user.Key,
		user.Passphrase,
		user.Secret,
	})
	fmt.Println("setup client config for user", user)
	fmt.Println()
}

func SetupWebsocket() {
	fmt.Println("setting up websocket for user", user)
	conn, _, err := wsDialer.Dial(WsUrl, nil)
	if err != nil {
		panic(err)
	} else {
		wsConn = conn
		defer func(wsConn *ws.Conn) {
			if err = wsConn.Close(); err != nil {
				fmt.Println(err)
			}
		}(wsConn)
	}
	fmt.Println("set up websocket for user", user)
	fmt.Println()
}

func GetPrice() float64 {

	fmt.Println("getting price")

	if err := wsConn.WriteJSON(&cb.Message{
		Type: subscribe,
		Channels: []cb.MessageChannel{{ticker, []string{target.ProductId}}},
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

	price := Float(receivedMessage.Price)

	fmt.Println("got price", price)
	fmt.Println()

	return price
}

func CreateOrder(order *cb.Order, attempt int) (*cb.Order, error) {
	fmt.Println("creating order", order)
	if r, err := client.CreateOrder(order); err == nil {
		fmt.Println("order created", r)
		fmt.Println()
		return &r, nil
	} else {
		fmt.Println("failed to create", order)
		attempt++
		if attempt < 10 {
			return CreateOrder(order, attempt)
		} else {
			return nil, errors.New(fmt.Sprintf("entry failed %v ", err))
		}
	}
}

func BuildRates() []Rate {

	fmt.Println("building rates for", target)

	var rates []Rate
	for {

		for _, rate := range getHistoricRates(0) {
			fmt.Println("found rate", rate)
			rates = append(rates, Rate{
				rate.Time.UnixNano(),
				target.ProductId,
				rate.Low,
				rate.High,
				rate.Open,
				rate.Close,
				rate.Volume,
			})
			fmt.Println("appended rate")
		}

		if to.After(time.Now()) {
			fmt.Println("built rates for", target)
			fmt.Println()
			return rates
		}

		fmt.Println("looping for more rates")
		from = to
		to = to.Add(time.Hour * 4)
	}

}

func getHistoricRates(attempt int) []cb.HistoricRate {
	fmt.Println("getting historic rates, attempt", attempt)
	if rates, err := client.GetHistoricRates(target.ProductId, cb.GetHistoricRatesParams{
		from,
		to,
		granularity,
	}); err != nil {
		attempt++
		if attempt < 100 {
			return getHistoricRates(attempt)
		}
		panic(err)
	} else {
		fmt.Println("returning historic rates, attempt", attempt)
		fmt.Println()
		return rates
	}
}