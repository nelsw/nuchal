package pkg

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nchl/pkg/util"
	"time"
)

const (
	granularity   = 60
	ticker        = "ticker"
	subscribe     = "subscribe"
	subscriptions = "subscriptions"
	BaseUrl       = "https://api.pro.coinbase.com"
	WsUrl         = "wss://ws-feed.pro.coinbase.com"
)

var (
	client = cb.NewClient()
)

func SetupClientConfig() {
	fmt.Println("setting up client config for", user.Name)
	client.UpdateConfig(&cb.ClientConfig{
		BaseUrl,
		user.Key,
		user.Passphrase,
		user.Secret,
	})
	fmt.Println("setup client config for", user.Name)
	fmt.Println()
}

func GetLastPrice(productId string) string {
	ticker, err := client.GetTicker(productId)
	if err != nil {
		panic(err)
	}
	return ticker.Price
}

func GetPrice(wsConn *ws.Conn) float64 {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     subscribe,
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

	return util.Float(receivedMessage.Price)
}

func GetOrders() []cb.Order {

	cursor := client.ListOrders(cb.ListOrdersParams{
		ProductID: target.ProductId,
	})
	var chunks, allChunks []cb.Order
	for cursor.HasMore {
		if err := cursor.NextPage(&chunks); err != nil {
			panic(err)
		}
		for _, o := range chunks {
			allChunks = append(allChunks, o)
		}
	}
	return allChunks
}

func GetFills() []cb.Fill {
	cursor := client.ListFills(cb.ListFillsParams{
		OrderID:    "",
		ProductID:  target.ProductId,
		Pagination: cb.PaginationParams{},
	})

	var fills []cb.Fill
	var allFills []cb.Fill

	for cursor.HasMore {
		if err := cursor.NextPage(&fills); err != nil {
			panic(err)
		}
		for _, o := range fills {
			allFills = append(allFills, o)
		}
	}

	return allFills
}

func GetOrder(id string) (cb.Order, error) {
	fmt.Println("getting order", id)
	order, err := client.GetOrder(id)
	if err != nil {
		fmt.Println("error getting order", id)
	} else {
		fmt.Println("got order", util.Print(order))
	}
	return order, err
}

func CreateOrder(order *cb.Order, attempt int) (*cb.Order, error) {
	fmt.Println("creating order", util.Print(order))
	if r, err := client.CreateOrder(order); err == nil {
		fmt.Println("order created", util.Print(r))
		fmt.Println()
		return &r, nil
	} else {
		fmt.Println("failed to create", util.Print(order))
		attempt++
		if attempt < 10 {
			return CreateOrder(order, attempt)
		} else {
			return nil, errors.New(fmt.Sprintf("entry failed %v ", err))
		}
	}
}

func BuildRates() []Rate {

	fmt.Println("building rates for", target.ProductId)

	var rates []Rate
	for {

		for _, rate := range getHistoricRates(0) {
			rates = append(rates, Rate{
				rate.Time.UnixNano(),
				target.ProductId,
				rate.Low,
				rate.High,
				rate.Open,
				rate.Close,
				rate.Volume,
			})
		}

		if to.After(time.Now()) {
			fmt.Println("built rates")
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
		fmt.Println(err)
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

func printAccountInfo() {
	if accounts, err := client.GetAccounts(); err != nil {
		panic(err)
	} else {
		for _, account := range accounts {
			fmt.Println(util.Print(account))
		}
	}
}

func printPositions() {
	if accounts, err := client.GetAccounts(); err != nil {
		panic(err)
	} else {
		total := 0.0
		for _, account := range accounts {
			if account.Currency == "USD" {
				total += util.Float(account.Balance)
				continue
			}
			if util.Float(account.Balance) > 0 {
				fmt.Println(util.Print(account))
				total += util.Float(GetLastPrice(account.Currency+toUSD)) * util.Float(account.Balance)
			}
		}
		fmt.Println(total)
	}
}

func printCashBalance() {
	if accounts, err := client.GetAccounts(); err != nil {
		panic(err)
	} else {
		for _, account := range accounts {
			if account.Currency == "USD" {
				fmt.Println(util.Print(account))
			}
		}
	}
}
