package trade

import "testing"

func TestNew(t *testing.T) {

	if err := New("Carl Brutanandilewski", []string{"ADA", "MATIC", "XTZ"}); err != nil {
		t.Error(err)
	}

}
