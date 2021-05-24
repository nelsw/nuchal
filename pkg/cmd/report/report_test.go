package report

import (
	"nuchal/pkg/util"
	"testing"
)

func TestNew(t *testing.T) {

	if err := New(util.GuestName, false, false); err != nil {
		t.Error(err)
	}

}
