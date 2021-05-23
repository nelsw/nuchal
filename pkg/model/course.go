package model

import (
	"fmt"
)

type Course struct {
	Posture

	Charts []Chart

	PatternLen int

	Ether,
	Won,
	Lost,
	Winners,
	Losers,
	Vol float64
}

func NewCourse(posture Posture) *Course {

	course := new(Course)
	course.Posture = posture

	return course
}

func (c Course) Result() float64 {
	return c.Won + c.Lost
}

func (c Course) Log() {
	fmt.Println()
	fmt.Println("productId", c.ProductId())
	fmt.Println("  entries", c.PatternLen)
	fmt.Println("    exits", len(c.Charts))
	fmt.Println("    ether", c.Ether)
	fmt.Println("      won", c.Won)
	fmt.Println("     lost", c.Lost)
	fmt.Println("  winners", c.Winners)
	fmt.Println("   losers", c.Losers)
	fmt.Println("   volume", c.Vol)
	fmt.Println("   result", c.Result())
	fmt.Println()
}
