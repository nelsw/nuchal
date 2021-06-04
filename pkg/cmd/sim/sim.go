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
func New(usd []string, size, gain, loss, delta float64, winnersOnly, noLosers bool) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ... simulation")
	log.Info().Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")

	var ether, winners, losers, even int
	var won, lost, total, volume float64
	simulations := map[string]Simulation{}

	var results []Simulation
	for productId, product := range ses.Products {

		rates, err := GetRates(ses, productId)
		if err != nil {
			return err
		}

		simulation := NewSimulation(rates, product, ses.Maker, ses.Taker, ses.Period)
		if simulation.Volume() == 0 {
			continue
		}

		if (noLosers || winnersOnly) && simulation.LostLen() > 0 {
			continue
		}

		if winnersOnly && simulation.EtherLen() > 0 {
			continue
		}

		results = append(results, *simulation)

		winners += simulation.WonLen()
		losers += simulation.LostLen()
		ether += simulation.EtherLen()
		won += simulation.WonSum()
		lost += simulation.LostSum()
		total += simulation.Total()
		volume += simulation.Volume()
		even += len(simulation.Even)
		simulations[productId] = *simulation
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Net() < results[j].Net()
	})

	for _, simulation := range results {

		log.Info().Str("  product", simulation.Id).Msg(util.Sim + " ...")
		log.Info().Int("  trading", simulation.EtherLen()).Msg(util.Sim + " ...")
		log.Info().Int("  winners", simulation.WonLen()).Msg(util.Sim + " ...")
		log.Info().Int("   losers", simulation.LostLen()).Msg(util.Sim + " ...")
		log.Info().Int("     even", simulation.EvenLen()).Msg(util.Sim + " ...")
		log.Info().Float64("      won", simulation.WonSum()).Msg(util.Sim + " ...")
		log.Info().Float64("     lost", simulation.LostSum()).Msg(util.Sim + " ...")
		log.Info().Float64("    total", simulation.Total()).Msg(util.Sim + " ...")
		log.Info().Float64("   volume", simulation.Volume()).Msg(util.Sim + " ...")
		log.Info().Float64("        %", simulation.Net()).Msg(util.Sim + " ...")
		log.Info().Msg(util.Sim + " ..")
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

	for productId, simulation := range simulations {

		if simulation.WonLen() > 0 {
			if err := handlePage(productId, "won", simulation.Won); err != nil {
				return err
			}
		}
		if simulation.LostLen() > 0 {
			if err := handlePage(productId, "lost", simulation.Lost); err != nil {
				return err
			}
		}
		if simulation.EtherLen() > 0 {
			if err := handlePage(productId, "ether", simulation.Ether); err != nil {
				return err
			}
		}
	}

	fs := http.FileServer(http.Dir("html"))
	log.Info().Msgf("Charts successfully served, visit them at http://localhost:%d", ses.Port)
	log.Print(http.ListenAndServe(fmt.Sprintf("localhost:%d", ses.Port), logRequest(fs)))

	time.Sleep(ses.Duration)
	return nil
}

func handlePage(productId, dir string, charts []Chart) error {

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout
	page.PageTitle = "nuchal | simulation"

	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].Result() > charts[j].Result()
	})

	for _, s := range charts {
		page.AddCharts(s.Kline())
	}

	if err := makePath("html/" + productId); err != nil {
		return err
	}

	fileName := fmt.Sprintf("./html/%s/%s.html", productId, dir)

	if f, err := os.Create(fileName); err != nil {
		return err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}
	return nil
}

func GetRates(ses *config.Session, productId string) ([]cbp.Rate, error) {

	log.Debug().Msg("get rates for " + productId)

	pg := db.NewDB()

	var r cbp.Rate
	if err := pg.AutoMigrate(r); err != nil {
		return nil, err
	}

	pg.Where("product_id = ?", productId).
		Order("unix desc").
		First(&r)

	var from time.Time
	if r != (cbp.Rate{}) {
		log.Debug().Msg("found previous rate found for " + productId)
		from = r.Time()
	} else {
		log.Debug().Msg("no previous rate found for " + productId)
		from, _ = time.Parse(time.RFC3339, "2021-06-20T00:00:00+00:00")
	}

	to := from.Add(time.Hour * 4)
	for {

		rates, err := ses.GetClient().GetHistoricRates(productId, cb.GetHistoricRatesParams{from, to, 60})
		if err != nil {
			return nil, err
		}

		for _, r := range rates {
			rc := cbp.NewRate(productId, r)
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
	pg.Where("product_id = ?", productId).
		Where("unix >= ?", ses.Alpha.UnixNano()).
		Order("unix asc").
		Find(&savedRates)

	log.Debug().Msgf("got [%d] rates for [%s]", len(savedRates), productId)

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
