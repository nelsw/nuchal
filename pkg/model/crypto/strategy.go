package crypto

import (
	"encoding/json"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/model/statistic"
	"os"
	"sort"
	"strings"
)

type Strategy struct {
	Postures []Posture
}

func NewStrategy() (*Strategy, error) {

	log.Info().Msg("creating product strategy")

	c := new(Strategy)

	var products struct {
		All []cb.Product `json:"products"`
	}
	if file, err := os.Open("config/products.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open products.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&products); err != nil {
		log.Warn().Err(err).Msg("unable to decode products.json")
		return nil, err
	}

	var patterns struct {
		Tweezer []statistic.Pattern `json:"tweezer"`
	}

	if file, err := os.Open("config/patterns.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open patterns.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&patterns); err != nil {
		log.Warn().Err(err).Msg("unable to decode patterns.json")
		return nil, err
	}

	productMap := map[string]cb.Product{}
	for _, product := range products.All {
		productMap[product.ID] = product
	}

	for _, position := range patterns.Tweezer {
		if position.Enable {
			c.Postures = append(c.Postures, Posture{productMap[position.Id], position})
		}
	}

	sort.SliceStable(c.Postures, func(i, j int) bool {
		return strings.Compare(c.Postures[i].ProductId(), c.Postures[j].ProductId()) < 1
	})

	var names []string
	for _, p := range c.Postures {
		names = append(names, p.ProductId())
	}

	csv := strings.Join(names, ", ")
	log.Info().Msgf("created product strategy [%v]", csv)

	return c, nil
}

func (s Strategy) GetPosture(productId string) Posture {
	for _, p := range s.Postures {
		if p.ProductId() == productId {
			return p
		}
	}
	panic("no posture found for " + productId)
}
