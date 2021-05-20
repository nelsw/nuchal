package account

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/util"
	"sort"
	"strings"
)

type Position struct {
	ProductId            string
	Value, Balance, Hold float64
	fills, Buys, Sells   []cb.Fill
}

func NewUsdPosition(balance string) *Position {
	b := util.Float64(balance)
	return &Position{
		ProductId: "USD",
		Balance:   b,
		Hold:      0.0,
		Value:     b,
	}
}

func NewPosition(productId, hold string, value, balance float64, fills []cb.Fill) *Position {

	p := new(Position)

	p.ProductId = productId
	p.Value = value
	p.Balance = balance
	p.Hold = util.Float64(hold)
	p.fills = fills

	for _, fill := range fills {
		if fill.Side == "buy" {
			p.Buys = append(p.Buys, fill)
		} else {
			p.Sells = append(p.Sells, fill)
		}
	}

	return p
}

func (p Position) HasOrphanBuyFills() bool {
	return p.OrphanBuyFillsLen() > 0
}

func (p Position) Url() string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, p.ProductId)
}

func (p Position) Symbol() string {
	return strings.Split(p.ProductId, "-")[0]
}

func (p *Position) LonelyBuyFills() []cb.Fill {
	qty := util.MinInt(len(p.Buys), len(p.Sells))
	result := p.Buys[qty:]
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Time().After(result[j].CreatedAt.Time())
	})
	return result
}

func (p *Position) OrphanBuyFillsLen() int {
	return len(p.OrphanBuyFills())
}

func (p *Position) LonelyBuyFillsLen() int {
	return len(p.LonelyBuyFills())
}

func (p *Position) OrphanBuyFills() []cb.Fill {
	var fills []cb.Fill
	var hold = p.Hold
	for _, fill := range p.fills {
		if p.Balance == hold {
			break
		}
		fills = append(fills, fill)
		hold += util.Float64(fill.Size)
	}
	return fills
}

func (p *Position) Log() {

	//log.Info().Msg(p.Url())
	log.Info().
		Str("#", fmt.Sprintf(`%s`, p.Symbol())).
		Str("$", util.Round2Places(p.Value/p.Balance)).
		Str("bal", fmt.Sprintf(`%.3f`, p.Balance)).
		Str("hld", fmt.Sprintf(`%.3f`, p.Hold)).
		Str("val", util.Usd(p.Value)).
		Send()

	o := p.OrphanBuyFillsLen()
	l := p.LonelyBuyFillsLen()
	lvl := zerolog.InfoLevel

	for i, f := range p.LonelyBuyFills() {

		if i == l-o {
			lvl = zerolog.WarnLevel
		}

		log.WithLevel(lvl).
			Str("#", fmt.Sprintf(`%s`, p.Symbol())).
			Float64("$", util.Float64(f.Price)).
			Send()
	}
}
