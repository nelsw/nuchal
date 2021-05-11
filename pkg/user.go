package pkg

import (
	"encoding/json"
	"fmt"
)

type Portfolio struct {
	Username, Value string
	Positions       []Position
}

type Position struct {
	ProductId, Value string
	Balance          float64
}

// USD represents US dollar amount in terms of cents
type USD int64

// NewUSD creates a float64 to USD
// e.g. 1.23 to $1.23, 1.345 to $1.35
func NewUSD(f float64) USD {
	return USD((f * 100) + 0.5)
}

// Float64 converts a USD to float64
func (m USD) Float64() float64 {
	x := float64(m)
	x = x / 100
	return x
}

// String returns a formatted USD value
func (m USD) String() string {
	x := float64(m)
	x = x / 100
	return fmt.Sprintf("$%.2f", x)
}

func DisplayAccountInfo(user User) {

	var positions []Position
	var total float64
	for _, account := range GetAccounts(user) {

		if Float64(account.Balance) == 0.0 && Float64(account.Hold) == 0.0 {
			continue
		}

		productId := account.Currency + "-USD"
		balance := Float64(account.Balance)

		if account.Currency == "USD" {
			total += balance
			positions = append(positions, Position{
				ProductId: productId,
				Balance:   balance,
				Value:     NewUSD(balance).String(),
			})
			continue
		}

		price := Float64(GetTicker(user, productId))
		value := price * balance
		total += value
		positions = append(positions, Position{
			ProductId: productId,
			Balance:   balance,
			Value:     NewUSD(value).String(),
		})
	}

	portfolio := Portfolio{user.Name, NewUSD(total).String(), positions}
	b, _ := json.MarshalIndent(&portfolio, "", "  ")
	fmt.Println(string(b))
}
