package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Target struct {
	ProductId string `json:"product_id" gorm:"primaryKey"`
	Tweezer float64
	Gain    float64
	Loss    float64
}

func (t Target) Print(pretty ...bool) string {
	b, _ := json.Marshal(&t)
	if len(pretty) > 0 && pretty[0] {
		var bb bytes.Buffer
		_ = json.Indent(&bb, b, "", "  ")
	}
	return string(b)
}

var (
	target Target
	targets []Target
)

const toUSD = "-USD"

func init() {
	if err := db.AutoMigrate(target); err != nil {
		panic(err)
	}
}

func SetupTarget() {
	productId := strings.ToUpper(os.Args[2]) + toUSD
	fmt.Println("setting up target for productId", productId)
	db.Where(query, productId).First(&target)
	if target == (Target{}) {
		fmt.Println("target not found")
		target = Target{productId, Float(os.Args[3]), Float(os.Args[4]), Float(os.Args[5])}
		db.Save(&target)
		fmt.Println("created target", &target)
	}
	fmt.Println("setup target", target)
	fmt.Println()
	setupTargets()
}

func setupTargets() {
	fmt.Println("setting up targets")

	for g := 0.5; g <= 1.5; g+=0.1 {
		for l := 0.5; l <= 1.5; l+=0.1 {
			targets = append(targets, Target{
				ProductId: target.ProductId,
				Tweezer:   .0001,
				Gain:      target.Gain * g,
				Loss:      target.Loss * l,
			})
		}
	}

	fmt.Println("setup targets")
	fmt.Println()
}

func Size() string {
	switch target.ProductId {
	case "ETC" + toUSD:
		return "1"
	case "EOS" + toUSD:
		return "5"
	case "XTZ" + toUSD:
		return "5"
	case "MATIC" + toUSD:
		return "10"
	case "MANA" + toUSD:
		return "5"
	case "BAT" + toUSD:
		return "10"
	default:
		return "1"
	}
}
