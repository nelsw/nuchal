package pkg

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

// GetPrice gets the latest ticker price for the given productId.
// Note that we omit logging from this method to avoid blowing up the logs.
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
	return float(receivedMessage.Price)
}

func GetOrder(username, productId, id string, attempt ...int) cb.Order {
	log(username, productId, "find settled order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if order, err := getClient(username).GetOrder(id); err != nil {
		log(username, productId, "error finding settled order", err)
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetOrder(username, productId, id, i)
	} else if !order.Settled {
		log(username, productId, "found unsettled order")
		time.Sleep(1 * time.Second)
		return GetOrder(username, productId, id, 0)
	} else if order.Status == "pending" {
		log(username, productId, "found pending order")
		time.Sleep(1 * time.Second)
		return GetOrder(username, productId, id, 0)
	} else {
		log(username, productId, "found settled order")
		return order
	}
}

func FindFillsByOrderId(username, orderId string) []cb.Fill {
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

func CreateMarketBuyOrder(username, productId, size string) (float64, string) {
	log(username, productId, "creating market buy order")
	if order, err := createOrder(username, &cb.Order{
		ProductID: productId,
		Side:      "buy",
		Size:      size,
		Type:      "market",
	}); err != nil {
		log(username, productId, "error creating market buy order", err)
		handleError(err)
		return CreateMarketBuyOrder(username, productId, size)
	} else {
		log(username, productId, "created market buy order", order)
		settledOrder := GetOrder(username, productId, order.ID)
		return float(settledOrder.ExecutedValue) / float(settledOrder.Size), settledOrder.Size
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
	log(username, productId, "creating stop loss order")
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
		log(username, productId, "error creating stop loss order", err)
		i++
		if i > 10 {
			handleError(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return CreateStopLossOrder(username, productId, size, price, i)
	} else {
		log(username, productId, "created stop loss order", stopLossOrder)
		return stopLossOrder.ID
	}
}

func createOrder(username string, order *cb.Order, attempt ...int) (*cb.Order, error) {
	log(username, order.ProductID, "creating order", order)
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if r, err := getClient(username).CreateOrder(order); err != nil {
		log(username, order.ProductID, "error creating order", err)
		i++
		if i > 10 {
			return nil, err
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return createOrder(username, order, i)
	} else {
		log(username, order.ProductID, "order created", r)
		return &r, nil
	}
}

func CancelOrder(username, productId, orderId string, attempt ...int) {
	log(username, productId, "cancelling order")
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if err := getClient(username).CancelOrder(orderId); err != nil {
		log(username, productId, "error canceling order", err)
		i++
		if i > 10 {
			handleError(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		CancelOrder(username, productId, orderId, i)
	} else {
		log(username, productId, "cancelled order")
	}
}

func CreateHistoricRates(username, productId string, from time.Time) []Rate {
	log(username, productId, "getting new rates")
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
			log(username, productId, "got new rates")
			return rates
		}
		from = to
		to = to.Add(time.Hour * 4)
	}
}

func GetHistoricRates(username, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {
	log(username, productId, "getting historic rates")
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
		log(username, productId, "got historic rates")
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
