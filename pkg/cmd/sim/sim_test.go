package sim

import (
	"testing"
)

func TestNew(t *testing.T) {
	if err := New("Carl Brutanandilewski", false); err != nil {
		t.Error(err)
	}
}
