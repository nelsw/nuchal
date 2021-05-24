package sim

import (
	"nuchal/pkg/util"
	"testing"
)

func TestNew(t *testing.T) {
	if err := New(util.GuestName, false); err != nil {
		t.Error(err)
	}
}
