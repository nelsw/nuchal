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

import "time"

type Period struct {
	Alpha time.Time `envconfig:"PERIOD_ALPHA" yaml:"alpha"`

	Omega time.Time `envconfig:"PERIOD_OMEGA" yaml:"omega"`

	Duration time.Duration `envconfig:"PERIOD_DURATION" yaml:"duration"`
}

// InRange is an exclusive range function to determine if the given time is in greater than alpha and lesser than omega.
func (p *Period) InRange(t time.Time) bool {
	return p.Alpha.Before(t) && p.Omega.After(t)
}

func (p *Period) IsTimeToExit() bool {
	return time.Now().After(p.Omega) || p.Alpha.Add(p.Duration).After(p.Omega)
}
