/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package trade

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"time"
)

func NewHolds(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	positions, err := ses.GetTradingPositions()
	if err != nil {
		return err
	}

	for _, position := range *positions {

		log.Info().Str(util.Currency, position.ProductId()).Msg("hold")
		for _, trade := range position.GetActiveTrades() {

			price := position.GoalPrice(trade.Price())

			size := trade.Fill.Size
			hold := position.NewLimitSellEntryOrder(price, size)

			if _, err := ses.GetClient().CreateOrder(hold); err != nil {
				return err
			}
		}
		log.Info().Str(util.Currency, position.ProductId()).Msg("held")

	}

	return nil
}

func NewSells(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	positions, err := ses.GetActivePositions()
	if err != nil {
		return err
	}

	exit := make(chan error)
	go func() {
		for _, position := range *positions {
			log.Info().Str(util.Currency, position.ProductId()).Msg("sell")
			for _, trade := range position.GetActiveTrades() {

				log.Info().Str("#", position.ProductId()).Msg("sell")

				goal := position.GoalPrice(trade.Price())
				size := trade.Fill.Size

				_, err := sell(ses, goal, size, position.Product)
				if err != nil {
					exit <- err
					return
				}

				log.Info().Str("#", position.ProductId()).Msg("sold")
			}
			log.Info().Str("#", position.ProductId()).Msg("sold")
		}
	}()
	for {
		select {
		case err := <-exit:
			return err
		}
	}
}

func NewExits(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	positions, err := ses.GetActivePositions()
	if err != nil {
		return err
	}

	for _, position := range *positions {

		log.Info().Str("#", position.ProductId()).Msg("exit")

		for _, trade := range position.GetActiveTrades() {

			size := trade.Fill.Size
			hold := position.NewMarketSellOrder(size)

			_, err := ses.GetClient().CreateOrder(hold)
			if err != nil {
				return err
			}
		}
		log.Info().Str("#", position.ProductId()).Msg("exited")
	}
	return nil
}

func New(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}
	log.Info().Msg("trading")

	if util.IsEnvVarTrue("TEST") {
		return nil
	}

	return util.DoIndefinitely(func() {
		for _, p := range ses.Products {
			go trade(ses, p)
		}
	})
}

func trade(user *config.Session, p cbp.Product) {

	log.Info().Str("#", p.ID).Msg("trading")

	var then, that cbp.Rate
	for {
		if this, err := getRate(p.ID); err != nil {
			then = cbp.Rate{}
			that = cbp.Rate{}
		} else if !p.MatchesTweezerBottomPattern(then, that, *this) {
			then = that
			that = *this
		} else {
			go buy(user, p)
			then = cbp.Rate{}
			that = cbp.Rate{}
		}
	}
}

func buy(u *config.Session, p cbp.Product) {

	log.Info().
		Str("#", p.ID).
		Str("@", u.User()).
		Msg("buying")

	if order, err := createOrder(u, p.NewMarketBuyOrder()); err == nil {

		size := order.Size
		price := util.Float64(order.ExecutedValue) / util.Float64(size)

		log.Info().Str("#", p.ID).Str("@", u.User()).Str("$", util.Usd(price)).Msg("entry")

		if exit, err := sell(u, price, size, p); err != nil {
			log.Error().Send()
			log.Error().Err(err).Str("@", u.User()).Str("#", p.ID).Msg("selling")
			log.Error().Send()
		} else {
			log.Info().Str("#", p.ID).Str("@", u.User()).Str("$", util.Usd(*exit)).Msg("exit")
		}

	} else if util.IsInsufficientFunds(err) {

		log.Warn().
			Err(err).
			Str("@", u.User()).
			Str("#", p.ID).
			Msg("Insufficient funds ... sleeping ...")

		time.Sleep(time.Hour) // todo check if has funds and if more sleep required

	} else {
		log.Error().Send()
		log.Error().Err(err).Str("@", u.User()).Str("#", p.ID).Msg("buying")
		log.Error().Send()
	}
}

func sell(u *config.Session, price float64, size string, p cbp.Product) (*float64, error) {

	goal := p.GoalPrice(price)

	log.Info().Str("@", u.User()).Str("#", p.ID).Str("$", util.Usd(goal)).Msg("selling")

	// ws conn
	var wsDialer ws.Dialer
	var wsConn *ws.Conn
	var err error
	if wsConn, _, err = wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
		log.Error().Err(err).Str("@", u.User()).Str("#", p.ID).Msg("opening ws")
		if _, err = createOrder(u, p.NewLimitSellEntryOrder(goal, size)); err != nil {
			log.Error().Err(err).Str("@", u.User()).Str("#", p.ID).Msg("while creating entry order")
		}
		return nil, err
	}
	defer func(wsConn *ws.Conn) {
		if err = wsConn.Close(); err != nil {
			log.Error().Err(err).Str("@", u.User()).Str("#", p.ID).Msg("closing ws")
		}
	}(wsConn)

	// when do we start to sweat?
	timeToBreakEven := time.Now().Add(time.Minute * 45)
	for {

		// get the last price
		var lastPrice *float64
		lastPrice, err = getPrice(wsConn, p.ID)
		// rarely if ever happens
		if err != nil {
			log.Error().Err(err).Msg("while getting sell price")
			// but jic, create a stop entry with our goal price
			if _, err = createOrder(u, p.NewLimitSellEntryOrder(goal, size)); err != nil {
				// this is the worst case scenario
				log.Error().Err(err).Msg("while creating stop entry order")
			}
			// propagate error back to mother ship
			return nil, err
		}

		// if we've met or exceeded our goal price, or ...
		if *lastPrice >= goal ||
			// if we haven't met our goal, has it been 45 minutes yet?
			(time.Now().After(timeToBreakEven) &&
				// if we can get our money back, with fees
				*lastPrice >= price+(price*u.Maker)) {
			// then anchor and climb.
			return anchor(u, goal, *lastPrice, size, p)
		}

		// else, get the next price and keep the dream alive that it meets or exceeds our goal price.
	}
}

func anchor(u *config.Session, goal float64, price float64, size string, p cbp.Product) (*float64, error) {

	// create a stop loss
	order, err := createOrder(u, p.NewLimitSellLossOrder(price, size))
	if err != nil {
		// this one is bad
		return &price, err
	}

	// you're safe to climb, try and find a better sell price.
	return climb(u, goal, size, order.ID, p)
}

func climb(u *config.Session, goal float64, size, orderId string, p cbp.Product) (*float64, error) {

	log.Info().Str("@", u.User()).Str("#", p.ID).Msg("climbing")

	bestPrice := goal
	var err error
	for {

		var rate *cbp.Rate
		rate, err = getRate(p.ID)
		if err != nil {
			break
		}

		if rate.Low <= goal {
			log.Info().Str("@", u.User()).Str("#", p.ID).Msg("low <= goal :(")
			break
		}

		if rate.Close > goal {

			bestPrice = rate.Close

			log.Info().Str("@", u.User()).Str("#", p.ID).Msg("close > goal :)")

			err = cancelOrder(u, orderId)
			if err != nil {
				break
			}

			var order *cb.Order
			order, err = createOrder(u, p.NewLimitSellLossOrder(bestPrice, size))
			if err != nil {
				break
			}

			_, err = climb(u, bestPrice, size, order.ID, p)
			if err != nil {
				break
			}
		}
	}

	log.Info().Str("@", u.User()).Str("#", p.ID).Msg("climbed")

	return &bestPrice, err
}

func getRate(productId string) (*cbp.Rate, error) {

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

		now := time.Now()
		if now.After(end) {
			log.Debug().Str("product", productId).Msg("rate")
			return &cbp.Rate{
				now.UnixNano(),
				productId,
				cb.HistoricRate{now, low, high, open, *price, vol},
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
func createOrder(u *config.Session, order *cb.Order, attempt ...int) (*cb.Order, error) {

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
func getOrder(u *config.Session, id string, attempt ...int) (*cb.Order, error) {

	log.Info().Str("user", u.User()).Str("order", id).Msg("get order")

	order, err := u.GetClient().GetOrder(id)

	if err != nil {

		i := util.FirstIntOrZero(attempt)

		log.Debug().Err(err).Str("user", u.User()).Str("id", id).Send()

		if i > 10 {
			log.Error().Err(err).Str("user", u.User()).Str("id", id).Send()
			return nil, err
		}

		i++
		time.Sleep(time.Duration(i*3) * time.Second)
		return getOrder(u, id, i)
	}

	if order.Status == "pending" {

		log.Debug().
			Str("user", u.User()).
			Str("product", order.ProductID).
			Str("order", id).
			Str("side", order.Side).
			Str("type", order.Type).
			Msg("got order, but it's pending")

		time.Sleep(1 * time.Second)
		return getOrder(u, id, 0)
	}

	log.Info().Str("user", u.User()).Str("product", order.ProductID).Str("order", id).Msg("got order")

	return &order, nil
}

// cancelOrder is a recursive function that cancels an order equal to the given id.
func cancelOrder(u *config.Session, id string, attempt ...int) error {

	log.Info().
		Str("user", u.User()).
		Str("order", id).
		Msg("canceling order")

	err := u.GetClient().CancelOrder(id)
	if err == nil {
		log.Info().Str("user", u.User()).Str("order", id).Msg("canceled order")
		return nil
	}

	i := util.FirstIntOrZero(attempt)

	log.Error().
		Err(err).
		Str("user", u.User()).
		Str("order", id).
		Int("attempt", i).
		Msg("error canceling order")

	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return cancelOrder(u, id, i)
}
