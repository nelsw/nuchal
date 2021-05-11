package simulation

import (
	"fmt"
	"math"
	"nchl/pkg/config"
	"nchl/pkg/history"
	"nchl/pkg/product"
	"nchl/pkg/util"
	"time"
)

type Simulation struct {
	Won, Lost, Vol float64
	Scenarios      []Scenario
	ProductId      string
	From, To       time.Time
}

func (s Simulation) Sum() float64 {
	return s.Won + s.Lost
}

func (s Simulation) Result() float64 {
	return s.Sum() / s.Vol
}

type Scenario struct {
	Time                        time.Time
	Rates                       []history.Rate
	Market, Entry, Exit, Result float64
}

func NewRecentSimulation(name, productId string) Simulation {
	fmt.Println("creating recent simulation")
	s := newSimulation(history.GetRecentRates(name, productId), productId)
	fmt.Println("created recent simulation")
	return s
}

func NewSimulation(name, productId string) Simulation {
	fmt.Println("creating simulation")
	s := newSimulation(history.GetRates(name, productId), productId)
	fmt.Println("crated simulation")
	return s
}

func newSimulation(rates []history.Rate, productId string) Simulation {
	var positionIndexes []int
	var then, that history.Rate

	for i, this := range rates {
		if then != (history.Rate{}) && that != (history.Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Abs(thatFloor-thisFloor) <= 0.01 {
				positionIndexes = append(positionIndexes, i)
			}
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

		gain := config.PricePlusStopGain(market)
		loss := config.PriceMinusStopLoss(market)

		for j, rate := range rates[i:] {

			if rate.High >= gain {
				entry = gain
				if rate.Low <= entry {
					exit = entry
				} else if rate.Close >= exit {
					exit = rate.Close
					continue
				}
				result = exit - market - (market * config.Fee()) - (exit * config.Fee())
			} else if rate.Low <= loss {
				result = loss - market - (market * config.Fee()) - (loss * config.Fee())
			} else {
				continue
			}

			result *= util.Float(product.Size(market))
			if result > 0 {
				won += result
			} else {
				lost += result
			}

			vol += entry

			scenarios = append(scenarios, Scenario{
				rate.Time(),
				rates[alpha : i+j+2],
				market,
				entry,
				exit,
				result,
			})
			break
		}
	}

	fmt.Println("created simulation")

	return Simulation{
		won,
		lost,
		vol,
		scenarios,
		productId,
		rates[0].Time(),
		that.Time(),
	}
}
