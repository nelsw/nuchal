package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/util"
	"sort"
)

type Position struct {
	cb.Account
	cb.Ticker
	fills,
	buys,
	sells []cb.Fill
}

func (p Position) Url() string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, p.ProductId())
}

func (p Position) ProductId() string {
	return p.Currency + "-USD"
}

func (p Position) Balance() float64 {
	return util.Float64(p.Account.Balance)
}

func (p Position) Hold() float64 {
	return util.Float64(p.Account.Hold)
}

func (p Position) Value() float64 {
	return util.Float64(p.Price) * p.Balance()
}

func NewPosition(account cb.Account, ticker cb.Ticker, fills []cb.Fill) *Position {

	p := new(Position)
	p.Account = account
	p.Ticker = ticker
	p.fills = fills

	for _, fill := range fills {
		if fill.Side == "buy" {
			p.buys = append(p.buys, fill)
		} else {
			p.sells = append(p.sells, fill)
		}
	}

	return p
}

func (p Position) HasOrphanBuyFills() bool {
	return p.OrphanBuyFillsLen() > 0
}

func (p *Position) LonelyBuyFills() []cb.Fill {
	qty := util.MinInt(len(p.buys), len(p.sells))
	result := p.buys
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Time().Before(result[j].CreatedAt.Time())
	})
	result = p.buys[qty:]
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
	var hold = p.Hold()
	for _, fill := range p.LonelyBuyFills() {
		if p.Balance() == hold {
			break
		}
		fills = append(fills, fill)
		hold += util.Float64(fill.Size)
	}
	sort.SliceStable(fills, func(i, j int) bool {
		return fills[i].CreatedAt.Time().After(fills[j].CreatedAt.Time())
	})
	return fills
}

func (p *Position) Log() {

	log.Info().Msg(p.Url())
	log.Info().
		Str("#", p.Currency).
		Str("$", p.Price).
		Str("bal", fmt.Sprintf(`%.3f`, p.Balance())).
		Str("hld", fmt.Sprintf(`%.3f`, p.Hold())).
		Str("val", util.Usd(p.Value())).
		Send()

	o := p.OrphanBuyFillsLen()
	l := p.LonelyBuyFillsLen()
	lvl := zerolog.InfoLevel

	for i, f := range p.LonelyBuyFills() {

		if i == l-o {
			lvl = zerolog.WarnLevel
		}

		log.WithLevel(lvl).
			Str("#", fmt.Sprintf(`%s`, p.Currency)).
			Float64("$", util.Float64(f.Price)).
			Send()
	}
}
