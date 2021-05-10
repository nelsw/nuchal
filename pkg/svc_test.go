package pkg

import (
	"fmt"
	"nchl/pkg/util"
	"testing"
)

func TestGetAccountInfo(t *testing.T) {

	SetUser("Bryce")
	//SetUser("Bricen")
	//SetUser("Connor")
	SetupClientConfig()
	SetTarget("ZRX")
	printCashBalance()
	printPositions()

	orders := GetOrders()
	for _, order := range orders {
		fmt.Println(util.Print(order))
	}

	//o, err := createEntryOrder("4.06", "9.89")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(util.Print(o))

	fills := GetFills()
	for _, fill := range fills {
		fmt.Println(util.Print(fill))
	}
}
