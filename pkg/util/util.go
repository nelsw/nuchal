package util

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Fish     = `🐠`
	Quantity = `꠹`
	Dollar   = `$`
	Currency = `¤`
	Sigma    = `𝚺`
	Banner   = `

____________________________________________/\\\_________________________/\\\\\\____        
 ___________________________________________\/\\\________________________\////\\\____       
  ___________________________________________\/\\\___________________________\/\\\____      
   __/\\/\\\\\\____/\\\____/\\\_____/\\\\\\\\_\/\\\__________/\\\\\\\\\_______\/\\\____     
    _\/\\\////\\\__\/\\\___\/\\\___/\\\//////__\/\\\\\\\\\\__\////////\\\______\/\\\____    
     _\/\\\__\//\\\_\/\\\___\/\\\__/\\\_________\/\\\/////\\\___/\\\\\\\\\\_____\/\\\____   
      _\/\\\___\/\\\_\/\\\___\/\\\_\//\\\________\/\\\___\/\\\__/\\\/////\\\_____\/\\\____  
       _\/\\\___\/\\\_\//\\\\\\\\\___\///\\\\\\\\_\/\\\___\/\\\_\//\\\\\\\\/\\__/\\\\\\\\\_ 
        _\///____\///___\/////////______\////////__\///____\///___\////////\//__\/////////__

`
)

func Float64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		log.Error().Err(err).Send()
		return 0.0
	} else {
		return f
	}
}

func Round2Places(f float64) string {
	x := (f * 100) + 0.5
	x = x / 100
	return fmt.Sprintf("%.3f", x)
}

func Usd(f float64) string {
	rounded := Round2Places(f)
	chunks := strings.Split(rounded, `.`)
	dollars := chunks[0]
	cents := chunks[1]

	places := len(dollars)

	if places < 4 {
		return fmt.Sprintf("$%s.%s", dollars, cents)
	}

	pivot := places - 3
	var newFields []string
	for i, oldField := range dollars {
		if i == pivot {
			newFields = append(newFields, ",")
		}
		newFields = append(newFields, string(oldField))
	}
	rounded = strings.Join(newFields, ``)
	return fmt.Sprintf("$%s.%s", rounded, cents)
}

func Money(f float64) string {

	rounded := Round2Places(f)
	chunks := strings.Split(rounded, `.`)
	dollars := chunks[0]
	cents := chunks[1]

	places := len(dollars)

	if places < 4 {
		return fmt.Sprintf("$%s.%s", dollars, cents)
	}

	pivot := places - 3
	var newFields []string
	for i, oldField := range dollars {
		if i == pivot {
			newFields = append(newFields, ",")
		}
		newFields = append(newFields, string(oldField))
	}
	rounded = strings.Join(newFields, ``)
	return fmt.Sprintf("%s.%s", rounded, cents)
}

func FirstIntOrZero(arr []int) int {
	if arr != nil && len(arr) > 0 {
		return arr[0]
	}
	return 0
}

func IsInsufficientFunds(err error) bool {
	return err != nil && err.Error() == "Insufficient funds"
}

func DoIndefinitely(fun func()) error {
	exit := make(chan string)
	go fun()
	for {
		select {
		case <-exit:
			return nil
		}
	}
}

func IsZero(s string) bool {
	return Float64(s) == 0.0
}

func Sleep(d time.Duration) {
	exit := time.Now().Add(d)
	for {
		fmt.Println()
		time.Sleep(d)
		if time.Now().After(exit) {
			break
		}
		d = time.Duration(d.Nanoseconds() / 2)
	}
}

func LogBanner() {
	fmt.Println(Banner)
}

func PrintlnBanner() {
	fmt.Println(Banner)
}

func PrintCursor() {
	fmt.Print("> ")
}

func PrintNewLine() {
	fmt.Print("\n")
}

func PrintlnPrettyJson(v interface{}) {
	fmt.Println(PrettyJson(v))
}

func PrettyJson(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

func ConfigFromYml(v interface{}) error {

	log.Debug().Interface("ConfigFromYml", v).Send()

	f, err := os.Open("config.yml")
	if err != nil {
		log.Debug().Err(err).Send()
		return err
	}

	d := yaml.NewDecoder(f)
	if err := d.Decode(v); err != nil {
		log.Debug().Err(err).Send()
		return err
	}

	log.Debug().Interface("ConfigFromYml", v).Send()
	return nil
}

func ConfigFromEnv(v interface{}) error {

	log.Debug().Interface("ConfigFromEnv", v).Send()

	if err := envconfig.Process("", v); err != nil {
		log.Debug().Err(err).Send()
		return err
	}

	log.Debug().Interface("ConfigFromEnv", v).Send()
	return nil
}
