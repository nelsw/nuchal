package model

import (
	"fmt"
	"testing"
)

func TestNewCoinbaseApiFromEnv(t *testing.T) {
	c, err := NewCoinbaseApi()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(c)
}
