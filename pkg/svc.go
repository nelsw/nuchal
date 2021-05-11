package pkg

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nchl/pkg/util"
	"net/http"
	"time"
)

// GetPrice gets the latest ticker price for the given productId.
// Note that we omit Logging from this method to avoid blowing up the Logs.
func GetPrice(wsConn *ws.Conn, productId string) float64 {
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
	return util.Float(receivedMessage.Price)
}

func GetOrderSizeAndPrice(username, productId, id string) (string, float64) {
	order := GetOrder(username, productId, id)
	size := order.Size
	price := util.Float(order.ExecutedValue) / util.Float(size)
	return size, price
}

func GetOrder(username, productId, id string, attempt ...int) cb.Order {
	util.Log(username, productId, "find settled order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if order, err := getClient(username).GetOrder(id); err != nil {
		util.Log(username, productId, "error finding settled order", err)
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetOrder(username, productId, id, i)
	} else if !order.Settled {
		util.Log(username, productId, "found unsettled order")
		time.Sleep(1 * time.Second)
		return GetOrder(username, productId, id, 0)
	} else if order.Status == "pending" {
		util.Log(username, productId, "found pending order")
		time.Sleep(1 * time.Second)
		return GetOrder(username, productId, id, 0)
	} else {
		util.Log(username, productId, "found settled order")
		return order
	}
}

func GetFills(username, orderId string) []cb.Fill {
	cursor := getClient(username).ListFills(cb.ListFillsParams{
		OrderID: orderId,
	})
	var newChunks, allChunks []cb.Fill
	for cursor.HasMore {
		if err := cursor.NextPage(&newChunks); err != nil {
			handleError(err)
		}
		for _, chunk := range newChunks {
			allChunks = append(allChunks, chunk)
		}
	}
	return allChunks
}

func GetLedgers(username, accountId string) []cb.LedgerEntry {
	cursor := getClient(username).ListAccountLedger(accountId)
	var newChunks, allChunks []cb.LedgerEntry
	for cursor.HasMore {
		if err := cursor.NextPage(&newChunks); err != nil {
			handleError(err)
		}
		for _, chunk := range newChunks {
			allChunks = append(allChunks, chunk)
		}
	}
	return allChunks
}

func GetAccounts(username string) []cb.Account {
	if accounts, err := getClient(username).GetAccounts(); err != nil {
		handleError(err)
		return GetAccounts(username)
	} else {
		return accounts
	}
}

func CreateOrder(username string, order *cb.Order, attempt ...int) (*string, error) {
	util.Log(username, order.ProductID, "creating order", order)
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if r, err := getClient(username).CreateOrder(order); err != nil {
		util.Log(username, order.ProductID, "error creating order", err)
		i++
		if i > 10 {
			return nil, err
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return CreateOrder(username, order, i)
	} else {
		util.Log(username, order.ProductID, "order created", r)
		return &r.ID, nil
	}
}

func CancelOrder(username, productId, orderId string, attempt ...int) {
	util.Log(username, productId, "cancelling order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if err := getClient(username).CancelOrder(orderId); err != nil {
		util.Log(username, productId, "error canceling order", err)
		i++
		if i > 10 {
			handleError(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		CancelOrder(username, productId, orderId, i)
	} else {
		util.Log(username, productId, "cancelled order")
	}
}

func CreateHistoricRates(username, productId string, from time.Time) []Rate {
	util.Log(username, productId, "getting new rates")
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
			util.Log(username, productId, "got new rates")
			return rates
		}
		from = to
		to = to.Add(time.Hour * 4)
	}
}

func GetHistoricRates(username, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {
	util.Log(username, productId, "getting historic rates")
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
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetHistoricRates(username, productId, from, to, i)
	} else {
		util.Log(username, productId, "got historic rates")
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

func getClient(username string) *cb.Client {
	key, pass, secret := GetUserConfig(username)
	return &cb.Client{
		"https://api.pro.coinbase.com",
		*secret,
		*key,
		*pass,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}
