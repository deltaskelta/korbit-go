package korbit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Prices models the JSON that is returned by the  API
type Prices struct {
	Timestamp    *time.Time
	TimestampStr int64   `json:"timestamp"`
	Last         int64   `json:"last,string"`
	Bid          int64   `json:"bid,string"`
	Ask          int64   `json:"ask,string"`
	Low          int64   `json:"low,string"`
	High         int64   `json:"high,string"`
	Volume       float64 `json:"volume,string"`
}

// GetPrices hits the  server to get the current prices
func (k *API) GetPrices(coin string) (*Prices, error) {
	// korbitCalls are the abbreviations for the coins on korbit
	URL := fmt.Sprintf("https://api.korbit.co.kr/v1/ticker/detailed?currency_pair=%s", coin)

	req, err := k.NewRequest(URL, "GET", nil)
	if err != nil {
		return nil, errors.Wrap(err, "getting korbit prices")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "get korbit at %s", coin)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	var prices Prices
	err = json.NewDecoder(resp.Body).Decode(&prices)
	if err != nil {
		return nil, errors.Wrapf(err, "decoding korbit bytes at %s", coin)
	}

	return &prices, nil
}

// OrderbookResp contains all the bid and ask orders on the books.
type OrderbookResp struct {
	Timestamp int64      `json:"timestamp"`
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
}

// Orderbook is for returning the orderbook to the user in a form without strings.
type Orderbook struct {
	Timestamp int64
	Asks      []OrderbookOrder
	Bids      []OrderbookOrder
}

// OrderbookOrder is what is in the slice of orders from the orderbook.
type OrderbookOrder struct {
	Price int64
	Qty   float64
}

// Transform is what turned the korbit response into something usable, the korbit api had
// them as lists which each index had a specific meaning, these are better expressed as
// key-value pairs.
func (o *OrderbookResp) Transform() (*Orderbook, error) {
	ret := Orderbook{Timestamp: o.Timestamp}

	for _, v := range o.Bids {
		price, err := strconv.ParseInt(v[0], 10, 64)
		if err != nil {
			return nil, errors.New("error changing bid price to int64")
		}

		qty, err := strconv.ParseFloat(v[1], 10)
		if err != nil {
			return nil, errors.New("error changing bid qty to float64")
		}

		bid := OrderbookOrder{
			Price: price,
			Qty:   qty,
		}

		ret.Bids = append(ret.Bids, bid)
	}

	for _, v := range o.Asks {
		price, err := strconv.ParseInt(v[0], 10, 64)
		if err != nil {
			return nil, errors.New("error changing ask price to int64")
		}

		qty, err := strconv.ParseFloat(v[1], 10)
		if err != nil {
			return nil, errors.New("error changing ask qty to float64")
		}

		ask := OrderbookOrder{
			Price: price,
			Qty:   qty,
		}

		ret.Asks = append(ret.Asks, ask)
	}

	return &ret, nil
}

// GetOrderbook fetches the orderbook for the given coin.
func (k *API) GetOrderbook(coin string) (*Orderbook, error) {
	url := fmt.Sprintf("%s?currency_pair=%s", GetOrderbook, coin)

	req, err := k.NewRequest(url, "GET", nil)
	if err != nil {
		return nil, errors.Wrap(err, "korbit get orderbook")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "korbit orderbook fetch")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	var OBResp OrderbookResp
	err = json.NewDecoder(resp.Body).Decode(&OBResp)
	if err != nil {
		return nil, errors.Wrap(err, "korbit orderbook json decode")
	}

	retResp, err := OBResp.Transform()
	if err != nil {
		return nil, errors.Wrap(err, "transforming OBresp")
	}

	return retResp, nil
}
