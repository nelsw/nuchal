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
	"time"
)

// NewSells attempts to sell the available balance at or beyond goal prices.
func NewSells(session *config.Session) error {

	positions, err := cbp.GetTradingPositions()
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
	log.Info().Int("positions", len(positionIds)).Msg(util.Trade + " ...")
	log.Info().Int("   trades", len(tradeIds)).Msg(util.Trade + " ...")
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

	for _, productID := range session.UsdSelectionProductIDs() {

		position := positions[productID]

		for _, trade := range position.GetActiveTrades() {

			go func(productID string, trade cbp.Trade) {

				tradeID := trade.CreatedAt.Time()
				size := trade.Fill.Size
				entryPrice := trade.Price()
				entryTime := trade.CreatedAt.Time()
				goalPrice := session.GetPattern(productID).GoalPrice(entryPrice)

				if currentPrice, err := cbp.GetTickerPrice(productID); err == nil {
					prt(zerolog.InfoLevel, tradeID, productID, entryPrice, *currentPrice, goalPrice, util.Look)
				}

				if exitPrice, err := NewSell(session, tradeID, productID, size, entryPrice, goalPrice, entryTime); err == nil {
					prt(zerolog.InfoLevel, tradeID, productID, entryPrice, *exitPrice, goalPrice, "sold")
				}

				done <- err
				return
			}(productID, trade)
		}
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
	tradeID time.Time,
	productID,
	size string,
	entryPrice,
	goalPrice float64,
	entryTime time.Time) (*float64, error) {

	// websocket connection
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().Err(err).Str("action", "open").Msgf("%s ... %5s ...", util.Trade, util.GetCurrency(productID))
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().Err(err).Str("action", "close").Msgf("%s ... %5s ...", util.Trade, util.GetCurrency(productID))
		}
	}(wsConn)

	// loop infinitely until we sell
	var i int
	for {

		// get the last known price for this product
		currentPrice, err := cbp.GetPrice(wsConn, productID)
		if err != nil {
			if err = wsConn.Close(); err != nil {
				log.Error().Err(err).Msgf("%s ... %s ... %s ", util.Trade, util.GetCurrency(productID), util.Ex)
			}
			if wsConn, _, err = wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
				log.Error().Err(err).Str("action", "open").Msgf("%s ... %5s ...", util.Trade, util.GetCurrency(productID))
				return nil, err
			}
			if currentPrice, err = cbp.GetTickerPrice(productID); err != nil {
				log.Error().Msgf("%s ... %s ... %s ", util.Trade, util.GetCurrency(productID), util.Ex)
				time.Sleep(time.Second * 5 * time.Duration(i))
				continue
			}
		}

		// if we've met or exceeded our goal price, or ...
		if *currentPrice >= goalPrice || // or
			// if we haven't met our goal, but it has been at least 45 minutes
			(entryTime.Add(time.Minute*45).After(time.Now()) && // and
				// if we can get our money back, with fees
				*currentPrice >= entryPrice+(entryPrice*cbp.Maker())) {
			// then anchor and climb.
			return anchor(session, tradeID, size, productID, entryPrice, *currentPrice, *currentPrice)
		}

		// else, get the next price and keep the dream alive that it meets or exceeds our goal price.
		i++
		if i%10 == 0 {
			prt(zerolog.InfoLevel, tradeID, productID, entryPrice, *currentPrice, goalPrice, util.Look)
		}
	}
}

// anchor attempts to create a new limit loss order for the given balance return climb.
func anchor(session *config.Session, id time.Time, size, productID string, entryPrice, currentPrice, goalPrice float64) (*float64, error) {
	prt(zerolog.WarnLevel, id, productID, entryPrice, currentPrice, goalPrice, util.Anchor)
	order, err := cbp.CreateOrder(session.GetPattern(productID).NewLimitLossOrder(currentPrice, size))
	if err != nil {
		prt(zerolog.ErrorLevel, id, productID, entryPrice, currentPrice, goalPrice, err.Error())
		return nil, err
	}
	return climb(session, id, size, order.ID, productID, entryPrice, currentPrice, goalPrice)
}

// climb attempts to sell the given available balance at a price greater than goal price.
// climb polls live ticker rates and looks for a rate that closes higher than the given goal price.
// climb recognizes limit loss order executions through rates with a low that is less than the given goal price.
// climb attempts to cancel the given limit loss order when a higher goal price has been found, and returns anchor.
func climb(session *config.Session, tradeID time.Time, size, orderID, productID string, entryPrice, currentPrice, goalPrice float64) (*float64, error) {

	for {

		prt(zerolog.WarnLevel, tradeID, productID, entryPrice, currentPrice, goalPrice, util.Climb)

		rate, err := cbp.GetRate(productID)
		if err != nil {
			prt(zerolog.ErrorLevel, tradeID, productID, entryPrice, currentPrice, goalPrice, err.Error())
			return nil, err
		}

		if rate.Close > goalPrice {
			prt(zerolog.WarnLevel, tradeID, productID, entryPrice, rate.Close, rate.Close, util.Camp)
			if err := cbp.CancelOrder(orderID); err != nil {
				prt(zerolog.ErrorLevel, tradeID, productID, entryPrice, rate.Close, rate.Close, err.Error())
				return nil, err
			}
			return anchor(session, tradeID, size, productID, entryPrice, rate.Close, rate.Close)
		}

		if rate.Low < goalPrice { // already sold
			prt(zerolog.WarnLevel, tradeID, productID, entryPrice, rate.Close, goalPrice, util.Fell)
			return &goalPrice, nil
		}
	}
}

func prt(
	level zerolog.Level,
	id time.Time,
	productID string,
	entry,
	current,
	goal float64,
	args ...string) {

	msg := util.Trade + " ... " + util.GetCurrency(productID)
	if args != nil && len(args) > 0 {
		msg = msg + " ... " + args[0]
	}

	log.WithLevel(level).
		Time("", id).
		Str("1."+util.Arrival, fmt.Sprintf("%.3f", entry)).
		Str("2."+util.Current, fmt.Sprintf("%.3f", current)).
		Str("3."+util.Target, fmt.Sprintf("%.3f", goal)).
		Msg(msg)
}
