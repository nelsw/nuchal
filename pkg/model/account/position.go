package account

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/util"
)

type Position struct {
	ProductId string    `json:"product_id"`
	Value     float64   `json:"value"`
	Balance   float64   `json:"balance"`
	Hold      float64   `json:"hold"`
	Fills     []cb.Fill `json:"fills,omitempty"`
}

func (p *Position) Log() {

	log.Info().
		Str("productId", p.ProductId).
		Str("value", util.Usd(p.Value)).
		Msg("position")

	if p.Fills != nil && len(p.Fills) > 0 {
		log.Warn().
			Float64("balance", p.Balance).
			Float64("hold", p.Hold).
			Msg("position")
	}
}
