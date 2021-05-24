package model

import (
	"fmt"
	"nuchal/pkg/util"
	"time"
)

type Series struct {

	// Posture is an aggregate of the product to trade, and the pattern which used to trade.
	Posture

	Won   []Chart
	Lost  []Chart
	Ether []Chart
}

func NewSeries(rates []Rate, posture Posture, makerFee, takerFee float64) *Series {

	series := new(Series)
	series.Posture = posture

	open, _ := time.Parse(time.RFC3339, "2021-05-23T11:00:00+00:00")
	clos, _ := time.Parse(time.RFC3339, "2021-05-23T21:59:59+00:00")

	var then, that Rate
	for i, this := range rates {

		if this.Time().Before(open) || this.Time().After(clos) {
			continue
		}

		if posture.matchesTweezerBottomPattern(then, that, this) {

			chart := NewChart(makerFee, takerFee, rates[i-2:], posture)
			if chart.IsWinner() {
				series.Won = append(series.Won, *chart)
			} else if chart.IsLoser() {
				series.Lost = append(series.Lost, *chart)
			} else if chart.IsEther() {
				series.Ether = append(series.Ether, *chart)
			}
		}
		then = that
		that = this
	}
	return series
}

func (s *Series) WonLen() int {
	return len(s.Won)
}

func (s *Series) LostLen() int {
	return len(s.Lost)
}

func (s *Series) EtherLen() int {
	return len(s.Ether)
}

func (s *Series) WonSum() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.Result()
	}
	return sum
}

func (s *Series) LostSum() float64 {
	sum := 0.0
	for _, l := range s.Lost {
		sum += l.Result()
	}
	return sum
}

func (s *Series) Volume() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.EntryPlusFee()
	}
	for _, l := range s.Lost {
		sum += l.EntryPlusFee()
	}
	for _, e := range s.Ether {
		sum += e.EntryPlusFee()
	}
	return sum * util.Float64(s.Size)
}

func (s *Series) Result() float64 {
	return s.WonSum() + s.LostSum()
}

func (s *Series) Log() {
	fmt.Println()
	fmt.Println("productId", s.ProductId())
	fmt.Println("      won", s.WonLen())
	fmt.Println("     lost", s.LostLen())
	fmt.Println("    ether", s.EtherLen())
	fmt.Println("   volume", s.Volume())
	fmt.Println("   result", s.Result())
	fmt.Println()
}
