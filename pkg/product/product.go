package product

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"os"
)

type Strategy struct {
	Postures []Posture
}

type Posture struct {
	Product
	Position
}

func (p Posture) ProductId() string {
	return p.Product.Id
}

type Position struct {
	Id     string `json:"id"`
	Gain   string `json:"gain"`
	Loss   string `json:"loss"`
	Delta  string `json:"delta"`
	Size   string `json:"size"`
	Enable bool   `json:"enable,omitempty"`
}

type Product struct {
	Id             string `json:"id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`
}

func NewStrategy() (*Strategy, error) {

	log.Info().Msg("creating product strategy")

	c := new(Strategy)

	var products struct {
		All []Product `json:"products"`
	}
	if file, err := os.Open("assets/products.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open products.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&products); err != nil {
		log.Warn().Err(err).Msg("unable to decode products.json")
		return nil, err
	}

	var positions struct {
		All []Position `json:"positions"`
	}

	if file, err := os.Open("assets/positions.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open positions.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&positions); err != nil {
		log.Warn().Err(err).Msg("unable to decode positions.json")
		return nil, err
	}

	productMap := map[string]Product{}
	for _, product := range products.All {
		productMap[product.Id] = product
	}

	for _, position := range positions.All {
		if position.Enable {
			c.Postures = append(c.Postures, Posture{productMap[position.Id], position})
		}
	}

	log.Info().Msg("created product strategy")

	return c, nil
}
