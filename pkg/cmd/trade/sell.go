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
	ws "github.com/gorilla/websocket"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"time"
)

func NewSells(session *config.Session) error {

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade --sell")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " .")

	positions, err := session.GetTradingPositions()
	if err != nil {
		return err
	}

	log.Info().Msg(util.Trade + " ..")

	if len(positions) < 1 {
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ... no available balance found to sell")
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ..")
		log.Info().Msg(util.Trade + " .")
		return nil
	}

	fail := make(chan error)
	done := make(chan bool)
	go func() {

		for productId, position := range positions {

			log.Info().Msg(util.Trade + " ... " + productId)

			for _, trade := range position.GetActiveTrades() {

				size := trade.Fill.Size
				entryPrice := trade.Price()
				goalPrice := position.GoalPrice(entryPrice)
				entryTime := trade.CreatedAt.Time()

				log.Info().Msg(util.Trade + " ... sell")
				log.Info().Int("#", trade.TradeID).Msg(util.Trade + " ...")
				log.Info().Str(util.Currency, productId).Msg(util.Trade + " ...")
				log.Info().Float64(util.Dollar, entryPrice).Msg(util.Trade + " ...")
				log.Info().Float64(util.Target, goalPrice).Msg(util.Trade + " ...")

				exitPrice, err := NewSell(session, productId, size, entryPrice, goalPrice, entryTime)
				if err != nil {
					fail <- err
					continue
				}

				log.Info().Msg(util.Trade + " ... sold")
				log.Info().Int("#", trade.TradeID).Msg(util.Trade + " ...")
				log.Info().Str(util.Currency, productId).Msg(util.Trade + " ...")
				log.Info().Float64(util.Dollar, *exitPrice).Msg(util.Trade + " ...")
				log.Info().Float64(util.Target, goalPrice).Msg(util.Trade + " ...")
			}
			log.Info().Msg(util.Trade + " ..")
		}
		log.Info().Msg(util.Trade + " .")
		done <- true
	}()

	for {
		select {
		case err := <-fail:
			log.Error().Err(err).Send()
		case _ = <-done:
			return nil
		}
	}
}

// NewSell is responsible for selling an available product balance at a goal price or better.
func NewSell(
	session *config.Session,
	productId,
	size string,
	entryPrice,
	goalPrice float64,
	entryTime time.Time) (*float64, error) {

	// in the event we have an error, create a limit sell entry order
	product := session.Products[productId]
	limitSellEntryOrder := product.NewLimitSellEntryOrder(goalPrice, size)

	/*
		ws connection setup
	*/
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().Err(err).Str(util.Currency, productId).Msg("opening ws")
		if _, err := session.CreateOrder(limitSellEntryOrder); err != nil {
			log.Error().Err(err).Str(util.Currency, productId).Msg("creating limit order")
			return nil, err
		}
		return nil, err
	}
	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().Err(err).Msg("closing ws")
		}
	}(wsConn)

	/*
		loop infinitely until we sell
	*/
	for {

		// get the last known price for this product
		lastPrice, err := session.GetPrice(wsConn, productId)
		if err != nil {
			log.Error().Err(err).Str(util.Currency, productId).Msg("getting sell price")
			if _, err = session.CreateOrder(limitSellEntryOrder); err != nil {
				log.Error().Err(err).Str(util.Currency, productId).Msg("creating limit order")
			}
			return nil, err
		}

		// if we've met or exceeded our goal price, or ...
		if *lastPrice >= goalPrice || // or
			// if we haven't met our goal, but it has been at least 45 minutes
			(entryTime.Add(time.Minute*45).After(time.Now()) && // and
				// if we can get our money back, with fees
				*lastPrice >= entryPrice+(entryPrice*session.Maker)) {
			// then anchor and climb.
			return anchor(session, goalPrice, size, product)
		}

		// else, get the next price and keep the dream alive that it meets or exceeds our goal price.
	}
}

// anchor attempts to create a new limit loss order for the given balance return climb.
func anchor(session *config.Session, goalPrice float64, size string, product cbp.Product) (*float64, error) {
	log.Info().
		Float64(util.Target, goalPrice).
		Str(util.Balance, size).
		Str(util.Currency, product.ID).
		Msg(util.Trade + " ... anchor")
	if order, err := session.CreateOrder(product.NewLimitLossOrder(goalPrice, size)); err != nil {
		return nil, err
	} else {
		log.Info().
			Str("#", order.ID).
			Float64(util.Target, goalPrice).
			Str(util.Balance, size).
			Str(util.Currency, product.ID).
			Msg(util.Trade + " ... ")
		return climb(session, goalPrice, size, order.ID, product)
	}
}

// climb attempts to sell the given available balance at a price greater than goal price.
// climb polls live ticker rates and looks for a rate that closes higher than the given goal price.
// climb recognizes limit loss order executions through rates with a low that is less than the given goal price.
// climb attempts to cancel the given limit loss order when a higher goal price has been found, and returns anchor.
func climb(session *config.Session, goalPrice float64, size, orderId string, product cbp.Product) (*float64, error) {

	log.Info().
		Str("#", orderId).
		Float64(util.Target, goalPrice).
		Str(util.Balance, size).
		Str(util.Currency, product.ID).
		Msg(util.Trade + " ... climb")

	for {

		rate, err := getRate(session, product.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("#", orderId).
				Float64(util.Target, goalPrice).
				Str(util.Balance, size).
				Str(util.Currency, product.ID).
				Msg(util.Trade + " ...")
			return nil, err
		}

		if rate.Low <= goalPrice { // already sold
			log.Info().
				Str("#", orderId).
				Float64(util.Target, goalPrice).
				Str(util.Balance, size).
				Str(util.Currency, product.ID).
				Msg(util.Trade + " ... fell")
			return &goalPrice, nil
		}

		if rate.Close > goalPrice {
			log.Info().
				Str("#", orderId).
				Float64(util.Target, goalPrice).
				Str(util.Balance, size).
				Str(util.Currency, product.ID).
				Msg(util.Trade + " ... camp")
			if err := session.CancelOrder(orderId); err != nil {
				return nil, err
			}
			return anchor(session, goalPrice, size, product)
		}

		log.Info().
			Str("#", orderId).
			Float64(util.Target, goalPrice).
			Str(util.Balance, size).
			Str(util.Currency, product.ID).
			Msg(util.Trade + " ...")
	}
}
