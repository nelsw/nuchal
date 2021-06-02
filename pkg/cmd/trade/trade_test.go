package trade

import (
	"testing"
)

func TestNew(t *testing.T) {

	//if err := New(); err != nil {
	//	t.Error(err)
	//}

}

func TestNewHolds(t *testing.T) {

	if err := NewHolds(); err != nil {
		t.Error(err)
	}

}
