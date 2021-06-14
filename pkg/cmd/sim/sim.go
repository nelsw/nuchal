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

package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var rates []cbp.Rate

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New(session *config.Session, winnersOnly, noLosers bool) error {

	log.Info().Msg(util.Tuna + " .")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Msg(util.Tuna + " ... simulation")
	log.Info().Msg(util.Tuna + " ..")

	var simulations []simulation

	start := time.Now()

	for _, productID := range session.UsdSelectionProductIDs() {
		initRates(session, productID)
	}
	log.Info().Msg(util.Tuna + " ..")

	for _, productID := range session.UsdSelectionProductIDs() {

		var s simulation
		newSimulation(session, productID, rates, &s)
		if s.TotalEntries() == 0 ||
			((noLosers || winnersOnly) && s.LostLen() > 0) ||
			winnersOnly && s.TradingLen() > 0 {
			continue
		}

		simulations = append(simulations, s)

		log.Info().Msg(util.Tuna + util.Break + util.GetCurrency(productID) + util.Break + util.Flag)
		log.Info().Msg(util.Tuna + " ..")
	}

	go NewResult(session, simulations, start)

	return newSite(simulations)
}

func initRates(session *config.Session, productID string) {

	alpha := *session.Alpha
	omega := *session.Omega
	pg := db.NewDB(&cbp.Rate{})

	pg.Where("product_id = ?", productID).
		Where("unix BETWEEN ? AND ?", alpha.UnixNano(), omega.UnixNano()).
		Order("unix asc").
		Find(&rates)

	if len(rates) == 0 ||
		rates[0].Time().Sub(alpha).Minutes() > 3 ||
		rates[len(rates)-1].Time().Sub(omega).Minutes() > 3 {
		if out, err := cbp.GetRates(productID, session.RateParams()); err != nil {
			log.Debug().Err(err).Msgf("%s ... %s", util.Tuna, util.GetCurrency(productID))
			return
		} else {
			log.Info().
				Int("coinbase", len(out)).
				Msgf("%s ... %s ... %s", util.Tuna, util.GetCurrency(productID), util.Check)
			for _, rate := range out {
				rates = append(rates, rate)
				pg.Create(&rate)
			}
		}
	} else {
		log.Info().
			Int("database", len(rates)).
			Msgf("%s ... %s ... %s", util.Tuna, util.GetCurrency(productID), util.Check)
	}

}

func newSite(simulations []simulation) error {

	if err := util.MakePath("html"); err != nil {
		return err
	}

	for _, simulation := range simulations {
		if err := newPage(simulation.productID, simulation.symbol(), "won", simulation.Won); err != nil {
			return err
		}
		if err := newPage(simulation.productID, simulation.symbol(), "lst", simulation.Lost); err != nil {
			return err
		}
		if err := newPage(simulation.productID, simulation.symbol(), "dnf", simulation.Trading); err != nil {
			return err
		}
	}

	fs := http.FileServer(http.Dir("html"))

	log.Print(http.ListenAndServe(fmt.Sprintf("localhost:%d", port()), logRequest(fs)))

	return nil
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msgf("%s ... %s %s %s", util.Tuna, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func port() int {
	if prt, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		return prt
	}
	return 8080
}

func newPage(productID, symbol, dir string, charts []Chart) error {

	if len(charts) < 1 {
		return nil
	}

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout
	currency := strings.TrimSpace(util.GetCurrency(productID))
	size := strconv.Itoa(len(charts))
	page.PageTitle = fmt.Sprintf("nuchal | %s > %s (%s)", symbol, currency, size)

	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].result() > charts[j].result()
	})

	for _, s := range charts {
		page.AddCharts(s.kline())
	}

	if err := util.MakePath("html/" + productID); err != nil {
		return err
	} else if f, err := os.Create(fmt.Sprintf("./html/%s/%s.html", productID, dir)); err != nil {
		return err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	return nil
}
