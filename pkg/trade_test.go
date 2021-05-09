package pkg

import (
	"fmt"
	"testing"
)

func TestExitLossOrder(t *testing.T) {

	SetTarget("CTSI")
	SetUser("Bricen Lee Miller")
	SetupClientConfig()

	size := "1.0"
	stopLoss := "1.59"
	if lossExit, err := createEntryOrder(size, stopLoss); err != nil {
		fmt.Println("error creating loss exit order", err)
	} else {
		fmt.Println("successfully created loss exit order", lossExit)
	}

	//orders := GetOrders()
	//fills := GetFills()
	//fmt.Println("fills", Print(fills))
	//fmt.Println("orders", Print(orders))
	//
	//// 750ee80a-09a4-405e-8907-0ae3d9b10bcd
	//
	//cancelOrder("a40f5169-b641-4bb7-9b3a-1f5f3155cbe9")
}
