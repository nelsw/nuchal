package pkg

import (
	"fmt"
	"math"
	"time"
)

type Simulation struct {
	Won, Lost float64
	Plays     []Play
	ProductId string
}

type Play struct {
	Time                time.Time
	Rates               []Rate
	Enter, Exit, Result float64
}

func NewSimulation(name, productId string) Simulation {

	fmt.Println("creating simulation")

	allRates := Rates(name, productId)

	var positionIndexes []int
	var then, that Rate

	for i, this := range allRates {
		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Abs(thatFloor-thisFloor) <= 0.01 {
				positionIndexes = append(positionIndexes, i)
			}
		}
		then = that
		that = this
	}

	var won, lost float64
	var plays []Play

	for _, i := range positionIndexes {

		alpha := i - 2

		entry := allRates[i].Open

		exitGain := entry + (entry * 0.0195)
		exitLoss := entry - (entry * 0.495)

		for j, rate := range allRates[i:] {

			if rate.High < exitGain && rate.Low > exitLoss {
				continue
			}

			var exit float64
			if rate.High >= exitGain {
				exit = exitGain
			} else {
				exit = exitLoss
			}

			result := exit - entry - (entry * 0.005) - (exit * 0.005)
			if result > 0 {
				won += result
			} else {
				lost += result
			}

			plays = append(plays, Play{
				rate.Time(),
				allRates[alpha : i+j+1],
				entry,
				exit,
				result,
			})
			break
		}
	}

	simulation := Simulation{won, lost, plays, productId}

	fmt.Println("created simulation")

	return simulation
}
