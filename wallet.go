package korbit

import (
	"encoding/json"
	"fmt"
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

// Wallets is for getting the wallet balances for all coins in korbit
type Wallets struct {
	BTC *Wallet
	ETC *NonBtcWallet
	ETH *NonBtcWallet
}

// GetWallets gives back all of the wallets for a user.
func (k *API) GetWallets() (*Wallets, error) {

	var wallets Wallets

	for _, v := range currencies {
		url := fmt.Sprintf("%s?currency_pair=%s", WalletStatus, v)
		req, err := k.NewRequest(url, "GET", nil)
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

		switch v {
		case BTCKRW:
			var wallet Wallet
			err = json.NewDecoder(resp.Body).Decode(&wallet)
			wallets.BTC = &wallet
		case ETHKRW:
			var wallet NonBtcWallet
			err = json.NewDecoder(resp.Body).Decode(&wallet)
			wallets.ETH = &wallet
		case ETCKRW:
			var wallet NonBtcWallet
			err = json.NewDecoder(resp.Body).Decode(&wallet)
			wallets.ETC = &wallet
		default:
			return nil, errors.New("received and unknown coin type")
		}

		if err != nil {
			return nil, errors.Wrapf(err, "korbit wallet unmarshal at: %s", v)
		}
	}

	return &wallets, nil
}

// GetCoinBalance goes through the wallet specified with coin and returns the balance in
// krw and in the currency of coin
func (k *API) GetCoinBalance(coin string, wallets *Wallets) (coins, krw float64) {
	switch coin {
	case BTCKRW:
		for _, v := range wallets.BTC.Balance {
			switch v.Currency {
			case BTC:
				coins = v.Value
			case KRW:
				krw = v.Value
			}
		}
	case ETHKRW:
		for _, v := range wallets.ETH.Balance {
			switch v.Currency {
			case ETH:
				coins = v.Value
			case KRW:
				krw = v.Value
			}
		}
	case ETCKRW:
		for _, v := range wallets.ETC.Balance {
			switch v.Currency {
			case ETC:
				coins = v.Value
			case KRW:
				krw = v.Value
			}
		}
	}
	return coins, krw
}
