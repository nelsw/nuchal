package pkg

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rs/zerolog/log"
	"io"
	"nchl/config"
	"nchl/pkg/db"
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
	Rates                       []Candlestick
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

func GetRates(c *config.Config, productId string, start *time.Time) []Candlestick {

	fmt.Println("finding rates")

	pg := db.NewDB()

	var from time.Time
	if start != nil {
		from = *start
	} else {
		var r Candlestick
		pg.Where(query, productId).Order(desc).First(&r)
		if r != (Candlestick{}) {
			from = r.Time()
		} else {
			from, _ = time.Parse(time.RFC3339, timeVal)
		}
	}
	pg.Save(CreateHistoricRates(c.User(), productId, from))

	var allRates []Candlestick
	pg.Where(query, productId).
		Where("unix > ?", from.UnixNano()).
		Order(asc).
		Find(&allRates)

	fmt.Println("found rates", len(allRates))

	return allRates
}

func NewSim(c *config.Config) {

	dur, err := time.ParseDuration(c.Duration)
	if err != nil {
		panic(err)
	}
	from := time.Now().Add(-dur)

	var results []Result
	for _, posture := range c.Postures {
		results = append(results, NewSimulation(c, &from, posture))
	}

	render()
}

func NewSimulation(c *config.Config, from *time.Time, posture config.Posture) Result {

	var positionIndexes []int
	var then, that Candlestick

	rates := GetRates(c, posture.ProductId(), from)

	for i, this := range rates {
		if IsTweezer(then, that, this) {
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

		gain := market + (market * Float64(posture.Gain))
		loss := market - (market * Float64(posture.Loss))

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

			result *= Float64(posture.Size)
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

	return simulation
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
