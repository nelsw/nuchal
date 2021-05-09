package pkg

import (
	"fmt"
	"math"
	"sort"
)

type Play struct {
	Rates []Rate
	Enter, Exit, Result float64
	Duration string
	Target
}

var (
	pivots []int
	plays, best []Play
)

func CreateSim() {
	for _, target := range targets {
		pivots = []int{}
		plays = []Play{}
		buildPositions(target)
		buildPostures(target)
	}

	sort.SliceStable(plays, func(i, j int) bool {
		return plays[i].Result > plays[j].Result
	})

	best = plays[:10]
}

func buildPositions(target Target) {

	var then, that Rate
	for i, this := range rates {

		if then == (Rate{}) {
			then = this
			continue
		}

		if that == (Rate{}) {
			that = this
			continue
		}

		if then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Abs(thatFloor- thisFloor) <= target.Tweezer {
				pivots = append(pivots, i)
			}
		}

		then = that
		that = this
	}
}

func buildPostures(target Target)  {

	totalLost := 0.0
	totalWon := 0.0

	for _, i := range pivots {

		alpha := i-2

		entryPrice := rates[i].Open
		entryFee := entryPrice * Fee

		exitGain := entryPrice + (entryPrice * target.Gain)
		exitLoss := entryPrice - (entryPrice * target.Loss)

		for j, exit := range rates[i:] {

			if exit.High >= exitGain {

				exitMargin := exitGain - entryPrice
				exitFee := exitGain * Fee
				result := exitMargin - entryFee - exitFee
				duration := exit.Time().Sub(rates[alpha].Time()).String()

				plays = append(plays, Play{
					rates[alpha : i+j+1],
					entryPrice,
					exitGain,
					result,
					duration,
					target,
				})
				totalWon += result
				break
			}

			if exit.Low <= exitLoss {

				exitMargin := entryPrice - exitLoss
				exitFee := exitLoss * Fee
				result := exitMargin - entryFee - exitFee
				duration := exit.Time().Sub(rates[alpha].Time()).String()

				plays = append(plays, Play{
					rates[alpha : i+j],
					entryPrice,
					exitLoss,
					result,
					duration,
					target,
					})
				totalLost += result
				break
			}
		}
	}

	fmt.Println("   target", target)
	fmt.Println("    plays", len(plays))
	fmt.Println("      won", totalWon)
	fmt.Println("     lost", totalLost)
	fmt.Println("   result", totalWon+totalLost)
	fmt.Println()
}