package korbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Limit and other variables here are for making sure all the strings are the same across
// use cases and easy to change.
var (
	Limit   = "limit"
	Market  = "market"
	Buy     = "buy"
	Sell    = "sell"
	Ask     = "ask"
	Bid     = "bid"
	Success = "success"
	Nonce   = int64(0)
)

// BTCKRW and other symbols here are the ticker names that korbit uses.
const (
	BTCKRW = "btc_krw"
	ETHKRW = "eth_krw"
	ETCKRW = "etc_krw"
	XRPKRW = "xrp_krw"
	KRW    = "krw"
	BTC    = "btc"
	ETC    = "etc"
	ETH    = "eth"
	XRP    = "xrp"
)

// currencies are the currencies that are currently active on the exchange.
var currencies = []string{BTCKRW, ETHKRW, ETCKRW}

// LoginURL and other things here are the urls that are used for korbit.
var (
	LoginURL           = "https://api.korbit.co.kr/v1/oauth2/access_token"
	WalletStatus       = "https://api.korbit.co.kr/v1/user/balances"
	BtcWithdrawal      = "https://api.korbit.co.kr/v1/user/coins/out"
	PlaceBid           = "https://api.korbit.co.kr/v1/user/orders/buy"
	PlaceAsk           = "https://api.korbit.co.kr/v1/user/orders/sell"
	CancelOpenOrders   = "https://api.korbit.co.kr/v1/user/orders/cancel"
	ListOpenOrders     = "https://api.korbit.co.kr/v1/user/orders/open"
	TransactionHistory = "https://api.korbit.co.kr/v1/user/transactions"
	TradeVolumeAndFees = "https://api.korbit.co.kr/v1/user/volume"
	GetOrderbook       = "https://api.korbit.co.kr/v1/orderbook"
)

// API is the object that holds the client and has all of the API methods.
type API struct {
	Token        *Token
	Client       *http.Client
	Nonce        int64
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

// Token has the token and refresh token which takes care of the authentication
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Timestamp    time.Time `json:"timestamp"`
	RefreshToken string    `json:"refresh_token"`
}

// NewKorbitAPI returns a korbit API object with all the necessary fields.
func NewKorbitAPI(clientID, clientSecret, username, password string) *API {

	api := API{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		Client:       &http.Client{},
		Nonce:        time.Now().Unix(),
	}

	return &api
}

// Login takes care of logging in the user, it is not in the newAPI because it is possible
// that the api is only being used for public endpoints.
func (k *API) Login() error {
	body := url.Values{
		"client_id":     {k.ClientID},
		"client_secret": {k.ClientSecret},
		"username":      {k.Username},
		"password":      {k.Password},
		"grant_type":    {"password"},
	}

	resp, err := http.PostForm(LoginURL, body)
	if err != nil {
		return errors.Wrap(err, "korbit post login")
	}
	defer resp.Body.Close()

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return errors.Wrap(err, "json decode")
	}

	token.Timestamp = time.Now()
	k.Token = &token

	return nil
}

// GetNonce returns an ever increasing nonce for the requests to the API.
func (k *API) GetNonce() string {
	k.Nonce++
	return fmt.Sprintf("%d", k.Nonce)
}

// ShouldRefresh returns true if there is less than 10 minutes left on the token
// before it will expire.
func (k *API) ShouldRefresh() bool {
	minusTen := time.Duration(k.Token.ExpiresIn - 600)
	tenMinsLeft := k.Token.Timestamp.Add(minusTen * time.Second)
	return time.Now().After(tenMinsLeft)
}

// NewRequest makes a new request of the given type with the proper authorization headers
// for Korbit.
func (k *API) NewRequest(url, method string, body url.Values) (*http.Request, error) {

	b := bytes.NewBufferString(body.Encode())
	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, errors.Wrap(err, "make korbit req")
	}

	token := fmt.Sprintf("%s %s", k.Token.TokenType, k.Token.AccessToken)
	req.Header.Set("Authorization", token)

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}

// RefreshToken takes care of refreshing the access token which is needed every hour (and
// is tracked by the server.)
func (k *API) RefreshToken() error {
	body := url.Values{
		"client_id":     {k.ClientID},
		"client_secret": {k.ClientSecret},
		"refresh_token": {k.Token.RefreshToken},
		"grant_type":    {"refresh_token"},
	}

	resp, err := http.PostForm(LoginURL, body)
	if err != nil {
		return errors.Wrap(err, "korbit post login")
	}
	defer resp.Body.Close()

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return errors.Wrap(err, "json decode refresh token")
	}

	token.Timestamp = time.Now()
	k.Token = &token

	return nil
}
