package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/render"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"nuchal/pkg/config"
	"nuchal/pkg/db"
	"nuchal/pkg/model"
	"nuchal/pkg/util"
	"os"
	"sort"
	"time"
)

const (
	htmlDir = "html"
)

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New(username string, serve bool) error {

	var err error

	c, err := config.NewConfig()
	if err != nil {
		return err
	}

	user, err := c.GetUser(username)
	if err != nil {
		return err
	}

	makerFee := user.MakerFee
	takerFee := user.TakerFee

	simulation := model.NewSimulation(*c.Duration)

	for _, posture := range c.Postures {

		rates := GetRates(c, posture.ProductId())

		series := model.NewSeries(rates, posture, makerFee, takerFee)

		simulation.Series = append(simulation.Series, *series)
	}

	simulation.Log()

	if c.IsTestMode() {
		return nil
	}

	if err := makePath(htmlDir); err != nil {
		return err
	}

	for _, series := range simulation.Series {

		productId := series.Posture.ProductId()

		if series.WonLen() > 0 {
			if err := handlePage(productId, "won", series.Won); err != nil {
				return err
			}
		}
		if series.LostLen() > 0 {
			if err := handlePage(productId, "lost", series.Lost); err != nil {
				return err
			}
		}
		if series.EtherLen() > 0 {
			if err := handlePage(productId, "ether", series.Ether); err != nil {
				return err
			}
		}
	}

	return util.DoIndefinitely(func() {
		fs := http.FileServer(http.Dir(htmlDir))
		fmt.Println("served charts at http://localhost:8089")
		log.Print(http.ListenAndServe("localhost:8089", logRequest(fs)))
	})
}

func handlePage(productId, dir string, charts []model.Chart) error {

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout

	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].Result() > charts[j].Result()
	})

	for _, s := range charts {
		page.AddCharts(s.Kline())
	}

	if err := makePath(htmlDir + "/" + productId); err != nil {
		return err
	}

	fileName := fmt.Sprintf("./%s/%s/%s.html", htmlDir, productId, dir)

	if f, err := os.Create(fileName); err != nil {
		return err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}
	return nil
}

func GetRates(c *config.Config, productId string) []model.Rate {

	log.Info().Msg("get rates for " + productId)

	pg := db.NewDB()

	var r model.Rate
	if err := pg.AutoMigrate(r); err != nil {
		panic(err)
	}

	pg.Where("product_id = ?", productId).
		Order("unix desc").
		First(&r)

	var from time.Time
	if r != (model.Rate{}) {
		log.Info().Msg("found previous rate found for " + productId)
		from = r.Time()
	} else {
		log.Info().Msg("no previous rate found for " + productId)
		from, _ = time.Parse(time.RFC3339, "2021-05-20T00:00:00+00:00")
	}

	to := from.Add(time.Hour * 4)
	for {

		oldRates := getHistoricRates(c.RandomClient(), productId, from, to)

		for _, r := range oldRates {
			rc := model.NewRate(productId, r)
			pg.Create(&rc)
		}

		if to.After(time.Now()) {
			break
		}

		from = to
		to = to.Add(time.Hour * 4)
		log.Info().Int("... building simulation data", len(oldRates)).Send()
	}

	var savedRates []model.Rate
	pg.Where("product_id = ?", productId).
		Where("unix >= ?", c.StartTime().UnixNano()).
		Order("unix asc").
		Find(&savedRates)

	log.Info().Msgf("got [%d] rates for [%s]", len(savedRates), productId)

	return savedRates
}

func getHistoricRates(client *cb.Client, productId string, from, to time.Time, attempt ...int) []cb.HistoricRate {
	var i int
	if attempt != nil && len(attempt) > 0 {
		i = attempt[0]
	}
	if rates, err := client.GetHistoricRates(productId, cb.GetHistoricRatesParams{
		from,
		to,
		60,
	}); err != nil {
		log.Error().Err(err).Msg("error getting historic rate")
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return getHistoricRates(client, productId, from, to, i)
	} else {
		log.Debug().Int("qty", len(rates)).Msg("get historic rates")
		return rates
	}
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
