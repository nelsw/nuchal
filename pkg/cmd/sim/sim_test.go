package sim

import (
	"testing"
)

func TestNew(t *testing.T) {
	if err := New("Carl Brutanandilewski", true); err != nil {
		t.Error(err)
	}
}
