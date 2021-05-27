package model

import (
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
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
		Str(util.Dollar, p.Cash).
		Str(util.Currency, p.Crypto).
		Str(util.Sigma, p.Value).
		Msg(p.Username)
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
