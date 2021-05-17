package trade

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/model/account"
	"nchl/pkg/model/config"
	"nchl/pkg/model/crypto"
	"nchl/pkg/model/statistic"
	"nchl/pkg/util"
	"time"
)

func New() error {

	log.Info().Msg("creating new trades")

	c, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err).Msg("error creating nuchal config")
		return err
	}

	log.Info().Msg("trading indefinitely")
	return util.DoIndefinitely(func() {
		for _, p := range c.Postures {
			go trade(c.Group, p)
		}
	})
}

func trade(g *account.Group, p crypto.Posture) {

	log.Info().
		Str("productId", p.ProductId()).
		Msg("creating trades")

	var then, that statistic.Candlestick
	for {
		if this, err := getRate(p.ProductId()); err != nil {
			then = statistic.Candlestick{}
			that = statistic.Candlestick{}
			// logging in getRate
		} else if !statistic.IsTweezer(then, that, *this, p.DeltaFloat()) { // logging in IsTweezer
			then = that
			that = *this
		} else {
			for _, u := range g.Users {
				go buy(u, p)
			}
			then = statistic.Candlestick{}
			that = statistic.Candlestick{}
		}
	}
}

func buy(u account.User, p crypto.Posture) {

	log.Info().
		Str("user", u.Name).
		Str("productId", p.ProductId()).
		Msg("... buying")

	if order, err := createOrder(u, p.MarketEntryOrder()); err == nil {

		log.Info().
			Str("user", u.Name).
			Str("productId", p.ProductId()).
			Str("orderId", order.ID).
			Msg("created order")

		entryPrice := util.Float64(order.ExecutedValue) / util.Float64(order.Size)
		exitPrice := entryPrice + (entryPrice * p.GainFloat())

		sell(u, exitPrice, order.Size, p)

	} else if util.IsInsufficientFunds(err) {

		log.Warn().
			Err(err).
			Str("user", u.Name).
			Str("productId", p.ProductId()).
			Msg("Insufficient funds ... sleeping ...")

		time.Sleep(time.Hour) // todo check if has funds and if more sleep required

	} else {
		log.Error().
			Err(err).
			Str("user", u.Name).
			Str("productId", p.ProductId()).
			Msg("error buying")
	}
}

func sell(u account.User, exitPrice float64, size string, p crypto.Posture) {

	log.Info().
		Str("user", u.Name).
		Str("productId", p.ProductId()).
		Float64("exitPrice", exitPrice).
		Str("size", size).
		Msg("... selling")

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {

		log.Error().
			Err(err).
			Str("user", u.Name).
			Str("productId", p.ProductId()).
			Msg("error opening websocket connection")

		if _, err := createOrder(u, p.StopEntryOrder(exitPrice, size)); err != nil {
			log.Error().
				Err(err).
				Str("user", u.Name).
				Str("productId", p.ProductId()).
				Msg("error while creating entry order")
		}

		return
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Str("user", u.Name).
				Str("productId", p.ProductId()).
				Msg("error while close websocket connection")
		}
	}(wsConn)

	for {

		lastPrice, err := getPrice(wsConn, p.ProductId())
		if err != nil {
			log.Error().
				Err(err).
				Msg("error while getting sell price")
			if _, err := createOrder(u, p.StopEntryOrder(exitPrice, size)); err != nil {
				log.Error().
					Err(err).
					Msg("error while creating stop entry order")
			}
			return
		}

		if *lastPrice < exitPrice {
			continue
		}

		var stopLossOrder *cb.Order
		if stopLossOrder, err = createOrder(u, p.StopLossOrder(*lastPrice, size)); err != nil {

			log.Error().
				Err(err).
				Msg("error while creating stop loss order")
			return
		}

		for {
			if r, err := getRate(p.ProductId()); err != nil {
				log.Error().
					Err(err).
					Msg("error while getting rate during stop loss climb")
				return
			} else if r.Low <= exitPrice {
				return // stop loss executed
			} else if r.Close > exitPrice {

				log.Info().
					Msg("found better price!")

				if err := cancelOrder(u, stopLossOrder.ID); err != nil {
					log.Error().
						Err(err).
						Msg("error while canceling order")
					return
				}

				exitPrice = r.Close
				if stopLossOrder, err = createOrder(u, p.StopLossOrder(exitPrice, size)); err != nil {
					log.Error().
						Err(err).
						Msg("error while creating stop loss order")
					return
				}
			}
		}
	}
}

func getRate(productId string) (*statistic.Candlestick, error) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("productId", productId).
			Msg("error while opening websocket connection")
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Str("productId", productId).
				Msg("error closing websocket connection")
		}
	}(wsConn)

	end := time.Now().Add(time.Minute)

	var low, high, open, vol float64
	for {

		price, err := getPrice(wsConn, productId)
		if err != nil {
			log.Error().
				Err(err).
				Str("productId", productId).
				Msg("error getting price")
			return nil, err
		}

		vol++

		if low == 0 {
			low = *price
			high = *price
			open = *price
		} else if high < *price {
			high = *price
		} else if low > *price {
			low = *price
		}

		if time.Now().After(end) {

			log.Info().
				Str("productId", productId).
				Msg("...built rate")

			return &statistic.Candlestick{
				time.Now().UnixNano(),
				productId,
				low,
				high,
				open,
				*price,
				vol,
			}, nil
		}
	}
}

// getPrice gets the latest ticker price for the given productId. This method does not perform logging as it is executed
// thousands of times per second.
func getPrice(wsConn *ws.Conn, productId string) (*float64, error) {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productId}}},
	}); err != nil {
		log.Error().
			Err(err).
			Str("productId", productId).
			Msg("error writing message to websocket")
		return nil, err
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().
				Err(err).
				Str("productId", productId).
				Msg("error reading from websocket")
			return nil, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := errors.New(fmt.Sprintf("message type != ticker, %v", receivedMessage))
		log.Error().
			Err(err).
			Str("productId", productId).
			Msg("error getting ticker message from websocket")
		return nil, err
	}

	f := util.Float64(receivedMessage.Price)
	return &f, nil
}

// createOrder creates an order on Coinbase and returns the order once it is no longer pending and has settled.
// Given that there are many different types of orders that can be created in many different scenarios, it is the
// responsibility of the method calling this function to perform logging.
func createOrder(u account.User, order *cb.Order, attempt ...int) (*cb.Order, error) {

	r, err := u.GetClient().CreateOrder(order)
	if err == nil {
		return getOrder(u, r.ID)
	}

	i := util.FirstIntOrZero(attempt)
	if util.IsInsufficientFunds(err) || i > 10 {
		return nil, err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return createOrder(u, order, i)
}

// getOrder is a recursive function that returns an order equal to the given id once it is settled and not pending.
// This function also performs extensive logging given its variable and seriously critical nature.
func getOrder(u account.User, id string, attempt ...int) (*cb.Order, error) {

	log.Info().Str("user", u.Name).Str("orderId", id).Msg("get order")

	order, err := u.GetClient().GetOrder(id)

	if err != nil {

		i := util.FirstIntOrZero(attempt)

		log.Error().
			Err(err).
			Str("user", u.Name).
			Str("orderId", id).
			Int("attempt", i).
			Msg("error getting order")

		if i > 10 {
			return nil, err
		}

		i++
		time.Sleep(time.Duration(i*3) * time.Second)
		return getOrder(u, id, i)
	}

	if !order.Settled || order.Status == "pending" {

		log.Warn().
			Str("user", u.Name).
			Str("product", order.ProductID).
			Str("orderId", id).
			Str("side", order.Side).
			Str("type", order.Type).
			Msg("got order, but it's pending or unsettled")

		time.Sleep(1 * time.Second)
		return getOrder(u, id, 0)
	}

	log.Info().
		Str("user", u.Name).
		Str("product", order.ProductID).
		Str("orderId", id).
		Str("side", order.Side).
		Str("type", order.Type).
		Msg("got order")
	return &order, nil
}

// cancelOrder is a recursive function that cancels an order equal to the given id.
func cancelOrder(u account.User, id string, attempt ...int) error {

	log.Info().
		Str("user", u.Name).
		Str("user", u.Name).
		Str("orderId", id).
		Msg("canceling order")

	err := u.GetClient().CancelOrder(id)
	if err == nil {
		log.Info().
			Str("user", u.Name).
			Str("orderId", id).
			Msg("canceled order")
		return nil
	}

	i := util.FirstIntOrZero(attempt)

	log.Error().
		Err(err).
		Str("user", u.Name).
		Str("orderId", id).
		Int("attempt", i).
		Msg("error canceling order")

	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return cancelOrder(u, id, i)
}
