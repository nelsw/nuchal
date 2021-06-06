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
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"time"
)

// New will attempt to buy and sell automagically.
func New(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Time(util.Alpha, *ses.Start()).Msg(util.Trade + " ...")
	log.Info().Time(util.Omega, *ses.Cease()).Msg(util.Trade + " ...")
	log.Info().Strs(util.Currency, *ses.ProductIds()).Msg(util.Trade + " ...")
	log.Info().Msg(util.Trade + " ..")

	if util.IsEnvVarTrue("TEST") {
		return nil
	}

	return util.DoIndefinitely(func() {
		for _, p := range ses.Products {
			go trade(ses, p)
		}
	})
}

func trade(session *config.Session, p cbp.Product) {

	log.Info().Str("#", p.ID).Msg("trading")

	var then, that cbp.Rate
	for {
		if this, err := getRate(session, p.ID); err != nil {
			then = cbp.Rate{}
			that = cbp.Rate{}
		} else if !p.MatchesTweezerBottomPattern(then, that, *this) {
			then = that
			that = *this
		} else {
			go buy(session, p)
			then = cbp.Rate{}
			that = cbp.Rate{}
		}
	}
}

func buy(session *config.Session, product cbp.Product) {

	log.Info().
		Str("#", product.ID).
		Msg("buying")

	order, err := session.CreateOrder(product.NewMarketBuyOrder())
	if err == nil {

		productID := product.ID
		size := order.Size
		entryPrice := util.Float64(order.ExecutedValue) / util.Float64(size)
		goalPrice := product.GoalPrice(entryPrice)
		entryTime := order.CreatedAt.Time()

		if _, err := NewSell(session, order.ID, productID, size, entryPrice, goalPrice, entryTime); err != nil {
			log.Error().Err(err).Msg("while selling")
		}
		return
	}

	if util.IsInsufficientFunds(err) {
		log.Warn().Err(err).Msg("Insufficient funds ... sleeping ...")
		time.Sleep(time.Hour) // todo check if has funds and if more sleep required
		return
	}

	log.Error().Send()
	log.Error().Err(err).Str(util.Currency, product.ID).Msg("buying")
	log.Error().Send()
}

func getRate(session *config.Session, productID string) (*cbp.Rate, error) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("productID", productID).
			Msg("error while opening websocket connection")
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Str("productID", productID).
				Msg("error closing websocket connection")
		}
	}(wsConn)

	end := time.Now().Add(time.Minute)

	var low, high, open, vol float64
	for {

		price, err := session.GetPrice(wsConn, productID)
		if err != nil {
			log.Error().
				Err(err).
				Str("productID", productID).
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
			log.Debug().Str("product", productID).Msg("rate")
			return &cbp.Rate{
				now.UnixNano(),
				productID,
				cb.HistoricRate{now, low, high, open, *price, vol},
			}, nil
		}
	}
}
