package trade

import "testing"

func TestNew(t *testing.T) {

	if err := New(); err != nil {
		t.Error(err)
	}

}
