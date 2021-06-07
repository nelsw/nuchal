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
	"strconv"
	"time"
)

// Period handles time.
type Period struct {

	// Alpha defines when command functionality should start.
	Alpha time.Time `envconfig:"PERIOD_ALPHA" yaml:"alpha"`

	// Omega defines when command functionality should cease.
	Omega time.Time `envconfig:"PERIOD_OMEGA" yaml:"omega"`

	// Duration is the amount of time the command should be available.
	// sim uses this as the amount of time to host result pages.
	// trade uses this to override Alpha and Omega values.
	Duration time.Duration `envconfig:"PERIOD_DURATION" yaml:"duration"`

	started *time.Time
}

// InPeriod is an exclusive range function to determine if the given time falls within the defined period.
func (p *Period) InPeriod(t time.Time) bool {
	return p.Start().Before(t) && p.Stop().After(t)
}

// Start returns the configured Start time. If no time is configured, Start returns today at noon UTC.
func (p *Period) Start() *time.Time {
	then := p.Alpha
	if then.Year() == 1 {
		then, _ = time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT12:00:00+00:00", year(), month(), day()))
	}
	return &then
}

// Stop returns the configured Stop time. If no time is configured, Stop returns today at 10pm UTC.
func (p *Period) Stop() *time.Time {
	then := p.Omega
	if then.Year() == 1 {
		then, _ = time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT22:00:00+00:00", year(), month(), day()))
	}
	return &then
}

func year() int {
	return time.Now().Year()
}

func month() string {
	m := int(time.Now().Month())
	s := strconv.Itoa(m)
	if m < 10 {
		return "0" + s
	}
	return s
}

func day() string {
	d := time.Now().Day()
	s := strconv.Itoa(d)
	if d < 10 {
		return "0" + s
	}
	return s
}
