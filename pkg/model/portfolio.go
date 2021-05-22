package model

import (
	"github.com/rs/zerolog/log"
	"nuchal/pkg/util"
	"sort"
)

type Portfolio struct {
	Username,
	Cash,
	Crypto,
	Value string

	Positions []Position
}

func NewPortfolio(name string, positions []Position) *Portfolio {

	cash := 0.0
	crypto := 0.0
	for _, position := range positions {
		if position.Currency == "USD" {
			cash += position.Balance()
			continue
		}
		crypto += position.Value()
	}

	return &Portfolio{
		name,
		util.Usd(cash),
		util.Usd(crypto),
		util.Usd(cash + crypto),
		positions,
	}
}

func (p *Portfolio) Info() {
	log.Info().
		Str("#", p.Username).
		Str("cash", p.Cash).
		Str("coin", p.Crypto).
		Str("val", p.Value).
		Send()
}

func (p *Portfolio) CoinPositions() []Position {
	var positions []Position
	for _, position := range p.Positions {
		if position.Currency == "USD" || (position.Balance() == 0 && position.Hold() == 0) {
			continue
		}
		positions = append(positions, position)
	}
	sort.SliceStable(positions, func(i, j int) bool {
		return positions[i].Value() > positions[j].Value()
	})
	return positions
}
