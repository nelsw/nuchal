package pkg

import (
	"math"
	"sort"
	"time"
)

type Play struct {
	Time                time.Time
	Rates               []Rate
	Enter, Exit, Result float64
}

type Result struct {
	Won, Lost float64
	Plays     []Play
	Target
}

func (r Result) bestPlays() []Play {
	sort.SliceStable(r.Plays, func(i, j int) bool {
		return r.Plays[i].Result > r.Plays[j].Result
	})
	return r.Plays[:25]
}

func (r Result) recentPlays() []Play {
	sort.SliceStable(r.Plays, func(i, j int) bool {
		return r.Plays[i].Time.Unix() > r.Plays[j].Time.Unix()
	})
	return r.Plays[:25]
}

func (r Result) sum() float64 {
	return r.Won + r.Lost
}

var (
	pivots  []int
	results []Result
)

func CreateSim() {
	for _, target := range targets {
		pivots = []int{}
		buildPositions(target)
		buildPostures(target)
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].sum() > results[j].sum()
	})

	//newTarget := results[0].Target
	//db.Save(newTarget)
}

func buildPositions(target Target) {

	var then, that Rate
	for i, this := range rates {

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Abs(thatFloor-thisFloor) <= target.Tweezer {
				pivots = append(pivots, i)
			}
		}

		then = that
		that = this
	}
}

func buildPostures(target Target) {

	result := Result{0, 0, []Play{}, target}

	for _, i := range pivots {

		alpha := i - 2

		entryPrice := rates[i].Open

		exitGain := entryPrice + (entryPrice * target.Gain)
		exitLoss := entryPrice - (entryPrice * target.Loss)

		for j, exit := range rates[i:] {

			if exit.High >= exitGain {
				result.addPlay(exit.Time(), entryPrice, exitGain, rates[alpha:i+j+1])
				break
			}

			if exit.Low <= exitLoss {
				result.addPlay(exit.Time(), entryPrice, exitLoss, rates[alpha:i+j+1])
				break
			}
		}
	}

	results = append(results, result)

}

func (r *Result) addPlay(t time.Time, enter, exit float64, rates []Rate) {
	p := Play{
		t,
		rates,
		enter,
		exit,
		exit - enter - (enter * Fee) - (exit * Fee),
	}
	r.Plays = append(r.Plays, p)
	if p.Result > 0 {
		r.Won += p.Result
	} else {
		r.Lost += p.Result
	}

}
