package korbit

import (
	"log"
	"os"
	"testing"
)

var api *API

func TestMain(m *testing.M) {
	api = NewKorbitAPI(
		clientID,
		clientSecret,
		username,
		password,
	)

	err := api.Login()
	if err != nil {
		log.Println("error logging in, you may need to provide credentials.")
	}

	retCode := m.Run()
	os.Exit(retCode)
}

func TestGetTransactions(t *testing.T) {
	_, err := api.GetTransactionHistory(BTCKRW, "fills", "", "10000", "")
	if err != nil {
		t.Error(err)
	}
}

func TestBuy(t *testing.T) {
	buy := OrderArgs{
		CoinAmount:   "0.001",
		CurrencyPair: BTCKRW,
		Price:        3000000,
		Type:         "limit",
	}

	resp, err := api.Buy(&buy)
	if err != nil {
		t.Error(err)
	}

	cancelResp, err := api.CancelOpenOrders([]int64{resp.OrderID}, BTCKRW)
	if err != nil {
		t.Error(err)
	}

	if cancelResp[0].Status != "success" {
		t.Error("cancel not successful")
	}

}

func TestSell(t *testing.T) {
	sell := OrderArgs{
		CoinAmount:   "0.001",
		CurrencyPair: BTCKRW,
		Price:        3000000,
		Type:         "limit",
	}

	resp, err := api.Sell(&sell)
	if err != nil {
		t.Error(err)
	}

	cancelResp, err := api.CancelOpenOrders([]int64{resp.OrderID}, BTCKRW)
	if err != nil {
		t.Error(err)
	}

	if cancelResp[0].Status != "success" {
		t.Error("cancel not successful")
	}

}

func TestListOrders(t *testing.T) {
	_, err := api.ListOpenOrders(BTCKRW)
	if err != nil {
		t.Error(err)
	}
}
