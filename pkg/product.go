package pkg

import "strings"

// todo - dynamic product prices from simulations
// todo - symbol type
const (
	INCH  = "1INCH-USD"
	AAVE  = "AAVE-USD"
	ADA   = "ADA-USD"
	ALGO  = "ALGO-USD"
	ANKR  = "ANKR-USD"
	ATOM  = "ATOM-USD"
	BAL   = "BAL-USD"
	BAND  = "BAND-USD"
	BAT   = "BAT-USD"
	BCH   = "BCH-USD"
	BNT   = "BNT-USD"
	BTC   = "BTC-USD"
	Celo  = "CELO-USD"
	CGLD  = "CGLD-USD"
	COMP  = "COMP-USD"
	CRV   = "CRV-USD"
	CTSI  = "CTSI-USD"
	DAI   = "DAI-USD"
	DASH  = "DASH-USD"
	ENJ   = "ENJ-USD"
	EOS   = "EOS-USD"
	ETC   = "ETC-USD"
	ETH   = "ETH-USD"
	FORTH = "FORTH-USD"
	FIL   = "FIL-USD"
	GRT   = "GRT-USD"
	KNC   = "KNC-USD"
	LINK  = "LINK-USD"
	LRC   = "LRC-USD"
	LTC   = "LTC-USD"
	MANA  = "MANA-USD"
	MATIC = "MATIC-USD"
	MIR   = "MIR-USD"
	MKR   = "MKR-USD"
	NMR   = "NMR-USD"
	NKN   = "NKN-USD"
	NU    = "NU-USD"
	OGN   = "OGN-USD"
	OMG   = "OMG-USD"
	OXT   = "OXT-USD"
	REN   = "REN-USD"
	REP   = "REP-USD"
	RLC   = "RLC-USD"
	SUSHI = "SUSHI-USD"
	SKL   = "SKL-USD"
	SNX   = "SNX-USD"
	STORJ = "STORJ-USD"
	TRB   = "TRB-USD"
	USDT  = "USDT-USD"
	UMA   = "UMA-USD"
	UNI   = "UNI-USD"
	WBTC  = "WBTC-USD"
	XLM   = "XLM-USD"
	XTZ   = "XTZ-USD"
	YFI   = "YFI-USD"
	ZEC   = "ZEC-USD"
	ZRX   = "ZRX-USD"
)

var ProductIds = []string{
	INCH, AAVE, ADA, ALGO, ANKR, ATOM, BAL, BAND, BAT, BCH, BNT, BTC, Celo, CGLD, COMP, CRV, CTSI, DAI, DASH, ENJ, EOS,
	ETC, ETH, FORTH, FIL, GRT, KNC, LINK, LRC, LTC, MANA, MATIC, MIR, MKR, NMR, NKN, NU, OGN, OMG, OXT, REN, REP, RLC,
	SUSHI, SKL, SNX, STORJ, TRB, USDT, UMA, UNI, WBTC, XLM, XTZ, YFI, ZEC, ZRX,
}

func ProductId(symbol *string) string {
	return strings.ToUpper(*symbol) + "-USD"
}

func Currency(productId string) string {
	return strings.Split(productId, "-USD")[0]
}

// todo - load prices to product
func size(price float64) string {
	if price < 1 {
		return "10"
	} else if price < 2 {
		return "5"
	} else {
		return "1"
	}
}
