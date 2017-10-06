package korbit

import (
	"testing"
)

func TestGetPrices(t *testing.T) {
	for _, v := range currencies {
		_, err := api.GetPrices(v)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGetOrderbook(t *testing.T) {
	for _, v := range currencies {
		_, err := api.GetOrderbook(v)
		if err != nil {
			t.Error(err)
		}
	}
}
