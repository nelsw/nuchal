package trade

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/account"
	"nchl/pkg/nuchal"
	"nchl/pkg/product"
	"nchl/pkg/rate"
	"nchl/pkg/util"
	"time"
)

type orderType string

func (t orderType) String() string {
	return string(t)
}

type orderSide string

func (t orderSide) String() string {
	return string(t)
}

type orderStop string

func (t orderStop) String() string {
	return string(t)
}

const (
	market orderType = "market"
	limit  orderType = "limit"
	buy    orderSide = "buy"
	sell   orderSide = "sell"
	loss   orderStop = "loss"
	entry  orderStop = "entry"
)

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f) // todo - get increment units dynamically from cb api
}

func New() error {

	log.Info().Msg("creating new trades")

	if c, err := nuchal.NewConfig(); err != nil {
		log.Error().Err(err)
		return err
	} else {

		var wsDialer ws.Dialer
		if w, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
			log.Error().Err(err)
			return err
		} else {

			defer func(wsConn *ws.Conn) {
				if err := wsConn.Close(); err != nil {
					log.Error().Err(err)
				}
			}(w)

			exit := make(chan string)

			for _, p := range c.Postures {
				for _, u := range c.Users {
					go createUserTrades(w, u, p)
				}
			}

			for {
				select {
				case <-exit:
					return nil
				}
			}
		}
	}
}

func createUserTrades(wsConn *ws.Conn, u account.User, p product.Posture) {

	fmt.Println("creating trades")

	var then, that rate.Candlestick
	for {

		this, err := getRate(wsConn, p.ProductId())
		if err != nil {
			log.Error().Err(err).Msg("err getting rate for " + p.ProductId())
			then = rate.Candlestick{}
			that = rate.Candlestick{}
			continue
		}

		if rate.IsTweezer(this, then, that) {

			marketBuyOrder := &cb.Order{
				ProductID: p.ProductId(),
				Side:      buy.String(),
				Size:      p.Size,
				Type:      market.String(),
			}

			marketBuyOrderId, err := createOrder(u.GetClient(), marketBuyOrder)
			if err != nil && err.Error() == "Insufficient Funds" {
				log.Warn().Err(err).Msg("sleeping...")
				then = rate.Candlestick{}
				that = rate.Candlestick{}
				time.Sleep(time.Hour)
				continue
			} else if err != nil {
				log.Error().Err(err).Msg("error creating market buy order, exiting...")
				return
			}

			settledMarketBuyOrder, err := getOrder(u.GetClient(), *marketBuyOrderId)
			if err != nil {
				log.Error().Err(err).Msg("error getting settled market buy order, exiting...")
				return
			}

			price := util.Float64(settledMarketBuyOrder.ExecutedValue) / util.Float64(settledMarketBuyOrder.Size)
			price += price * util.Float64(p.Gain)

			stopGainOrder := &cb.Order{
				Price:     Price(price),
				ProductID: p.ProductId(),
				Side:      sell.String(),
				Size:      settledMarketBuyOrder.Size,
				Type:      limit.String(),
				StopPrice: Price(price),
				Stop:      entry.String(),
			}

			if _, err = createOrder(u.GetClient(), stopGainOrder); err != nil {
				log.Error().Err(err).Msg("error creating stop gain order, exiting...")
				return
			}
		}

		then = that
		that = this
	}
}

func getRate(wsConn *ws.Conn, productId string) (rate.Candlestick, error) {

	fmt.Println("building rate")

	end := time.Now().Add(time.Minute)

	var low, high, open, vol float64
	for {

		price, err := getPrice(wsConn, productId)
		if err != nil {
			return rate.Candlestick{}, err
		}

		vol++

		if low == 0 {
			low = price
			high = price
			open = price
		} else if high < price {
			high = price
		} else if low > price {
			low = price
		}

		if time.Now().After(end) {
			fmt.Println("built rate")
			return rate.Candlestick{
				time.Now().UnixNano(),
				productId,
				low,
				high,
				open,
				price,
				vol,
			}, nil
		}
	}
}

// GetPrice gets the latest ticker price for the given productId.
// Note that we omit Logging from this method to avoid blowing up the Logs.
func getPrice(wsConn *ws.Conn, productId string) (float64, error) {
	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productId}}},
	}); err != nil {
		panic(err)
	}
	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().Err(err)
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

func getOrder(client *cb.Client, orderId string, attempt ...int) (*cb.Order, error) {

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if order, err := client.GetOrder(orderId); err != nil {
		i++
		if i > 10 {
			return nil, err
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return getOrder(client, orderId, i)
	} else if !order.Settled {
		time.Sleep(1 * time.Second)
		return getOrder(client, orderId, 0)
	} else if order.Status == "pending" {
		time.Sleep(1 * time.Second)
		return getOrder(client, orderId, 0)
	} else {
		return &order, nil
	}
}

func createOrder(client *cb.Client, order *cb.Order, attempt ...int) (*string, error) {

	log.Info().Msg("creating order for " + order.ProductID)

	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}

	if r, err := client.CreateOrder(order); err != nil {
		log.Error().Err(err).Msg("error creating order for " + order.ProductID)
		i++
		if i > 10 {
			return nil, err
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return createOrder(client, order, i)
	} else {
		log.Info().Msg("created order for " + order.ProductID)
		return &r.ID, nil
	}
}
