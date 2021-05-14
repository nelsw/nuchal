package sim

import (
	"github.com/rs/zerolog/log"
	"testing"
)

func TestNew(t *testing.T) {
	if err := New(); err != nil {
		log.Error().Err(err)
		panic(err)
	}
}
