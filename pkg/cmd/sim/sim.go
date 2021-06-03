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
func New() error {

	cfg, err := config.NewSession()
	if err != nil {
		return err
	}

	var ether, winners, losers, even int
	var won, lost, total, volume float64
	simulations := map[string]Simulation{}

	for productId, product := range cfg.Products {

		rates, err := GetRates(cfg, productId)
		if err != nil {
			return err
		}

		simulation := NewSimulation(rates, product, cfg.Maker, cfg.Taker, cfg.Period)
		winners += simulation.WonLen()
		losers += simulation.LostLen()
		ether += simulation.EtherLen()
		won += simulation.WonSum()
		lost += simulation.LostSum()
		total += simulation.Total()
		volume += simulation.Volume()
		even += len(simulation.Even)
		simulations[productId] = *simulation

		simulation.Log()
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("  trading", ether)
	fmt.Println("  winners", winners)
	fmt.Println("   losers", losers)
	fmt.Println("     even", even)
	fmt.Println("      won", won)
	fmt.Println("     lost", lost)
	fmt.Println("    total", total)
	fmt.Println("   volume", volume)
	fmt.Println("        %", (total/volume)*100)
	fmt.Println()
	fmt.Println()

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

	return util.DoIndefinitely(func() {
		fs := http.FileServer(http.Dir("html"))
		log.Info().Msgf("Charts successfully served, visit them at http://%s", cfg.SimulationAddress())
		util.LogBanner()
		log.Print(http.ListenAndServe(cfg.SimulationAddress(), logRequest(fs)))
	})
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

func GetRates(c *config.Session, productId string) ([]cbp.Rate, error) {

	log.Debug().Msg("get rates for " + productId)

	pg := db.NewDB()

	var r cbp.Rate
	if err := pg.AutoMigrate(r); err != nil {
		panic(err)
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
		from, _ = time.Parse(time.RFC3339, "2021-05-20T00:00:00+00:00")
	}

	to := from.Add(time.Hour * 4)
	for {

		rates, err := c.GetClient().GetHistoricRates(productId, cb.GetHistoricRatesParams{from, to, 60})
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
		Where("unix >= ?", c.SimulationStart()).
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
