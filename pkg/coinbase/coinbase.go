package coinbase

import (
	"encoding/json"
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"log"
	"nchl/pkg/conf"
	"nchl/pkg/rate"
	"nchl/pkg/util"
	"net/http"
	"time"
)

// GetPrice gets the latest ticker price for the given productId.
// Note that we omit Logging from this method to avoid blowing up the Logs.
func GetPrice(wsConn *ws.Conn, productId string) (float64, error) {
	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productId}}},
	}); err != nil {
		panic(err)
	}
	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Printf("error: %v", err)
			return 0, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}
	if receivedMessage.Type != "ticker" {
		return 0, errors.New(fmt.Sprintf("message type != ticker, %v", receivedMessage))
	}
	return util.Float64(receivedMessage.Price), nil
}

// GetTicker gets the latest ticker price for the given productId.
func GetTicker(user conf.User, productId string) string {
	t, err := getClient(user).GetTicker(productId)
	if err != nil {
		panic(err)
	}
	return t.Price
}

func GetOrderSizeAndPrice(user conf.User, productId, id string) (string, float64) {
	order := GetOrder(user, productId, id)
	size := order.Size
	price := util.Float64(order.ExecutedValue) / util.Float64(size)
	return size, price
}

func GetOrder(user conf.User, productId, id string, attempt ...int) cb.Order {
	Log(user.Name, productId, "find settled order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if order, err := getClient(user).GetOrder(id); err != nil {
		Log(user.Name, productId, "error finding settled order", err)
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetOrder(user, productId, id, i)
	} else if !order.Settled {
		Log(user.Name, productId, "found unsettled order")
		time.Sleep(1 * time.Second)
		return GetOrder(user, productId, id, 0)
	} else if order.Status == "pending" {
		Log(user.Name, productId, "found pending order")
		time.Sleep(1 * time.Second)
		return GetOrder(user, productId, id, 0)
	} else {
		Log(user.Name, productId, "found settled order")
		return order
	}
}

func GetAccounts(user conf.User) []cb.Account {
	if accounts, err := getClient(user).GetAccounts(); err != nil {
		handleError(err)
		return GetAccounts(user)
	} else {
		return accounts
	}
}

func CreateOrder(user conf.User, order *cb.Order, attempt ...int) *string {
	Log(user.Name, order.ProductID, "creating order", order)
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if r, err := getClient(user).CreateOrder(order); err != nil {
		Log(user.Name, order.ProductID, "error creating order", err)
		i++
		if i > 10 {
			return nil
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return CreateOrder(user, order, i)
	} else {
		Log(user.Name, order.ProductID, "order created", r)
		return &r.ID
	}
}

func CreateHistoricRates(user conf.User, productId string, from time.Time) []rate.Candlestick {
	Log(user.Name, productId, "getting new rates")
	to := from.Add(time.Hour * 4)
	var rates []rate.Candlestick
	for {
		for _, r := range GetHistoricRates(user, productId, from, to) {
			rates = append(rates, rate.Candlestick{
				r.Time.UnixNano(),
				productId,
				r.Low,
				r.High,
				r.Open,
				r.Close,
				r.Volume,
			})
		}
		if to.After(time.Now()) {
			Log(user.Name, productId, "got new rates")
			return rates
		}
		from = to
		to = to.Add(time.Hour * 4)
	}
}

func GetHistoricRates(user conf.User, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {
	Log(user.Name, productId, "getting historic rates")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if rates, err := getClient(user).GetHistoricRates(productId, cb.GetHistoricRatesParams{
		from,
		to,
		60,
	}); err != nil {
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetHistoricRates(user, productId, from, to, i)
	} else {
		Log(user.Name, productId, "got historic rates")
		return rates
	}
}

func handleError(err error) {
	switch err.Error() {
	case "Private rate limit exceeded":
		time.Sleep(time.Duration(5) * time.Second)
	case "Insufficient funds":
		time.Sleep(time.Duration(1) * time.Minute)
	default:
		panic(err)
	}
}

func getClient(user conf.User) *cb.Client {
	return &cb.Client{
		"https://api.pro.coinbase.com",
		user.Secret,
		user.Key,
		user.Passphrase,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}

func Log(username, productId, message string, v ...interface{}) {
	if v == nil || len(v) == 0 {
		fmt.Println(fmt.Sprintf("%s - %s - %s", username, productId, message))
		return
	}
	b, _ := json.MarshalIndent(&v, "", "  ")
	fmt.Println(fmt.Sprintf("%s - %s - %s [%s]", username, productId, message, string(b)))
}
