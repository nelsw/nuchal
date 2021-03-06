/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var trueIgnoreCase = regexp.MustCompile("(?i)true\\b")

func IsEnvVarTrue(key string) bool {
	return trueIgnoreCase.MatchString(os.Getenv(key))
}

func Float64(s string) float64 {
	if s == "" {
		return 0.0
	}
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		log.Error().Err(err).Send()
		return 0.0
	} else {
		return f
	}
}

func Usd(f float64) string {
	if f == 0 {
		return "$0.00"
	}
	return "$" + Money(f)
}

func Money(f float64) string {
	x := (f * 100) + 0.5
	x = x / 100
	rounded := fmt.Sprintf("%.3f", x)
	chunks := strings.Split(rounded, `.`)
	dollars := chunks[0]
	var cents string
	if len(chunks) > 1 {
		cents = chunks[1]
	}

	isNegative := strings.Contains(dollars, "-")
	if isNegative {
		chunks = strings.Split(dollars, "-")
		dollars = chunks[1]
	}

	places := len(dollars)

	if places < 4 {
		if isNegative {
			return fmt.Sprintf("-%s.%s", dollars, cents)
		}
		return fmt.Sprintf("%s.%s", dollars, cents)
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
	if isNegative {
		return fmt.Sprintf("-%s.%s", rounded, cents)
	}

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

func IsZero(s string) bool {
	return Float64(s) == 0.0
}

func GetCurrency(productID string) string {
	return fmt.Sprintf("%5s", strings.Split(productID, "-")[0])
}

func CbUrl(productID string) string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, productID)
}

func MakePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func CbpUrl(productID string) string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, productID)
}
