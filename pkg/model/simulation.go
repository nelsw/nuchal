package model

import (
	"fmt"
	"time"
)

type Simulation struct {
	time.Duration

	Series []Series

	Ether,
	Winners,
	Losers int

	Won,
	Lost,
	Result,
	Volume float64
}

func NewSimulation(duration time.Duration) *Simulation {
	simulation := new(Simulation)
	simulation.Duration = duration
	return simulation
}

func (s Simulation) Log() {

	for _, series := range s.Series {

		s.Ether += series.EtherLen()
		s.Winners += series.WonLen()
		s.Losers += series.LostLen()
		s.Won += series.WonSum()
		s.Lost += series.LostSum()
		s.Volume += series.Volume()
		s.Result += series.Result()

		series.Log()
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("     won", s.Won)
	fmt.Println("    lost", s.Lost)
	fmt.Println("  result", s.Result)
	fmt.Println("  volume", s.Volume)
	fmt.Println("duration", s.Duration)
	fmt.Println()
	fmt.Println()
}
