package account

import (
	"github.com/rs/zerolog/log"
	"nchl/pkg/util"
)

type Portfolio struct {
	Username, Value string
	Positions       []Position
}

func NewPortfolio(name string, positions []Position) *Portfolio {

	valueFloat64 := 0.0
	for _, position := range positions {
		valueFloat64 += position.Value
	}

	valueString := util.Usd(valueFloat64)

	return &Portfolio{name, valueString, positions}
}

func (p *Portfolio) Log() {

	log.Info().
		Str("username", p.Username).
		Str("value", p.Value).
		Msg("portfolio")

	for _, position := range p.Positions {
		position.Log()
	}
}

func (p Portfolio) IsMissingOrders() bool {
	for _, position := range p.Positions {
		if position.Fills != nil && len(position.Fills) > 0 {
			return true
		}
	}
	return false
}
