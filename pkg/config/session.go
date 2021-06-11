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

package config

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sort"
	"strings"
	"time"
)

// Session of the application - only active while the executable is run within the defined period range.
type Session struct {
	*paragon
	*period
	*cull
}

// NewSession reads configuration from environment variables and validates it
func NewSession(cfg, dur string, usd []string, size, gain, loss, delta float64, debug ...bool) (*Session, error) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if util.IsEnvVarTrue("DEBUG") || debug != nil && len(debug) > 0 && debug[0] {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}

	fmt.Println(util.Banner)
	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msgf("%s ... hello %s", util.Fish, os.Getenv("USER"))
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")

	session := new(Session)

	// is the database established?
	if err := db.Init(); err != nil {
		return nil, err
	}
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... database " + util.Check)

	// can we connect to coinbase?
	var products []cbp.Product
	db.NewDB().Find(&products)
	now, err := cbp.Init(cfg, products)
	if err != nil {
		return nil, err
	}
	log.Info().Msg(util.Fish + " ... coinbase " + util.Check)
	log.Info().Msg(util.Fish + " .. ")

	allProductIDs := cbp.GetAllProductIDs()
	pg := db.NewDB(cbp.Product{})
	for _, productID := range allProductIDs {
		product := cbp.GetProduct(productID)
		pg.Create(&product)
	}

	session.period = NewPeriod(cfg, dur, now)
	log.Info().Time(util.Alpha, *session.Alpha).Msgf("%s ...   period %s", util.Fish, util.Check)
	log.Info().Time(util.Omega, *session.Omega).Msgf("%s ...          %s", util.Fish, util.Check)
	log.Info().Str(util.Duration, session.Duration.String()).Msgf("%s ...          %s", util.Fish, util.Check)
	log.Info().Msg(util.Fish + " .. ")

	log.Info().
		Int(util.Quantity, len(allProductIDs)).
		Msgf("%s ... products %s", util.Fish, util.Check)

	session.paragon = NewParagon(cfg, size, gain, loss, delta)
	var pat []string
	for _, pattern := range session.paragon.patterns {
		pat = append(pat, pattern.ID)
	}
	sort.Strings(pat)

	log.Info().
		Int(util.Quantity, len(pat)).
		Strs(util.Coin, *session.paragon.patternIDs()).
		Msgf("%s ... patterns %s", util.Fish, util.Check)

	if len(usd) > 0 {
		log.Info().
			Int(util.Quantity, len(usd)).
			Strs(util.Coin, usd).
			Msgf("%s ... selected %s", util.Fish, util.Check)
	} else {
		log.Info().Int(util.Quantity, len(usd)).Msgf("%s ... selected %s", util.Fish, util.ThumbsDn)
	}

	session.cull = NewCull(usd, pat, allProductIDs)
	log.Info().Msg(util.Fish + " ..")
	log.Info().
		Int(util.Quantity, len(session.cull.IDS())).
		Strs(util.Coin, session.cull.IDS()).
		Msgf("%s ... new cull %s", util.Fish, util.Check)
	log.Info().Msg(util.Fish + " .. ")

	if err := util.MakePath("html"); err != nil {
		return nil, err
	}

	return session, nil
}
