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

const (
	f0 = "%s ...    hello %s"
	g0 = "%s ... database %s"
	h0 = "%s ... coinbase %s"
	f1 = "%s ...   period %s"
	fn = "%s ...          %s"
	f2 = "%s ... products %s"
	f3 = "%s ... patterns %s"
	f4 = "%s ... selected %s"
	f5 = "%s ... new cull %s"
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
	log.Info().Msg(util.Cichlid + " . ")
	log.Info().Msg(util.Cichlid + " .. ")
	log.Info().Msgf(f0, util.Cichlid, os.Getenv("USER"))
	log.Info().Msg(util.Cichlid + " .. ")
	log.Info().Msg(util.Cichlid + " . ")

	// is the database established?
	if err := db.Init(); err != nil {
		return nil, err
	}
	log.Info().Msg(util.Cichlid + " .. ")
	log.Info().Msgf(g0, util.Cichlid, util.Check)

	// can we connect to coinbase?
	var products []cbp.Product
	pg := db.NewDB(cbp.Product{})
	pg.Find(&products)
	now, err := cbp.Init(cfg, &products)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf(h0, util.Cichlid, util.Check)
	log.Info().Msg(util.Cichlid + " .. ")

	allProductIDs := cbp.GetAllProductIDs()
	for _, productID := range allProductIDs {
		product := cbp.GetProduct(productID)
		pg.Create(&product)
	}

	session := new(Session)
	session.period = NewPeriod(cfg, dur, now)
	log.Info().Time(util.Alpha, *session.Alpha).Msgf(f1, util.Cichlid, util.Check)
	log.Info().Time(util.Omega, *session.Omega).Msgf(fn, util.Cichlid, util.Check)
	log.Info().Str(util.Duration, session.Duration.String()).Msgf(fn, util.Cichlid, util.Check)
	log.Info().Msg(util.Cichlid + " .. ")
	log.Info().Int(util.Quantity, len(allProductIDs)).Msgf(f2, util.Cichlid, util.Check)

	session.paragon = NewParagon(cfg, size, gain, loss, delta)
	var pat []string
	for _, pattern := range session.paragon.patterns {
		pat = append(pat, pattern.ID)
	}
	sort.Strings(pat)

	ids := *session.paragon.patternIDs()
	log.Info().Int(util.Quantity, len(pat)).Strs(util.Coin, ids).Msgf(f3, util.Cichlid, util.Check)
	log.Info().Int(util.Quantity, len(usd)).Strs(util.Coin, usd).Msgf(f4, util.Cichlid, util.Check)

	session.cull = NewCull(usd, pat, allProductIDs)
	cls := session.cull.IDS()
	log.Info().Msg(util.Cichlid + " .. ")
	log.Info().Int(util.Quantity, len(cls)).Strs(util.Coin, cls).Msgf(f5, util.Cichlid, util.Check)
	log.Info().Msg(util.Cichlid + " .. ")

	return session, util.MakePath("html")
}
