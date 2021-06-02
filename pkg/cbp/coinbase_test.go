package cbp

import (
	"github.com/nelsw/nuchal/pkg/util"
	"testing"
)

func TestNewCoinbaseApiFromEnv(t *testing.T) {

	if api, err := NewApi(); err != nil {
		t.Error(err)
	} else if err := api.validate(); err != nil {
		t.Error(err)
	}
}

func TestApi_GetProducts(t *testing.T) {
	if api, err := NewApi(); err != nil {
		t.Error(err)
	} else if err := api.validate(); err != nil {
		t.Error(err)
	} else if p, err := api.GetProducts(); err != nil {
		t.Error(err)
	} else {
		util.PrintlnPrettyJson(p)
	}
}
