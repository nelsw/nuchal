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
	"bufio"
	"fmt"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session of the application - only active while the executable is run within the defined period range.
type Session struct {
	Patterns []cbp.Pattern `yaml:"patterns"`

	// Period is a range of time representing when to start and stop executing the trade command.
	Period struct {

		// Alpha defines when command functionality should start.
		Alpha *time.Time `envconfig:"PERIOD_ALPHA" yaml:"alpha"`

		// Omega defines when command functionality should cease.
		Omega *time.Time `envconfig:"PERIOD_OMEGA" yaml:"omega"`

		// Duration is the amount of time the command should be available.
		// sim uses this as the amount of time to host result pages.
		// trade uses this to override Alpha and Omega values.
		Duration *time.Duration `envconfig:"PERIOD_DURATION" yaml:"duration"`
	} `yaml:"period"`

	patterns map[string]cbp.Pattern

	usdSelections map[string]string

	started *time.Time
}

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if util.IsEnvVarTrue("DEBUG") {
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
}

// NewSession reads configuration from environment variables and validates it
func NewSession(cfg string, usd []string, size, gain, loss, delta float64, debug ...bool) (*Session, error) {

	fmt.Println(util.Banner)
	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... hello " + os.Getenv("USER"))
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")

	if debug != nil && len(debug) > 0 && debug[0] {
		if err := os.Setenv("DEBUG", "true"); err != nil {
			return nil, err
		}
	}

	session := new(Session)

	// is the database established?
	if err := db.Init(); err != nil {
		return nil, err
	}
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... database connected")

	// can we connect to coinbase?
	if now, err := cbp.Init(cfg); err != nil {
		return nil, err
	} else {
		session.started = now
		log.Info().Msg(util.Fish + " ... coinbase validated")
		log.Info().Msg(util.Fish + " .. ")
		log.Info().Msgf("%s ... cryptocurrency products ready [%d]", util.Fish, len(cbp.GetAllProductIDs()))
	}

	if f, err := os.Open(cfg); err == nil {
		_ = yaml.NewDecoder(f).Decode(session)
	}

	// Map all patterns by product ID
	patterns := map[string]cbp.Pattern{}
	for _, pattern := range session.Patterns {
		patterns[pattern.ID] = pattern
	}
	log.Info().Msgf("%s ... patterns configurations found [%d]", util.Fish, len(patterns))

	for _, productID := range cbp.GetAllProductIDs() {
		if _, ok := patterns[productID]; ok {
			pattern := patterns[productID]
			pattern.InitPattern(size, gain, loss, delta)
			patterns[productID] = pattern
		} else {
			patterns[productID] = *cbp.NewPattern(productID, size, gain, loss, delta)
		}
	}
	log.Info().Msgf("%s ... patterns configurations ready [%d]", util.Fish, len(patterns))

	log.Info().Msgf("%s ... USD currency selections found [%d]", util.Fish, len(usd))
	session.usdSelections = map[string]string{}
	if len(usd) > 0 {
		for _, currency := range usd {
			productID := currency + "-USD"
			session.usdSelections[productID] = fmt.Sprintf("%5s", currency)
		}
	} else {
		for _, productID := range cbp.GetAllProductIDs() {
			currency := strings.Split(productID, "-")[0]
			session.usdSelections[productID] = fmt.Sprintf("%5s", currency)
		}
	}
	log.Info().Msgf("%s ... USD currency selections ready [%d]", util.Fish, len(session.usdSelections))
	log.Info().Msg(util.Fish + " .. ")

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				log.Info().Msg(util.Fish + " .")
				log.Info().Msg(util.Fish + " ..")
				log.Info().Msg(util.Fish + " ...")
				log.Info().Msg(util.Fish + " ... goodbye")
				log.Info().Msg(util.Fish + " ...")
				log.Info().Msg(util.Fish + " ..")
				log.Info().Msg(util.Fish + " .")
				os.Exit(0)
			} else {
				log.Info().Msgf("%s ... I'm not familiar with %s", util.Fish, line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error().Err(err).Send()
			panic(err)
		}
	}()

	return session, nil
}

func (s *Session) GetPattern(productID string) *cbp.Pattern {
	p := s.patterns[productID]
	return &p
}

func (s *Session) SimPort() int {

	port := os.Getenv("PORT")
	if len(port) < 4 {
		port = "8080"
	}

	prt, err := strconv.Atoi(port)
	if err != nil {
		log.Error().Err(err).Send()
		prt = 8080
	}

	return prt
}

func (s *Session) GetCurrency(productID string) *string {
	currency := s.usdSelections[productID]
	return &currency
}

func (s *Session) UsdSelectionProductIDs() []string {
	var productIDs []string
	for productID := range s.usdSelections {
		productIDs = append(productIDs, productID)
	}
	sort.Strings(productIDs)
	return productIDs
}

// GetTradingPositions returns a map of trading positions.
func (s *Session) GetTradingPositions() (map[string]cbp.Position, error) {

	positions, err := cbp.GetActivePositions()
	if err != nil {
		return nil, err
	}

	result := map[string]cbp.Position{}
	for productID, position := range positions {
		if position.Currency == "USD" || position.Balance() == position.Hold() {
			continue
		}
		result[productID] = position
	}

	return result, nil
}

// InPeriod is an exclusive range function to determine if the given time falls within the defined period.
func (s *Session) InPeriod(t time.Time) bool {
	return s.Start().Before(t) && s.Stop().After(t)
}

// Start returns the configured Start time. If no time is configured, Start returns today at noon UTC.
func (s *Session) Start() *time.Time {
	if s.Period.Alpha.Year() != 1 {
		return s.Period.Alpha
	}
	then, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT12:00:00+00:00", year(), month(), day()))
	return &then
}

// Stop returns the configured Stop time. If no time is configured, Stop returns today at 10pm UTC.
func (s *Session) Stop() *time.Time {
	if s.Period.Omega.Year() != 1 {
		return s.Period.Omega
	}
	then, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT22:00:00+00:00", year(), month(), day()))
	return &then
}

func year() int {
	return time.Now().Year()
}

func month() string {
	m := int(time.Now().Month())
	s := strconv.Itoa(m)
	if m < 10 {
		return "0" + s
	}
	return s
}

func day() string {
	d := time.Now().Day()
	s := strconv.Itoa(d)
	if d < 10 {
		return "0" + s
	}
	return s
}
