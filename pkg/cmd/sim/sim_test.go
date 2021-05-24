package sim

import (
	"github.com/nelsw/nuchal/pkg/util"
	"testing"
)

func TestNew(t *testing.T) {
	if err := New(util.GuestName, false); err != nil {
		t.Error(err)
	}
}
