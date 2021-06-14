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

package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gopkg.in/yaml.v2"
	"os"
	"strconv"
	"time"
)

// period is a range of time representing when to start and stop executing the trade command.
type period struct {

	// Alpha defines when command functionality should start.
	Alpha *time.Time `envconfig:"PERIOD_ALPHA" yaml:"alpha"`

	// Omega defines when command functionality should cease.
	Omega *time.Time `envconfig:"PERIOD_OMEGA" yaml:"omega"`

	// Duration is the amount of time the command should be available.
	// sim uses this as the amount of time to host result pages.
	// trade uses this to override Alpha and Omega values.
	Duration *time.Duration `envconfig:"PERIOD_DURATION" yaml:"duration"`

	start *time.Time
}

func NewPeriod(name, duration string, now *time.Time) *period {

	type periodConfig struct {
		Period period `yaml:"period"`
	}

	c := new(periodConfig)
	c.Period.start = now

	var err error

	// first check the environment
	err = envconfig.Process("", c)
	if err != nil || !c.Period.isValid() {
		// second check the config file
		var f *os.File
		if f, err = os.Open(name); err == nil {
			err = yaml.NewDecoder(f).Decode(&c)
		}
	}

	// we're either of these configuration sources successful?
	if err == nil && c.Period.isValid() {
		// set a duration in case it doesn't have one
		if c.Period.Duration == nil {
			duration := c.Period.Omega.Sub(*c.Period.Alpha)
			c.Period.Duration = &duration
		}
		return &c.Period
	}

	// third, check the given duration
	if dur, err := time.ParseDuration(duration); err == nil {
		alpha := now.Add(-dur)
		p := new(period)
		p.Alpha = &alpha
		p.Omega = now
		p.Duration = &dur
		return p
	}

	// last, create a new default period from the started time
	y := now.Year()
	m := strconv.Itoa(int(now.Month()))
	d := strconv.Itoa(now.Day())

	if int(now.Month()) < 10 {
		m = "0" + m
	}
	if now.Day() < 10 {
		d = "0" + d
	}

	alpha, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT12:00:00+00:00", y, m, d))
	omega, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT22:00:00+00:00", y, m, d))
	dur := omega.Sub(alpha)

	p := new(period)
	p.Alpha = &alpha
	p.Omega = &omega
	p.Duration = &dur

	return p
}

func (p *period) isValid() bool {
	return &p != nil &&
		p.Alpha != nil &&
		p.Alpha.Year() != 1 &&
		p.Omega != nil &&
		p.Omega.Year() != 1
}

// InPeriod is an exclusive range function to determine if the given time falls within the defined period.
func (p *period) InPeriod(t time.Time) bool {
	return p.Alpha.Before(t) && p.Omega.After(t)
}

func (p *period) Start() *time.Time {
	return p.Alpha
}

func (p *period) Stop() *time.Time {
	return p.Omega
}

func (p period) RateParams() *[]cb.GetHistoricRatesParams {
	start := *p.Alpha
	end := start.Add(time.Hour * 4)
	var results []cb.GetHistoricRatesParams
	for i := 0; i < 24; i += 4 {
		results = append(results, cb.GetHistoricRatesParams{start, end, 60})
		start = end
		end = end.Add(time.Hour * 4)
		if end.After(*p.Omega) {
			break
		}
	}
	return &results
}
