package user

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"nchl/pkg/coinbase"
	"nchl/pkg/model/usd"
	"nchl/pkg/util"
)

type Portfolio struct {
	Username, Value string
	Positions       []Position
}

type Position struct {
	ProductId, Balance, Value string
}

func DisplayAccountInfo(username string) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		panic(err)
	}
	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			fmt.Println(err)
		}
	}(wsConn)

	var positions []Position

	var total float64
	for _, account := range coinbase.GetAccounts(username) {

		if util.IsZero(account.Balance) && util.IsZero(account.Hold) {
			continue
		}

		productId := account.Currency + "-USD"
		balance := util.Float64(account.Balance)

		if account.Currency == "USD" {
			total += balance
			positions = append(positions, Position{
				ProductId: productId,
				Balance:   usd.NewUSD(balance).String(),
				Value:     usd.NewUSD(balance).String(),
			})
			continue
		}

		price := util.Float64(coinbase.GetTicker(username, productId))
		value := price * balance
		total += value
		positions = append(positions, Position{
			ProductId: productId,
			Balance:   usd.NewUSD(balance).String(),
			Value:     usd.NewUSD(value).String(),
		})
	}

	portfolio := Portfolio{username, usd.NewUSD(total).String(), positions}

	fmt.Println(util.Pretty(portfolio))

}
