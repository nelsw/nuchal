package pkg

import (
	"math"
	rule "nchl/pkg/store"
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
	buildPositions()
	buildPostures()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].sum() > results[j].sum()
	})
}

func buildPositions() {

	var then, that Rate
	for i, this := range rates {

		if then != (Rate{}) && that != (Rate{}) && then.IsDown() && that.IsDown() && this.IsUp() {
			thatFloor := math.Min(that.Low, that.Close)
			thisFloor := math.Min(this.Low, this.Open)
			if math.Abs(thatFloor-thisFloor) <= rule.Twz {
				appendPivot(i)
			}
		}

		then = that
		that = this
	}
}

func appendPivot(i int) {
	pivots = append(pivots, i)
}

func buildPostures() {

	result := Result{0, 0, []Play{}, target}

	for _, i := range pivots {

		alpha := i - 2

		entryPrice := rates[i].Open

		exitGain := entryPrice + (entryPrice * rule.Hi)
		exitLoss := entryPrice - (entryPrice * rule.Lo)

		for j, exit := range rates[i:] {

			if exit.High >= exitGain || exit.Low <= exitLoss {
				e := exitGain
				if exit.Low <= exitLoss {
					e = exitLoss
				}
				result.addPlay(exit.Time(), entryPrice, e, rates[alpha:i+j+1])
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
		exit - enter - (enter * rule.Fee) - (exit * rule.Fee),
	}
	r.Plays = append(r.Plays, p)
	if p.Result > 0 {
		r.Won += p.Result
	} else {
		r.Lost += p.Result
	}
}
