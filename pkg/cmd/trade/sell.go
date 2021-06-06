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

// NewSells attempts to sell the available balance at or beyond goal prices.
func NewSells(session *config.Session) error {

	positions, err := session.GetTradingPositions()
	if err != nil {
		return err
	}

	var positionIds []string
	for _, product := range positions {
		positionIds = append(positionIds, product.ID)
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

					tradeID := strconv.Itoa(trade.TradeID)
					productID := trade.ProductID
					size := trade.Fill.Size
					entryPrice := trade.Price()
					entryTime := trade.CreatedAt.Time()
					goalPrice := position.GoalPrice(entryPrice)

					prt(zerolog.InfoLevel, tradeID, productID, entryPrice, goalPrice, "sell")
					if exitPrice, err := NewSell(session, tradeID, productID, size, entryPrice, goalPrice, entryTime); err != nil {
						prt(zerolog.InfoLevel, tradeID, productID, entryPrice, goalPrice, err.Error())
						done <- err
					} else {
						prt(zerolog.InfoLevel, tradeID, productID, *exitPrice, goalPrice, "sold")
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
	tradeID,
	productID,
	size string,
	entryPrice,
	goalPrice float64,
	entryTime time.Time) (*float64, error) {

	// in the event we have an error, create a limit sell entry order
	product := session.Products[productID]
	limitSellEntryOrder := product.NewLimitSellEntryOrder(goalPrice, size)

	/*
		ws connection setup
	*/
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().Err(err).Str(util.Currency, productID).Msg("opening ws")
		if _, err := session.CreateOrder(limitSellEntryOrder); err != nil {
			log.Error().Err(err).Str(util.Currency, productID).Msg("creating limit order")
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
		lastPrice, err := session.GetPrice(wsConn, productID)
		if err != nil {
			log.Error().Err(err).Str(util.Currency, productID).Msg("getting sell price")
			if _, err = session.CreateOrder(limitSellEntryOrder); err != nil {
				log.Error().Err(err).Str(util.Currency, productID).Msg("creating limit order")
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
			return anchor(session, tradeID, size, productID, goalPrice, *lastPrice)
		}

		// else, get the next price and keep the dream alive that it meets or exceeds our goal price.
		i++
		if i%15 == 0 {
			prt(zerolog.InfoLevel, tradeID, productID, *lastPrice, goalPrice, "rate")
		}
	}
}

// anchor attempts to create a new limit loss order for the given balance return climb.
func anchor(session *config.Session, id, size, productID string, goalPrice, betterPrice float64) (*float64, error) {
	prt(zerolog.InfoLevel, id, productID, goalPrice, betterPrice, "nchr")
	product := session.Products[productID]
	order, err := session.CreateOrder(product.NewLimitLossOrder(betterPrice, size))
	if err != nil {
		prt(zerolog.ErrorLevel, id, productID, goalPrice, betterPrice, err.Error())
		return nil, err
	}
	return climb(session, id, size, order.ID, productID, goalPrice, betterPrice)
}

// climb attempts to sell the given available balance at a price greater than goal price.
// climb polls live ticker rates and looks for a rate that closes higher than the given goal price.
// climb recognizes limit loss order executions through rates with a low that is less than the given goal price.
// climb attempts to cancel the given limit loss order when a higher goal price has been found, and returns anchor.
func climb(session *config.Session, tradeID, size, orderID, productID string, goalPrice, betterPrice float64) (*float64, error) {

	for {

		prt(zerolog.InfoLevel, tradeID, productID, goalPrice, betterPrice, "clmb")

		rate, err := getRate(session, productID)
		if err != nil {
			prt(zerolog.ErrorLevel, tradeID, productID, goalPrice, betterPrice, err.Error())
			return nil, err
		}

		if rate.Low <= goalPrice { // already sold
			prt(zerolog.InfoLevel, tradeID, productID, goalPrice, betterPrice, "fell")
			return &goalPrice, nil
		}

		if rate.Close > goalPrice {
			prt(zerolog.WarnLevel, tradeID, productID, goalPrice, betterPrice, "camp")
			if err := session.CancelOrder(orderID); err != nil {
				prt(zerolog.ErrorLevel, tradeID, productID, goalPrice, betterPrice, err.Error())
				return nil, err
			}
			return anchor(session, tradeID, size, productID, goalPrice, rate.Close)
		}
	}
}

func prt(level zerolog.Level, id, productID string, dollar, target float64, args ...string) {

	currency := strings.ReplaceAll(productID, "-USD", "")

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
