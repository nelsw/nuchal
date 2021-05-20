package account

import "testing"

func TestNew(t *testing.T) {

	if err := New(false, false); err != nil {
		t.Error(err)
	}

}
