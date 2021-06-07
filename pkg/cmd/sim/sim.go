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
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New(session *config.Session, winnersOnly, noLosers bool) error {

	intro(session)

	simulations := map[string]simulation{}

	var results []simulation
	for productID, currency := range session.UsdSelections {

		product := session.GetProduct(productID)

		rates, err := getRates(session, productID)
		if err != nil {
			return err
		}

		simulation := newSimulation(rates, product, session.Maker, session.Taker, session.Period)
		log.Info().Msg(util.Sim + util.Break + currency + util.Break + "complete")

		if simulation.Volume() == 0 {
			continue
		}

		if (noLosers || winnersOnly) && simulation.LostLen() > 0 {
			continue
		}

		if winnersOnly && simulation.TradingLen() > 0 {
			continue
		}

		results = append(results, *simulation)
		simulations[productID] = *simulation
	}

	log.Info().Msg(util.Sim + " .. ")

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Net() < results[j].Net()
	})

	var ether, winners, losers, even int
	var won, lost, total, volume float64
	for _, simulation := range results {

		log.Info().
			Float64(util.Delta, simulation.Delta).
			Float64(util.UpArrow, simulation.Gain).
			Float64(util.Quantity, simulation.Size).
			Str(util.Hyperlink, simulation.Url()).
			Msg(util.Sim + util.Break + simulation.ID)

		if simulation.WonLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.WonLen()).
				Str(util.Sigma, util.Usd(simulation.WonSum())).
				Str(util.Hyperlink, resultUrl(simulation.ID, "won", session.SimPort())).
				Msg(util.Sim + " ... " + util.ThumbsUp)
		}

		if simulation.LostLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.LostLen()).
				Str(util.Sigma, util.Usd(simulation.LostSum())).
				Str(util.Hyperlink, resultUrl(simulation.ID, "lst", session.SimPort())).
				Msg(util.Sim + " ... " + util.ThumbsDn)
		}

		if simulation.EvenLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.EvenLen()).
				Str(util.Sigma, "$0.000").
				Str(util.Hyperlink, resultUrl(simulation.ID, "evn", session.SimPort())).
				Msg(util.Sim + " ... " + util.NoTrend)
		}

		if simulation.TradingLen() > 0 {
			symbol := util.UpTrend
			if simulation.TradingSum() < 0 {
				symbol = util.DnTrend
			}
			log.Info().
				Int(util.Quantity, simulation.TradingLen()).
				Str(util.Sigma, util.Usd(simulation.TradingSum())).
				Str(util.Hyperlink, resultUrl(simulation.ID, "dnf", session.SimPort())).
				Msg(util.Sim + " ... " + symbol)
		}

		log.Info().
			Str(util.Sigma, util.Usd(simulation.Total())).
			Str(util.Quantity, util.Usd(simulation.Volume())).
			Str("%", util.Money(simulation.Net())).
			Msg(util.Sim + util.Break + simulation.symbol())

		log.Info().Msg(util.Sim + " ..")

		winners += simulation.WonLen()
		losers += simulation.LostLen()
		ether += simulation.TradingLen()
		won += simulation.WonSum()
		lost += simulation.LostSum()
		total += simulation.Total()
		volume += simulation.Volume()
		even += len(simulation.Even)
	}

	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Int("  trading", ether).Msg(util.Sim + " ...")
	log.Info().Int("  winners", winners).Msg(util.Sim + " ...")
	log.Info().Int("   losers", losers).Msg(util.Sim + " ...")
	log.Info().Int("     even", even).Msg(util.Sim + " ...")
	log.Info().Float64("      won", won).Msg(util.Sim + " ...")
	log.Info().Float64("     lost", lost).Msg(util.Sim + " ...")
	log.Info().Float64("    total", total).Msg(util.Sim + " ...")
	log.Info().Float64("   volume", volume).Msg(util.Sim + " ...")
	log.Info().Float64("        %", (total/volume)*100).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")

	if err := makePath("html"); err != nil {
		return err
	}

	for productID, simulation := range simulations {

		if simulation.WonLen() > 0 {
			if err := handlePage(productID, "won", simulation.Won); err != nil {
				return err
			}
		}
		if simulation.LostLen() > 0 {
			if err := handlePage(productID, "lst", simulation.Lost); err != nil {
				return err
			}
		}
		if simulation.TradingLen() > 0 {
			if err := handlePage(productID, "dnf", simulation.Trading); err != nil {
				return err
			}
		}
	}

	fs := http.FileServer(http.Dir("html"))
	log.Info().Msgf("Charts successfully served, visit them at http://localhost:%d", session.SimPort())
	log.Print(http.ListenAndServe(fmt.Sprintf("localhost:%d", session.SimPort()), logRequest(fs)))

	time.Sleep(session.Duration)
	return nil
}

func intro(ses *config.Session) {
	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " ... simulation")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Time(util.Alpha, *ses.Start()).Msg(util.Sim + " ...")
	log.Info().Time(util.Omega, *ses.Stop()).Msg(util.Sim + " ...")
	log.Info().Strs(util.Currency, *ses.ProductIDs()).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")
}

func handlePage(productID, dir string, charts []Chart) error {

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout
	page.PageTitle = "nuchal | simulation"

	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].result() > charts[j].result()
	})

	for _, s := range charts {
		page.AddCharts(s.kline())
	}

	if err := makePath("html/" + productID); err != nil {
		return err
	}

	fileName := fmt.Sprintf("./html/%s/%s.html", productID, dir)

	if f, err := os.Create(fileName); err != nil {
		return err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}
	return nil
}

func resultUrl(productID, dir string, port int) string {
	return fmt.Sprintf("http://localhost:%d/%s/%s.html", port, productID, dir)
}

func getRates(ses *config.Session, productID string) ([]cbp.Rate, error) {

	log.Debug().Msg("get rates for " + productID)

	pg := db.NewDB()

	var r cbp.Rate
	if err := pg.AutoMigrate(r); err != nil {
		return nil, err
	}

	pg.Where("product_id = ?", productID).
		Order("unix desc").
		First(&r)

	var from time.Time
	if r != (cbp.Rate{}) {
		log.Debug().Msg("found previous rate found for " + productID)
		from = r.Time()
	} else {
		log.Debug().Msg("no previous rate found for " + productID)
		from = ses.Alpha
	}

	to := from.Add(time.Hour * 4)
	for {

		rates, err := ses.GetClient().GetHistoricRates(productID, cb.GetHistoricRatesParams{from, to, 60})
		if err != nil {
			return nil, err
		}

		for _, r := range rates {
			rc := cbp.NewRate(productID, r)
			pg.Create(&rc)
		}

		if to.After(time.Now()) {
			break
		}

		from = to
		to = to.Add(time.Hour * 4)
		log.Debug().Int("... building simulation data", len(rates)).Send()
	}

	var savedRates []cbp.Rate
	pg.Where("product_id = ?", productID).
		Where("unix >= ?", ses.Alpha.UnixNano()).
		Order("unix asc").
		Find(&savedRates)

	log.Debug().Msgf("got [%d] rates for [%s]", len(savedRates), productID)

	return savedRates, nil
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func makePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}
