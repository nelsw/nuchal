package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
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
	path = "html"
	fee  = 0.005
)

type Simulation struct {
	Situations []Situation
	Ether,
	Wins,
	Loss,
	Won,
	Lost,
	Sum,
	Volume float64
	Duration time.Duration
}

func (s Simulation) Log() {
	fmt.Println()
	fmt.Println()
	fmt.Println(" won ", s.Won)
	fmt.Println("lost ", s.Lost)
	fmt.Println(" sum ", s.Sum)
	fmt.Println(" vol ", s.Volume)
	fmt.Println(" dur ", s.Duration)
	fmt.Println()
	fmt.Println()
}

type Situation struct {
	model.Posture

	Charts []Chart

	PatternLen int

	Ether,
	Won,
	Lost,
	Winners,
	Losers,
	Vol float64

	From,
	To time.Time
}

type Chart struct {
	Rates    []model.Candlestick
	Duration time.Duration
	Entry,
	Goal,
	Exit,
	Result,
	Size float64
	FoundExit bool
}

func (c Chart) Volume() float64 {
	return c.Entry * c.Size
}

func (c Chart) Kline() *charts.Kline {

	kline := charts.NewKLine()

	title := fmt.Sprintf("RESULT: %f\tBOUGHT: %f\tGOAL: %f\tSOLD: %f\tVOL: %f",
		c.Result, c.Entry, c.Goal, c.Exit, c.Volume())
	subtitle := fmt.Sprintf("Duration: %s\t Exited: %v\t", c.Duration.String(), c.FoundExit)

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
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
	for i := 0; i < len(c.Rates); i++ {
		x = append(x, c.Rates[i].Time().String())
		y = append(y, opts.KlineData{Value: c.Rates[i].Data()})
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
	return kline
}

func (s Situation) Result() float64 {
	return s.Won + s.Lost
}

func (s Situation) Log() {
	fmt.Println()
	fmt.Println("productId", s.ProductId())
	fmt.Println("     from", s.From)
	fmt.Println("       to", s.To)
	fmt.Println("  entries", s.PatternLen)
	fmt.Println("    exits", len(s.Charts))
	fmt.Println("    ether", s.Ether)
	fmt.Println("      won", s.Won)
	fmt.Println("     lost", s.Lost)
	fmt.Println("  winners", s.Winners)
	fmt.Println("   losers", s.Losers)
	fmt.Println("   volume", s.Vol)
	fmt.Println("   result", s.Result())
	fmt.Println()
}

func init() {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			panic(err)
		}
	}
}

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New() error {

	c, err := config.NewConfig()
	if err != nil {
		return err
	}

	var won, lost, sum, vol float64
	for _, posture := range c.Postures {

		rates := GetRates(c, posture.ProductId())
		indexes := NewPositionIndexes(rates, posture)
		situation, err := NewSituation(rates, indexes, posture)
		if err != nil {
			return err
		}

		won += situation.Won
		lost += situation.Lost
		sum += situation.Result()
		vol += situation.Vol

		situation.Log()
	}

	simulation := new(Simulation)
	simulation.Won = won
	simulation.Lost = lost
	simulation.Sum = sum
	simulation.Volume = vol
	simulation.Duration = *c.Duration

	simulation.Log()

	if c.IsTestMode() {
		return nil
	}

	return util.DoIndefinitely(func() {
		fs := http.FileServer(http.Dir(path))
		fmt.Println("served charts at http://localhost:8089")
		log.Print(http.ListenAndServe("localhost:8089", logRequest(fs)))
	})

}

func NewPositionIndexes(rates []model.Candlestick, posture model.Posture) []int {

	var positionIndexes []int
	var then, that model.Candlestick

	for i, this := range rates {
		if model.IsTweezer(then, that, this, posture.DeltaFloat()) {
			positionIndexes = append(positionIndexes, i)
			then = model.Candlestick{}
			that = model.Candlestick{}
		} else {
			then = that
			that = this
		}
	}

	return positionIndexes
}

func NewSituation(rates []model.Candlestick, positionIndexes []int, posture model.Posture) (*Situation, error) {

	var won, winners, lost, losers, vol, ether float64
	var chartArray []Chart

	// i == this (green, tweezer)
	// i - 1 == that (red, tweezer)
	// i - 2 == then (red)
	for _, i := range positionIndexes {

		var foundExit, foundGoal bool
		var exit, result float64

		positionRates := rates[i:]
		positionRatesLen := len(positionRates)

		firstRate := positionRates[0]
		firstRateTime := firstRate.Time()
		entry := firstRate.Open

		gain := posture.GainPrice(entry)
		loss := posture.LossPrice(entry)

		var lastRate model.Candlestick
		for j, r := range positionRates {

			// first up, did we take a bath?
			if r.Low <= loss {
				exit = loss
				foundExit = true
			}

			// have we established an exit and now below said exit?
			if r.Low <= exit {
				foundExit = true
			}

			// k, cool, but did we meet our goal?
			if r.High >= gain {

				foundGoal = true

				// if this is the first candle
				if exit == 0 {
					exit = gain // and worst case scenario, we exit with the goal.
					foundExit = true
				}

				// if we're climbing
				if r.Close >= exit {
					exit = r.Close
					foundExit = true
				}

			}

			// tweezer tops?
			if foundGoal &&
				r.IsDown() &&
				lastRate.IsUp() &&
				model.IsTweezerTop(lastRate, r, posture.DeltaFloat()*5) {
				foundExit = true
			}

			if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*15 && r.High >= gain-(gain*.001) {
				exit = gain - (gain * .001)
				foundExit = true
			}

			if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*60 && r.High >= entry+(entry*fee) {
				exit = entry
				foundExit = true
			}

			if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*90 && r.High >= entry {
				exit = entry
				foundExit = true
			}

			lastRate = r

			if !foundExit {
				if positionRatesLen-1 == j {
					fmt.Println("found you")
				}
				continue
			}

			result = (exit - (exit * fee)) - (entry + (entry * fee))
			size := util.Float64(posture.Size)

			result *= size
			if result > 0 {
				won += result
				winners++
			} else {
				lost += result
				losers++
			}

			vol += entry * size

			chartArray = append(chartArray, Chart{
				rates[i-2 : i+j+3],
				lastRate.Time().Sub(firstRateTime),
				entry,
				gain,
				exit,
				result,
				size,
				foundExit,
			})
			break
		}

		if !foundExit {
			ether++
			chartArray = append(chartArray, Chart{
				rates[i-2:],
				lastRate.Time().Sub(firstRateTime),
				entry,
				gain,
				exit,
				result,
				util.Float64(posture.Size),
				foundExit,
			})
		}

	}

	simulation := new(Situation)
	simulation.Posture = posture
	simulation.Charts = chartArray
	simulation.PatternLen = len(positionIndexes)
	simulation.Ether = ether
	simulation.Won = won
	simulation.Lost = lost
	simulation.Winners = winners
	simulation.Losers = losers
	simulation.Vol = vol
	simulation.From = rates[0].Time()
	simulation.To = rates[len(rates)-1].Time()

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout

	sort.SliceStable(simulation.Charts, func(i, j int) bool {
		return simulation.Charts[i].Result > simulation.Charts[j].Result
	})

	for _, s := range simulation.Charts {
		page.AddCharts(s.Kline())
	}

	fileName := fmt.Sprintf("./%s/%s.html", path, simulation.ProductId())

	if f, err := os.Create(fileName); err != nil {
		return nil, err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return nil, err
	}

	return simulation, nil
}

func GetRates(c *config.Config, productId string) []model.Candlestick {

	log.Info().Msg("get rates for " + productId)

	pg := db.NewDB()

	var r model.Candlestick
	if err := pg.AutoMigrate(r); err != nil {
		panic(err)
	}

	pg.Where("product_id = ?", productId).
		Order("unix desc").
		First(&r)

	var from time.Time
	if r != (model.Candlestick{}) {
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
			rc := &model.Candlestick{
				r.Time.UnixNano(),
				productId,
				r.Low,
				r.High,
				r.Open,
				r.Close,
				r.Volume,
			}
			pg.Create(&rc)
		}

		if to.After(time.Now()) {
			break
		}

		from = to
		to = to.Add(time.Hour * 4)
		log.Info().Int("... building simulation data", len(oldRates)).Send()
	}

	var savedRates []model.Candlestick
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
