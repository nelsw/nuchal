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
)

func main() {

	SetupTarget()
	SetupUser()
	SetupClientConfig()

	switch os.Args[1] {

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
