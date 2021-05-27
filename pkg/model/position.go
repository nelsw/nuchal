package model

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"sort"
)

type Position struct {
	Product
	cb.Account
	cb.Ticker
	buys,
	sells []Trade
}

func (p Position) ProductId() string {
	return p.Currency + "-USD"
}

func (p Position) Url() string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, p.ProductId())
}

func (p Position) Balance() float64 {
	return util.Float64(p.Account.Balance)
}

func (p Position) Hold() float64 {
	return util.Float64(p.Account.Hold)
}

func (p Position) Value() float64 {
	return p.Price() * p.Balance()
}

func (p Position) Price() float64 {
	return util.Float64(p.Ticker.Price)
}

func NewUsdPosition(account cb.Account) *Position {
	return NewPosition(account, cb.Ticker{}, nil)
}

func NewPosition(account cb.Account, ticker cb.Ticker, fills []cb.Fill) *Position {

	p := new(Position)
	p.Account = account
	p.Ticker = ticker

	for _, fill := range fills {
		if fill.Side == "buy" {
			p.buys = append(p.buys, *NewTrade(fill))
		} else {
			p.sells = append(p.sells, *NewTrade(fill))
		}
	}

	return p
}

func (p *Position) Trading() []Trade {

	if p.Hold() == p.Balance() {
		return nil
	}

	buys := p.buys
	sort.SliceStable(buys, func(i, j int) bool {
		return buys[i].CreatedAt.Time().After(buys[j].CreatedAt.Time())
	})

	var trading []Trade
	hold := p.Hold()
	for _, trade := range buys {
		if hold >= p.Balance() {
			break
		}
		trading = append(trading, trade)
		hold += trade.Size()
	}

	sort.SliceStable(trading, func(i, j int) bool {
		return trading[i].Price() > trading[j].Price()
	})

	return trading
}

func (p *Position) Log() {

	log.Info().
		Str(util.Dollar, util.Money(p.Price())).
		Str("bal", fmt.Sprintf(`%.3f`, p.Balance())).
		Str("hld", fmt.Sprintf(`%.3f`, p.Hold())).
		Str(util.Sigma, util.Usd(p.Value())).
		Msg(p.Currency)

	totalResult := 0.0
	for i, trade := range p.Trading() {

		gain := p.Product.GainPrice(trade.Price())
		result := (gain - (gain * .005)) * trade.Size()

		totalResult += result

		log.Warn().
			Str(util.Dollar, fmt.Sprintf("%.3f", trade.Price())).
			Str(util.Quantity, fmt.Sprintf("%.0f", trade.Size())).
			Str(util.Sigma, fmt.Sprintf("%.3f", trade.Total())).
			Time("🗓", trade.CreatedAt.Time()).
			Str("🎯", fmt.Sprintf("%.3f", gain)).
			Str("💰", fmt.Sprintf("%.3f", result)).
			Send()

		if i == len(p.Trading())-1 {
			log.Warn().Str("💰", fmt.Sprintf("%.3f", totalResult)).Send()
		}
	}

}
