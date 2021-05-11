package pkg

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"math"
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

	for _, account := range GetAccounts(username) {

		if float(account.Available) == 0 || float(account.Balance) == 0 || account.Currency == "USD" {
			continue
		}

		var events []Transaction
		for _, entry := range GetLedgers(username, account.ID) {
			if float(entry.Amount) == 0 || float(entry.Balance) == 0 {
				break
			}
			fmt.Println(pretty(entry))
			for _, fill := range GetFills(username, entry.Details.OrderID) {
				events = append(events, Transaction{entry, fill})
			}
		}

		sort.SliceStable(events, func(i, j int) bool {
			return events[i].time().Before(events[j].time())
		})

		eventMap := map[float64]Transaction{}
		for _, event := range events {
			bal := float(event.Balance)
			amt := float(event.Amount)
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
			price := float(event.Price) + (float(event.Price) * stopGain)
			_, _ = CreateOrder(username, NewStopEntryOrder(event.ProductID, event.Size, price))
		}
	}
	fmt.Println(username, "created entry orders")
}
