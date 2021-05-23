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
	path = "html"
	fee  = 0.005
)

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New() error {

	c, err := config.NewConfig()
	if err != nil {
		return err
	}

	simulation := model.NewSimulation(*c.Duration)

	for _, posture := range c.Postures {

		rates := GetRates(c, posture.ProductId())
		indexes := NewPositionIndexes(rates, posture)
		course, err := NewCourse(rates, indexes, posture)
		if err != nil {
			return err
		}

		simulation.Courses = append(simulation.Courses, *course)
	}

	simulation.Log()

	if c.IsTestMode() {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}

	for _, course := range simulation.Courses {
		page := &components.Page{}
		page.Assets.InitAssets()
		page.Renderer = render.NewPageRender(page, page.Validate)
		page.Layout = components.PageFlexLayout

		sort.SliceStable(course.Charts, func(i, j int) bool {
			return course.Charts[i].Result > course.Charts[j].Result
		})

		for _, s := range course.Charts {
			page.AddCharts(s.Kline())
		}

		fileName := fmt.Sprintf("./%s/%s.html", path, course.ProductId())

		if f, err := os.Create(fileName); err != nil {
			return err
		} else if err := page.Render(io.MultiWriter(f)); err != nil {
			return err
		}
	}

	return util.DoIndefinitely(func() {
		fs := http.FileServer(http.Dir(path))
		fmt.Println("served charts at http://localhost:8089")
		log.Print(http.ListenAndServe("localhost:8089", logRequest(fs)))
	})
}

func NewPositionIndexes(rates []model.Rate, posture model.Posture) []int {

	var positionIndexes []int
	var then, that model.Rate

	for i, this := range rates {
		if model.IsTweezer(then, that, this, posture.DeltaFloat()) {
			positionIndexes = append(positionIndexes, i)
			then = model.Rate{}
			that = model.Rate{}
		} else {
			then = that
			that = this
		}
	}

	return positionIndexes
}

func NewCourse(rates []model.Rate, positionIndexes []int, posture model.Posture) (*model.Course, error) {

	course := new(model.Course)
	course.Posture = posture
	course.PatternLen = len(positionIndexes)

	// i == this (green, tweezer)
	// i - 1 == that (red, tweezer)
	// i - 2 == then (red)
	for _, i := range positionIndexes {

		var foundExit bool
		var exit, result float64

		positionRates := rates[i:]
		positionRatesLen := len(positionRates)

		firstRate := positionRates[0]
		firstRateTime := firstRate.Time()
		entry := firstRate.Close

		gain := posture.GainPrice(entry)
		loss := posture.LossPrice(entry)

		var lastRate model.Rate
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
			//if foundGoal &&
			//	r.IsDown() &&
			//	lastRate.IsUp() &&
			//	model.IsTweezerTop(lastRate, r, posture.DeltaFloat()*5) {
			//	foundExit = true
			//}

			//if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*15 && r.High >= gain-(gain*.001) {
			//	exit = gain - (gain * .001)
			//	foundExit = true
			//}

			if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*30 && r.High >= entry+(entry*fee) {
				exit = entry + (entry * fee)
				foundExit = true
			}

			if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*60 && r.High >= entry {
				exit = entry
				foundExit = true
			}

			//if !foundExit && lastRate.Time().Sub(firstRateTime) > time.Minute*120 && r.High >= gain-(gain*.01) {
			//	exit = gain - (gain * .01)
			//	foundExit = true
			//}

			// tweezer tops?
			//if r.IsDown() &&
			//	lastRate.IsUp() &&
			//	model.IsTweezerTop(lastRate, r, posture.DeltaFloat()*.1) {
			//	exit = r.Close - (r.Close * fee)
			//	foundExit = true
			//}

			lastRate = r

			if !foundExit {
				if positionRatesLen-1 == j {
					fmt.Println("lost chart to the ether")
				}
				continue
			}

			result = (exit - (exit * fee)) - (entry + (entry * fee))
			size := util.Float64(posture.Size)

			result *= size
			if result > 0 {
				course.Won += result
				course.Winners++
			} else {
				course.Lost += result
				course.Losers++
			}

			course.Vol += entry * size

			course.Charts = append(course.Charts, model.Chart{
				rates[i-2 : i+j+4],
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
			course.Ether++
			course.Charts = append(course.Charts, model.Chart{
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

	return course, nil
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
