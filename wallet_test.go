package korbit

import (
	"testing"
)

func TestGetWallets(t *testing.T) {
	wallets, err := api.GetWallets()
	if err != nil {
		t.Error(err)
	}

	for _, v := range currencies {
		_, _ = api.GetCoinBalance(v, wallets)
	}
}
