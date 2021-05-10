package pkg

import (
	"fmt"
	"nchl/pkg/util"
	"testing"
)

func TestExitLossOrder(t *testing.T) {

	//name := "Connor Ross Van Elswyk"
	name := "Bricen Lee Miller"
	symbol := "SKL"

	SetTarget(symbol)
	SetUser(name)
	SetupClientConfig()

	//size := "62.8"
	//price := "62.289"
	//if exit, err := createEntryOrder(size, price); err != nil {
	//	fmt.Println("error creating exit order", err)
	//} else {
	//	fmt.Println("successfully created exit order", exit)
	//}

	orders := GetOrders()
	//fills := GetFills()
	//fmt.Println("fills", Print(fills))
	fmt.Println("orders", util.Print(orders))
	//
	//// 750ee80a-09a4-405e-8907-0ae3d9b10bcd
	//
	//cancelOrder("c14e8dfa-9c55-414c-ba2c-885a9e6f8bb6")
	//cancelOrder("fd66b742-3875-4f4c-be37-1e1f7ca11cc7")
}
