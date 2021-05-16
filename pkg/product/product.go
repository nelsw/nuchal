package product

import (
	"encoding/json"
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/util"
	"os"
	"sort"
	"strings"
)

type Strategy struct {
	Postures []Posture
}

type Posture struct {
	Product
	Position
}

func (p *Posture) ProductId() string {
	return p.Product.Id
}

func (p *Posture) MarketEntryOrder() *cb.Order {
	return &cb.Order{
		ProductID: p.ProductId(),
		Side:      "buy",
		Size:      p.Size,
		Type:      "market",
	}
}

func (p *Posture) StopEntryOrder(price float64, size string) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: p.ProductId(),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: Price(price),
		Stop:      "entry",
	}
}

func (p *Posture) StopLossOrder(price float64, size string) *cb.Order {
	return &cb.Order{
		Price:     Price(price),
		ProductID: p.ProductId(),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: Price(price),
		Stop:      "loss",
	}
}

type Position struct {
	Id     string `json:"id"`
	Gain   string `json:"gain"`
	Loss   string `json:"loss"`
	Delta  string `json:"delta"`
	Size   string `json:"size"`
	Enable bool   `json:"enable,omitempty"`
}

func (p *Position) DeltaFloat() float64 {
	return util.Float64(p.Delta)
}

func (p *Position) GainFloat() float64 {
	return util.Float64(p.Gain)
}

func (p *Position) LossFloat() float64 {
	return util.Float64(p.Loss)
}

func (p *Position) SizeFloat() float64 {
	return util.Float64(p.Size)
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

	sort.SliceStable(c.Postures, func(i, j int) bool {
		return strings.Compare(c.Postures[i].ProductId(), c.Postures[j].ProductId()) < 1
	})

	log.Info().Msgf("created product strategy [%v]", c)

	return c, nil
}

func Price(f float64) string {
	return fmt.Sprintf("%.3f", f) // todo - get increment units dynamically from cb api
}
