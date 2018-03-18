package korbit

import (
	"testing"
)

func TestGetWallets(t *testing.T) {
	balances, err := api.GetBalances()
	if err != nil {
		t.Error(err)
	}
	println(balances)
}
