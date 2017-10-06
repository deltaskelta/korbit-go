package korbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// OrderArgs are the arguments for making a korbit order.
type OrderArgs struct {
	CurrencyPair string // PAIR_krw
	Type         string // limit or market
	Price        int64  // price in KRW
	CoinAmount   string // always the price in terms of KRW
	FiatAmount   string // only for placing market order
}

// OrderResponse is the response that is given to a korbit order.
type OrderResponse struct {
	OrderID      int64  `json:"orderId"`
	Status       string `json:"status"`
	CurrencyPair string `json:"currency_pair"`
	Side         string `json:"side"`
	Price        int64  `json:"price"`
	Type         string `json:"type"`
}

// Buy takes care of placing a bid order with the given arguments.
func (k *API) Buy(order *OrderArgs) (*OrderResponse, error) {

	if order.Type != Limit && order.Type != Market {
		return nil, errors.New("unrecognized order type")
	}

	data := url.Values{
		"nonce":         {k.GetNonce()},
		"currency_pair": {order.CurrencyPair},
		"type":          {order.Type},
		"price":         {fmt.Sprintf("%d", order.Price)},
		"coin_amount":   {order.CoinAmount},
		"fiat_amount":   {order.FiatAmount},
	}

	req, err := k.NewRequest(PlaceBid, "POST", data)
	if err != nil {
		return nil, errors.Wrap(err, "make korbit order request")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "placing korbit bid order")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading korbit response")
	}

	var orderResp OrderResponse
	err = json.Unmarshal(respBytes, &orderResp)
	if err != nil {
		log.Println(string(respBytes))
		log.Println(resp.Header)
		return nil, errors.Wrapf(err, "unmarhsal place korbit bid, resp code: %d", resp.StatusCode)
	}

	orderResp.Price = order.Price
	orderResp.Side = "buy"
	orderResp.Type = "limit"

	if orderResp.Status != "success" {
		return &orderResp, errors.Errorf("order not successful: %s", orderResp.Status)
	}

	return &orderResp, nil
}

// Sell takes care of placing korbit ask orders to the orderbook.
func (k *API) Sell(order *OrderArgs) (*OrderResponse, error) {

	if order.Type != Limit && order.Type != Market {
		return nil, errors.New("unrecognized order type")
	}

	data := url.Values{
		"currency_pair": {order.CurrencyPair},
		"type":          {order.Type},
		"price":         {fmt.Sprintf("%d", order.Price)},
		"coin_amount":   {order.CoinAmount},
		"nonce":         {k.GetNonce()},
	}

	req, err := k.NewRequest(PlaceAsk, "POST", data)
	if err != nil {
		return nil, errors.Wrap(err, "make korbit order request")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "placing korbit ask order")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	var orderResp OrderResponse
	err = json.NewDecoder(resp.Body).Decode(&orderResp)
	if err != nil {
		return nil, errors.Wrap(err, "json decode korbit ask")
	}

	orderResp.Price = order.Price
	orderResp.Side = "sell"
	orderResp.Type = "limit"

	if orderResp.Status != "success" {
		return &orderResp, errors.Errorf("order not successful: %s", orderResp.Status)
	}

	return &orderResp, nil
}

// CancelOrderResp is different from the normal order resp because in this case, the
// id comes back as a quoted string which must be decoded differently. This is to be used
// to decoding and then converted to the other type afterwards.
type CancelOrderResp struct {
	OrderID      int64  `json:"orderId,string"`
	Status       string `json:"status"`
	CurrencyPair string `json:"currency_pair"`
}

// CancelOpenOrders cancels all open orders that have the order id in the orders slice.
func (k *API) CancelOpenOrders(orders []int64, currency string) ([]CancelOrderResp, error) {

	orderStrings := []string{}
	for _, v := range orders {
		orderStrings = append(orderStrings, fmt.Sprintf("%d", v))
	}

	data := url.Values{
		"currency_pair": {currency},
		"id":            orderStrings,
		"nonce":         {k.GetNonce()},
	}

	req, err := k.NewRequest(CancelOpenOrders, "POST", data)
	if err != nil {
		return nil, errors.Wrap(err, "make korbit cancel order request")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "placing korbit cancel order")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	var orderResps []CancelOrderResp
	err = json.NewDecoder(resp.Body).Decode(&orderResps)
	if err != nil {
		return nil, errors.Wrap(err, "json decode korbit cancel")
	}

	return orderResps, nil
}

// ListOrderResp is the response that is given to list orders.
type ListOrderResp struct {
	Timestamp int64    `json:"timestamp"`
	ID        int64    `json:"id,string"`
	Type      string   `json:"type"`
	Price     Currency `json:"price"`
	Total     Currency `json:"total"`
	Open      Currency `json:"open"`
}

// ListOpenOrders lists the open orders that belong to the account.
func (k *API) ListOpenOrders(coin string) (*[]ListOrderResp, error) {

	// adding querystring parameters onto the base url
	url := fmt.Sprintf("%s?currency_pair=%s", ListOpenOrders, coin)
	req, err := k.NewRequest(url, "GET", nil)
	if err != nil {
		return nil, errors.Wrap(err, "make korbit list order request")
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "placing korbit list order")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	var orderResps []ListOrderResp
	err = json.NewDecoder(resp.Body).Decode(&orderResps)
	if err != nil {
		return nil, errors.Wrap(err, "unmarhsal korbit list order")
	}

	return &orderResps, nil
}

// TransactionsResponse is the response that comes from querying transactions. It differs
// from the BTC response because in this case the ID field comes unquoted from the API.
type TransactionsResponse struct {
	Timestamp   int64      `json:"timestamp"`
	CompletedAt int64      `json:"completedAt"`
	ID          int64      `json:"id"`
	Type        string     `json:"type"`
	Fee         Currency   `json:"fee"`
	Balances    []Currency `json:"balances"`
	FillsDetail FillDetail `json:"fillsDetail"`
}

// FillDetail is the details of a specific order fill.
type FillDetail struct {
	Price        Currency `json:"price"`
	Amount       Currency `json:"amount"`
	NativeAmount Currency `json:"native_amount"`
	OrderID      int64    `json:"orderID,string"`
}

// GetTransactionHistory is for getting the trade history of a user.
// explanation for the url parameters can be found at :
// https://apidocs.korbit.co.kr/#user-:-transaction-history---order-fills,-krw/btc-deposit-and-transfer
func (k *API) GetTransactionHistory(coin, category, offset, limit, orderID string) (*[]TransactionsResponse, error) {
	url := fmt.Sprintf("%s?", TransactionHistory)

	if coin == "" {
		return nil, errors.New("coin must be specified")
	}
	if category == "" {
		return nil, errors.New("category must be one of 'fills' 'fiats' or 'coins'")
	}
	if offset != "" {
		url = fmt.Sprintf("%soffset=%s&", url, offset)
	}
	if limit != "" {
		url = fmt.Sprintf("%slimit=%s&", url, limit)
	}
	if orderID != "" {
		url = fmt.Sprintf("%sorder_id=%s&", url, orderID)
	}

	url = fmt.Sprintf("%scurrency_pair=%s", url, coin)
	req, err := k.NewRequest(url, "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "korbit transaction history for: %s", coin)
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "korbit transaction history for: %s", coin)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d Header: %v", resp.StatusCode, resp.Header)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "transaction history read response")
	}

	var buffer *bytes.Buffer
	// if the coin is bitcoin then the id field is quoted, regexing to find and replace
	// with no quotes so it will be the same across coins.
	switch coin {
	case BTCKRW:
		re := regexp.MustCompile(`"id":"([0-9]+)"`)
		new := []byte(`"id":$1`)
		replaced := re.ReplaceAll(respBytes, new)
		buffer = bytes.NewBuffer(replaced)
	default:
		buffer = bytes.NewBuffer(respBytes)
	}

	var retResp []TransactionsResponse
	err = json.NewDecoder(buffer).Decode(&retResp)
	if err != nil {
		return nil, errors.Wrapf(err, "get transaction history for: %s", coin)
	}

	return &retResp, nil
}

// TotalBuySellHistory gives the buy and sell history for a transaction response.
func (k *API) TotalBuySellHistory(t []TransactionsResponse, orderSize float64, from, to *time.Time) (
	buys, sells float64, trades int) {

	for i, v := range t { // filter out the transactions that are not wanted.
		switch {
		case v.FillsDetail.Amount.Value != orderSize && orderSize != 0:
			t = append(t[:i], t[i+1:]...)
		case v.Timestamp < from.Unix() || v.Timestamp > to.Unix():
			t = append(t[:i], t[i+1:]...)
		}
	}

	for _, v := range t {
		switch v.Type {
		case "buy":
			buys += v.FillsDetail.NativeAmount.Value
			trades++
		case "sell":
			sells += v.FillsDetail.NativeAmount.Value
			trades++
		}
	}

	return buys, sells, trades
}
