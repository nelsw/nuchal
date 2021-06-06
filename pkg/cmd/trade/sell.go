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
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sort"
	"strconv"
	"strings"
	"time"
)

func NewSells(session *config.Session) error {

	positions, err := session.GetTradingPositions()
	if err != nil {
		return err
	}

	var positionIds []string
	for productId, _ := range positions {
		positionIds = append(positionIds, productId)
	}
	sort.Strings(positionIds)

	var tradeIds []int
	for _, position := range positions {
		for _, trade := range position.GetActiveTrades() {
			tradeIds = append(tradeIds, trade.TradeID)
		}
	}

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade --sell")
	log.Info().Strs(" products", *session.ProductIds()).Msg(util.Trade + " ...")
	log.Info().Strs("positions", positionIds).Msg(util.Trade + " ...")
	log.Info().Ints("   trades", tradeIds).Msg(util.Trade + " ...")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")

	if len(positions) < 1 {
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ... no available balance found to sell")
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ..")
		log.Info().Msg(util.Trade + " .")
		return nil
	}

	done := make(chan error)

	for _, position := range positions {
		go func(position cbp.Position) {

			for _, trade := range position.GetActiveTrades() {
				go func(trade cbp.Trade) {

					tradeId := strconv.Itoa(trade.TradeID)
					productId := trade.ProductID
					size := trade.Fill.Size
					entryPrice := trade.Price()
					entryTime := trade.CreatedAt.Time()
					goalPrice := position.GoalPrice(entryPrice)

					prt(zerolog.InfoLevel, tradeId, productId, entryPrice, goalPrice, "sell")
					if exitPrice, err := NewSell(session, tradeId, productId, size, entryPrice, goalPrice, entryTime); err != nil {
						prt(zerolog.InfoLevel, tradeId, productId, entryPrice, goalPrice, err.Error())
						done <- err
					} else {
						prt(zerolog.InfoLevel, tradeId, productId, *exitPrice, goalPrice, "sold")
						done <- nil
					}
					return
				}(trade)
			}
		}(position)
	}

	completions := 0
	for {
		select {
		case err := <-done:

			completions++

			if err != nil {
				log.Error().Err(err).Msg(util.Trade + " ...")
			}

			if completions == len(tradeIds) {
				log.Info().Msg(util.Trade + " ...")
				log.Info().Msg(util.Trade + " ... available balance sold, go party.")
				log.Info().Msg(util.Trade + " ...")
				log.Info().Msg(util.Trade + " ..")
				log.Info().Msg(util.Trade + " .")
				return nil
			}
		}
	}
}

// NewSell is responsible for selling an available product balance at a goal price or better.
func NewSell(
	session *config.Session,
	tradeId,
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
	var i int
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
			return anchor(session, tradeId, size, productId, goalPrice, *lastPrice)
		}

		// else, get the next price and keep the dream alive that it meets or exceeds our goal price.
		i++
		if i%15 == 0 {
			prt(zerolog.InfoLevel, tradeId, productId, *lastPrice, goalPrice, "rate")
		}
	}
}

// anchor attempts to create a new limit loss order for the given balance return climb.
func anchor(session *config.Session, id, size, productId string, goalPrice, betterPrice float64) (*float64, error) {
	prt(zerolog.InfoLevel, id, productId, goalPrice, betterPrice, "nchr")
	product := session.Products[productId]
	if order, err := session.CreateOrder(product.NewLimitLossOrder(betterPrice, size)); err != nil {
		prt(zerolog.ErrorLevel, id, productId, goalPrice, betterPrice, err.Error())
		return nil, err
	} else {
		return climb(session, id, size, order.ID, productId, goalPrice, betterPrice)
	}
}

// climb attempts to sell the given available balance at a price greater than goal price.
// climb polls live ticker rates and looks for a rate that closes higher than the given goal price.
// climb recognizes limit loss order executions through rates with a low that is less than the given goal price.
// climb attempts to cancel the given limit loss order when a higher goal price has been found, and returns anchor.
func climb(session *config.Session, tradeId, size, orderId, productId string, goalPrice, betterPrice float64) (*float64, error) {

	for {

		prt(zerolog.InfoLevel, tradeId, productId, goalPrice, betterPrice, "climb")

		rate, err := getRate(session, productId)
		if err != nil {
			prt(zerolog.ErrorLevel, tradeId, productId, goalPrice, betterPrice, err.Error())
			return nil, err
		}

		if rate.Low <= goalPrice { // already sold
			prt(zerolog.InfoLevel, tradeId, productId, goalPrice, betterPrice, "fell")
			return &goalPrice, nil
		}

		if rate.Close > goalPrice {
			prt(zerolog.WarnLevel, tradeId, productId, goalPrice, betterPrice, "camp")
			if err := session.CancelOrder(orderId); err != nil {
				prt(zerolog.ErrorLevel, tradeId, productId, goalPrice, betterPrice, err.Error())
				return nil, err
			}
			return anchor(session, tradeId, size, productId, goalPrice, rate.Close)
		}
	}
}

func prt(level zerolog.Level, id, productId string, dollar, target float64, args ...string) {

	currency := strings.ReplaceAll(productId, "-USD", "")

	msg := util.Trade + " ... " + fmt.Sprintf("%5s", currency)
	if args != nil && len(args) > 0 {
		msg = msg + " ... " + args[0]
	}
	log.WithLevel(level).
		Str("#", id).
		Str(util.Dollar, fmt.Sprintf("%.3f", dollar)).
		Str(util.Target, fmt.Sprintf("%.3f", target)).
		Msg(msg)
}
