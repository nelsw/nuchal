package sim

import (
	"testing"
)

func TestNew(t *testing.T) {
	if err := New(false); err != nil {
		t.Error(err)
	}
}
