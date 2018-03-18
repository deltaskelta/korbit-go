package korbit

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// Currency holds currency price information
type Currency struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value,string"`
}

// ConnectedAccount is used when querying korbit wallets.
type ConnectedAccount struct {
	Currency        string   `json:"currency"`
	Status          string   `json:"status"`
	RegisteredOwner string   `json:"registeredOwner"`
	Address         AcctInfo `json:"address"`
}

// AcctInfo is nested inside korbit wallets
type AcctInfo struct {
	Bank    string `json:"bank"`
	Account string `json:"account"`
	Owner   string `json:"address"`
}

// NonBtcWallet is for querying non btc currencies which are different than btc
// currencies.
type NonBtcWallet struct {
	Balance    []*Currency `json:"balance"`
	Tradable   []*Currency `json:"tradable"`
	TradeInUse []*Currency `json:"tradeInUse"`
}

// Wallet is for checking wallet status at .
type Wallet struct {
	In            []*ConnectedAccount `json:"in"`
	Out           []*ConnectedAccount `json:"out"`
	Balance       []*Currency         `json:"balance"`
	PendingOut    []*Currency         `json:"pendingOut"`
	PendingOrders []*Currency         `json:"pendingOrders"`
	Available     []*Currency         `json:"available"`
}

type Balance_ struct {
	Available       float64 `json:"available,string"`
	TradeInuse      float64 `json:"trade_in_use,string"`
	WithdrawalInUse float64 `json:"withdrawal_in_use,string"`
}

type Balances map[string]Balance_

// GetWallets gives back all of the wallets for a user.
func (k *API) GetBalances() (balances Balances, err error) {

	req, err := k.NewRequest(BalancesURL, "GET", nil)
	if err != nil {
		return nil, errors.Wrap(err, "making korbit wallet status request")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "getting korbit wallet status")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	err = json.NewDecoder(resp.Body).Decode(&balances)
	if err != nil {
		return nil, errors.Wrapf(err, "korbit wallet unmarshal failed")
	}

	return balances, nil
}
