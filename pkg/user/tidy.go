package user

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"math"
	"nchl/pkg/coinbase"
	order2 "nchl/pkg/model/order"
	"nchl/pkg/model/product"
	"nchl/pkg/util"
	"sort"
	"time"
)

type Transaction struct {
	cb.LedgerEntry
	cb.Fill
}

func (t Transaction) time() time.Time {
	return t.Fill.CreatedAt.Time()
}

// CreateEntryOrders creates entry orders available product in user accounts. This is a very tricky process.
func CreateEntryOrders(username string) {

	fmt.Println(username, "creating entry orders")

	for _, account := range coinbase.GetAccounts(username) {

		if util.Float64(account.Available) == 0 || util.Float64(account.Balance) == 0 || account.Currency == "USD" {
			continue
		}

		var events []Transaction
		for _, entry := range coinbase.GetLedgers(username, account.ID) {
			if util.Float64(entry.Amount) == 0 || util.Float64(entry.Balance) == 0 {
				break
			}
			fmt.Println(util.Pretty(entry))
			for _, fill := range coinbase.GetFills(username, entry.Details.OrderID) {
				events = append(events, Transaction{entry, fill})
			}
		}

		sort.SliceStable(events, func(i, j int) bool {
			return events[i].time().Before(events[j].time())
		})

		eventMap := map[float64]Transaction{}
		for _, event := range events {
			bal := util.Float64(event.Balance)
			amt := util.Float64(event.Amount)
			if amt > 0 {
				eventMap[bal] = event
				continue
			}
			key := math.Abs(amt) + bal
			if _, ok := eventMap[key]; ok {
				delete(eventMap, key)
			}
		}

		for _, event := range eventMap {
			price := product.PricePlusStopGain(event.ProductID, util.Float64(event.Price))
			_ = coinbase.CreateOrder(username, order2.NewStopEntryOrder(event.ProductID, event.Size, price))
		}
	}
	fmt.Println(username, "created entry orders")
}
