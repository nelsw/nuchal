package model

import (
	"fmt"
	"time"
)

type Simulation struct {
	Courses []Course
	Ether,
	Winners,
	Losers,
	Won,
	Lost,
	Result,
	Volume float64
	time.Duration
}

func NewSimulation(duration time.Duration) *Simulation {
	simulation := new(Simulation)
	simulation.Duration = duration
	return simulation
}

func (s Simulation) Log() {

	for _, course := range s.Courses {
		s.Ether += course.Ether
		s.Winners += course.Winners
		s.Won += course.Won
		s.Losers += course.Winners
		s.Lost += course.Lost
		s.Volume += course.Vol
		s.Result += course.Result()
		course.Log()
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
