package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Target struct {
	ProductId string `json:"product_id" gorm:"primaryKey"`
	Tweezer   float64
	Gain      float64
	Loss      float64
}

func (t Target) Json() string {
	b, _ := json.Marshal(&t)
	return string(b)
}

var (
	target  Target
	targets []Target
)

const toUSD = "-USD"

func init() {
	if err := db.AutoMigrate(target); err != nil {
		panic(err)
	}
}

func SetupTarget() {
	SetTarget(os.Args[2])
}

func SetTarget(symbol string) {
	productId := strings.ToUpper(symbol) + toUSD
	fmt.Println("setting up target for", productId)
	db.Where(query, productId).First(&target)
	if target == (Target{}) {
		fmt.Println("target not found")
		target = Target{productId, 0.0001, 0.0195, 0.35}
		db.Save(&target)
		fmt.Println("created target")
	}
	fmt.Println("setup target", Print(&target))
	fmt.Println()
	setupTargets()
}

func setupTargets() {
	fmt.Println("setting up targets")

	targets = append(targets, target)

	//for g := 0.01; g <= 0.2; g+=0.001 {
	//	for l := 0.01; l <= 0.6; l+=0.001 {
	//		targets = append(targets, Target{
	//			ProductId: target.ProductId,
	//			Tweezer:   .0001,
	//			Gain:      g,
	//			Loss:      l,
	//		})
	//	}
	//}

	fmt.Println("setup targets")
	fmt.Println()
}

func Size() string {

	switch target.ProductId {

	// 10
	case "MATIC" + toUSD:
		fallthrough
	case "BAT" + toUSD:
		return "10"

	//5
	case "EOS" + toUSD:
		fallthrough
	case "XTZ" + toUSD:
		fallthrough
	case "RLC" + toUSD:
		fallthrough
	case "MANA" + toUSD:
		return "5"

	//1
	case "ETC" + toUSD:
		fallthrough
	case "CTSI" + toUSD:
		fallthrough
	default:
		return "1"
	}
}
