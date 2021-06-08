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
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"time"
)

// New will attempt to buy and sell automagically.
func New(ses *config.Session) error {

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Time(util.Alpha, *ses.Start()).Msg(util.Trade + " ...")
	log.Info().Time(util.Omega, *ses.Stop()).Msg(util.Trade + " ...")
	log.Info().Strs(util.Currency, ses.UsdSelectionProductIDs()).Msg(util.Trade + " ...")
	log.Info().Msg(util.Trade + " ..")

	if util.IsEnvVarTrue("TEST") {
		return nil
	}

	exit := make(chan string)
	go func() {
		for _, productID := range ses.UsdSelectionProductIDs() {
			go trade(ses, productID)
		}
	}()
	for {
		select {
		case <-exit:
			return nil
		}
	}
}

func trade(session *config.Session, productID string) {

	log.Info().Time("", time.Now()).Msgf("%s ... %5s ... %s", util.Trade, productID, "poll")

	pattern := session.GetPattern(productID)

	var then, that cbp.Rate
	for {
		if this, err := cbp.GetRate(productID); err != nil {
			then = cbp.Rate{}
			that = cbp.Rate{}
		} else if !pattern.MatchesTweezerBottomPattern(then, that, *this) {
			then = that
			that = *this
		} else {
			go buy(session, pattern)
			then = cbp.Rate{}
			that = cbp.Rate{}
		}
	}
}

func buy(session *config.Session, product *cbp.Pattern) {

	log.Info().Time("", time.Now()).Msgf("%s ... %5s ... %s", util.Trade, product.ID, "buy")

	order, err := cbp.CreateOrder(product.NewMarketBuyOrder())
	if err == nil {

		productID := product.ID
		size := order.Size
		entryPrice := util.Float64(order.ExecutedValue) / util.Float64(size)
		goalPrice := product.GoalPrice(entryPrice)
		entryTime := order.CreatedAt.Time()

		if _, err := NewSell(session, entryTime, productID, size, entryPrice, goalPrice, entryTime); err != nil {
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
