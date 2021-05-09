package main

import (
	. "nchl/pkg"
	"os"
)

const (
	target = "tgt"
	trade = "trade"
	simulation = "sim"
	charts = "cht"
)

func main() {

	SetupTarget()
	SetupUser()
	SetupClientConfig()

	switch os.Args[1] {

	case charts:
		SetupRates()
		CreateSim()
		CreateCharts()

	case simulation:
		SetupRates()
		CreateSim()

	case trade:
		SetupWebsocket()
		CreateTrades()
	}

}
