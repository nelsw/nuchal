package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"io"
	"nchl/pkg/db"
	"nchl/pkg/nuchal"
	"nchl/pkg/product"
	"nchl/pkg/rate"
	"nchl/pkg/util"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	path    = "html"
	fee     = 0.005
	asc     = "unix asc"
	desc    = "unix desc"
	query   = "product_id = ?"
	timeVal = "2021-04-01T00:00:00+00:00"
)

type Result struct {
	Won, Lost, Vol float64
	Scenarios      []Scenario
	ProductId      string
	From, To       time.Time
}

type Scenario struct {
	Time                        time.Time
	Rates                       []rate.Candlestick
	Market, Entry, Exit, Result float64
}

func (s Result) Sum() float64 {
	return s.Won + s.Lost
}

func (s Result) Result() float64 {
	return s.Sum() / s.Vol
}

func init() {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			panic(err)
		}
	}
}

func GetRates(c *nuchal.Config, productId string) []rate.Candlestick {

	log.Info().Msg("get rates for " + productId)

	pg := db.NewDB()

	var from time.Time
	var r rate.Candlestick
	pg.Where(query, productId).Order(desc).First(&r)
	if r != (rate.Candlestick{}) {
		log.Info().Msg("found previous rate found for " + productId)
		from = r.Time()
	} else {
		log.Info().Msg("no previous rate found for " + productId)
		from, _ = time.Parse(time.RFC3339, timeVal)
	}

	to := from.Add(time.Hour * 4)
	for {
		for _, r := range getHistoricRates(c.User().GetClient(), productId, from, to) {
			pg.Save(&rate.Candlestick{
				r.Time.UnixNano(),
				productId,
				r.Low,
				r.High,
				r.Open,
				r.Close,
				r.Volume,
			})
		}
		if to.After(time.Now()) {
			break
		}
		from = to
		to = to.Add(time.Hour * 4)
	}

	var savedRates []rate.Candlestick
	pg.Where(query, productId).
		Order(asc).
		Find(&savedRates)
	log.Info().Msgf("got [%d] rates for [%s]", len(savedRates), productId)

	return savedRates
}

func New() error {

	c, err := nuchal.NewConfig()
	if err != nil {
		log.Error().Err(err)
		return err
	}
	for _, posture := range c.Postures {
		NewSimulation(c, posture)
	}

	//if c.TestMode {
	//	return nil
	//}

	render()
	return nil
}

func NewSimulation(c *nuchal.Config, posture product.Posture) {

	var positionIndexes []int
	var then, that rate.Candlestick

	rates := GetRates(c, posture.ProductId())

	for i, this := range rates {
		if rate.IsTweezer(then, that, this) {
			positionIndexes = append(positionIndexes, i)
		}
		then = that
		that = this
	}

	var won, lost, vol float64
	var scenarios []Scenario

	for _, i := range positionIndexes {

		alpha := i - 2

		var entry, exit, result float64
		market := rates[i].Open

		gain := market + (market * util.Float64(posture.Gain))
		loss := market - (market * util.Float64(posture.Loss))

		for j, r := range rates[i:] {

			if r.High >= gain {
				entry = gain
				if r.Low <= entry {
					exit = entry
				} else if r.Close >= exit {
					exit = r.Close
					continue
				}
				result = exit - market - (market * fee) - (exit * fee)
			} else if r.Low <= loss {
				result = loss - market - (market * fee) - (loss * fee)
			} else {
				continue
			}

			result *= util.Float64(posture.Size)
			if result > 0 {
				won += result
			} else {
				lost += result
			}

			vol += entry

			scenarios = append(scenarios, Scenario{
				r.Time(),
				rates[alpha : i+j+2],
				market,
				entry,
				exit,
				result,
			})
			break
		}
	}

	simulation := Result{
		won,
		lost,
		vol,
		scenarios,
		posture.ProductId(),
		rates[0].Time(),
		that.Time(),
	}

	fmt.Println()
	fmt.Println("productId", simulation.ProductId)
	fmt.Println("     from", simulation.From)
	fmt.Println("       to", simulation.To)
	fmt.Println("scenarios", len(simulation.Scenarios))
	fmt.Println("      won", simulation.Won)
	fmt.Println("     lost", simulation.Lost)
	fmt.Println("   report", simulation.Sum())
	fmt.Println("   volume", simulation.Vol)
	fmt.Println("   return", simulation.Result())
	fmt.Println()

	page := components.NewPage()

	sort.SliceStable(simulation.Scenarios, func(i, j int) bool {
		return simulation.Scenarios[i].Result > simulation.Scenarios[j].Result
	})

	for _, s := range simulation.Scenarios {
		kline := charts.NewKLine()
		t := fmt.Sprintf("RESULT: %f\tMARKET: %f\tENTRY: %f\tEXIT: %f\t", s.Result, s.Market, s.Entry, s.Exit)

		kline.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: t,
			}),
			charts.WithXAxisOpts(opts.XAxis{
				SplitNumber: 20,
			}),
			charts.WithYAxisOpts(opts.YAxis{
				SplitNumber: 10,
				Scale:       true,
			}),
			charts.WithDataZoomOpts(opts.DataZoom{
				Start:      0,
				End:        100,
				XAxisIndex: []int{0},
			}),
		)
		x := make([]string, 0)
		y := make([]opts.KlineData, 0)
		for i := 0; i < len(s.Rates); i++ {
			x = append(x, s.Rates[i].Time().String())
			y = append(y, opts.KlineData{Value: s.Rates[i].Data()})
		}

		kline.SetXAxis(x).AddSeries("kline", y).
			SetSeriesOptions(
				charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
					Label: &opts.Label{
						Show: true,
					},
				}),
				charts.WithItemStyleOpts(opts.ItemStyle{
					Color0:       "#ec0000",
					Color:        "#00da3c",
					BorderColor0: "#8A0000",
					BorderColor:  "#008F28",
				}),
			)
		page.AddCharts(kline)
	}

	fileName := fmt.Sprintf("./%s/%s.html", path, simulation.ProductId)

	if f, err := os.Create(fileName); err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("error creating filename [%s]", fileName))
		panic(err)
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		panic(err)
	}
}

func render() {
	fs := http.FileServer(http.Dir(path))
	fmt.Println("served charts at http://localhost:8089")
	log.Print(http.ListenAndServe("localhost:8089", logRequest(fs)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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
		log.Error().Err(err)
		i++
		if i > 10 {
			panic(err)
		}
		time.Sleep(time.Duration(i*3) * time.Second)
		return getHistoricRates(client, productId, from, to, i)
	} else {
		return rates
	}
}
