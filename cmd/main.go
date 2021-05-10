package main

import (
	. "nchl/pkg"
	"os"
)

const (
	target     = "tgt"
	trade      = "trade"
	simulation = "sim"
	charts     = "cht"
	history    = "hst"
	user       = "user"
)

func main() {

	domain := os.Args[1]

	if domain == user {
		CreateUser()
		return
	}

	SetupTarget()
	SetupUser()
	SetupClientConfig()

	switch domain {

	case history:
		SetupRates()

	case charts:
		SetupRates()
		CreateSim()
		CreateCharts()

	case simulation:
		SetupRates()
		CreateSim()
		CreateCharts()

	case trade:
		CreateTrades()
	}

}
