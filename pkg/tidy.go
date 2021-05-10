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

		if float(account.Available) == 0 || account.Currency != "ZRX" {
			continue
		}

		var events []Transaction
		for _, entry := range GetLedgers(username, account.ID) {
			if float(entry.Balance) == 0 {
				break
			}
			for _, f := range FindFillsByOrderId(username, entry.Details.OrderID) {
				events = append(events, Transaction{entry, f})
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
			} else {
				key := math.Abs(amt) + bal
				if _, ok := eventMap[key]; ok {
					delete(eventMap, key)
				} else {
					fmt.Println("fuck")
				}
			}
		}

		for _, v := range eventMap {
			fmt.Println(pretty(v))
			// calculate entry order
		}
	}
	fmt.Println(username, "created entry orders")
}
