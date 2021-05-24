package model

import (
	"fmt"
	"nuchal/pkg/util"
	"time"
)

type Simulation struct {

	// Posture is an aggregate of the product to trade, and the pattern which used to trade.
	Posture

	// Won are charts where we were profitable or broke even.
	Won []Chart

	// Lost are charts where we were not profitable.
	Lost []Chart

	// Ether are charts that never completed the simulation, these are bad.
	Ether []Chart

	// Even are charts that broke even, not bad.
	Even []Chart
}

func NewSimulation(rates []Rate, posture Posture, makerFee, takerFee float64) *Simulation {

	simulation := new(Simulation)
	simulation.Posture = posture

	open, _ := time.Parse(time.RFC3339, "2021-05-23T11:00:00+00:00")
	clos, _ := time.Parse(time.RFC3339, "2021-05-23T22:00:00+00:00")

	var then, that Rate
	for i, this := range rates {

		if this.Time().Before(open) || this.Time().After(clos) {
			continue
		}

		if posture.MatchesTweezerBottomPattern(then, that, this) {

			chart := NewChart(makerFee, takerFee, rates[i-2:], posture)
			if chart.IsWinner() {
				simulation.Won = append(simulation.Won, *chart)
			} else if chart.IsLoser() {
				simulation.Lost = append(simulation.Lost, *chart)
			} else if chart.IsEther() {
				simulation.Ether = append(simulation.Ether, *chart)
			} else if chart.IsEven() {
				simulation.Even = append(simulation.Even, *chart)
			}
		}
		then = that
		that = this
	}
	return simulation
}

func (s *Simulation) WonLen() int {
	return len(s.Won)
}

func (s *Simulation) LostLen() int {
	return len(s.Lost)
}

func (s *Simulation) EtherLen() int {
	return len(s.Ether)
}

func (s *Simulation) EvenLen() int {
	return len(s.Even)
}

func (s *Simulation) WonSum() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.Result()
	}
	return sum
}

func (s *Simulation) LostSum() float64 {
	sum := 0.0
	for _, l := range s.Lost {
		sum += l.Result()
	}
	return sum
}

func (s *Simulation) Volume() float64 {
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

func (s *Simulation) Total() float64 {
	return s.WonSum() + s.LostSum()
}

func (s *Simulation) Log() {
	fmt.Println()
	fmt.Println("productId", s.ProductId())
	fmt.Println("      won", s.WonLen())
	fmt.Println("     even", s.EvenLen())
	fmt.Println("     lost", s.LostLen())
	fmt.Println("    ether", s.EtherLen())
	fmt.Println("   volume", s.Volume())
	fmt.Println("    total", s.Total())
	fmt.Println("        %", (s.Total()/s.Volume())*100)
	fmt.Println()
}
