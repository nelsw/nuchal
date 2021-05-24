package trade

import (
	"github.com/nelsw/nuchal/pkg/util"
	"testing"
)

func TestNew(t *testing.T) {

	if err := New(util.GuestName, []string{"ADA", "MATIC", "XTZ"}); err != nil {
		t.Error(err)
	}

}
