package config

import "time"

type Period struct {

	// Alpha is a RFC3339 parsable time.Time value for defining when to begin a period of time.
	// e.g. "2021-05-30T14:00:00+00:00"
	Alpha string `envconfig:"PERIOD_ALPHA" yaml:"alpha"`

	// Omega is a RFC3339 parsable time.Time value for defining when to end a period of time.
	// e.g. "2021-05-30T22:59:59+00:00"
	Omega string `envconfig:"PERIOD_OMEGA" yaml:"omega"`
}

// InRange is an exclusive range function to determine if the given time is in greater than alpha and lesser than omega.
func (p *Period) InRange(t time.Time) bool {
	return p.Start().Before(t) && p.End().After(t)
}

func (p Period) Start() *time.Time {
	a, err := time.Parse(time.RFC3339, p.Alpha)
	if err != nil {
		a = p.End().Add(time.Hour * -6)
	}
	return &a
}

func (p Period) End() *time.Time {
	o, err := time.Parse(time.RFC3339, p.Omega)
	if err != nil {
		o = time.Now()
	}
	return &o
}
