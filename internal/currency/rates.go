package currency

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/spf13/viper"
)

// Currency is a list of the supported currencies by the system
type Currency string

func (c Currency) String() string {
	return string(c)
}

// RateResponse will return JSON decoded response from the exchange upstream
type RateResponse struct {
	Waves float64 `json:"WAVES,omitempty"`
	Btc   float64 `json:"BTC,omitempty"`
	Eth   float64 `json:"ETH,omitempty"`
}

var (
	// Globally accepted currencies by the governments
	EUR Currency = "EUR"

	// Criptocurrencies
	WAVES Currency = "WAVES"
	BTC   Currency = "BTC"
	ETH   Currency = "ETH"
)

func getCurrenciesAsStringMap(currencies []Currency) []string {
	var toReturn []string
	for _, currency := range currencies {
		toReturn = append(toReturn, currency.String())
	}
	return toReturn
}

func getCacheKey(source Currency, destinations []Currency) []byte {
	toJoin := []string{
		"exchange_rate_key",
		source.String(),
	}
	toJoin = append(toJoin, getCurrenciesAsStringMap(destinations)...)
	return []byte(strings.Join(toJoin, "_"))
}

func buildExchangeRateRequestURL(source Currency, destinations []Currency) (*url.URL, error) {
	if len(destinations) == 0 {
		return nil, errors.New("Pick at least one currency destination to build url")
	}

	v := url.Values{}
	v.Add("fsym", source.String())
	v.Add("tsyms", strings.Join(getCurrenciesAsStringMap(destinations), ","))

	toReturn := url.URL{
		Scheme:   viper.GetString("currency_exchange_rate_scheme"),
		Host:     viper.GetString("currency_exchange_host"),
		Path:     viper.GetString("currency_exchange_rate_path"),
		RawQuery: v.Encode(),
	}
	return &toReturn, nil
}

// GetExchangeRatesFromSourceAndDestinations will take the source (given) rate and return back
// all of the available exchange rates.
// TODO: Currently hardcoded to EUR, BTC, WAVES and ETH
func GetExchangeRatesFromSourceAndDestinations(source Currency, destinations []Currency) (*RateResponse, error) {
	client := &http.Client{
		Timeout: viper.GetDuration("currency_exchange_rate_client_timeout_s") * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: viper.GetDuration("currency_exchange_rate_transport_timeout_s") * time.Second,
			}).Dial,
			TLSHandshakeTimeout: viper.GetDuration("currency_exchange_rate_transport_handshake_timeout_s") * time.Second,
		},
	}

	rURL, rURLErr := buildExchangeRateRequestURL(source, destinations)
	if rURLErr != nil {
		return nil, rURLErr
	}

	req, reqErr := http.NewRequest(http.MethodGet, rURL.String(), nil)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Content-Type", "application/json")

	res, resErr := client.Do(req)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	var toReturn *RateResponse
	if err := json.Unmarshal(body, &toReturn); err != nil {
		return nil, err
	}

	return toReturn, nil
}

// GetExchangeRatesFromSourceAndDestinationsWithCache similar to GetExchangeRatesFromSourceAndDestinations but with cache
// configurable by the configuration values
func GetExchangeRatesFromSourceAndDestinationsWithCache(cache *freecache.Cache, source Currency, destinations []Currency) (*RateResponse, error) {
	if !viper.GetBool("currency_exchange_rate_cache_enabled") {
		res, err := GetExchangeRatesFromSourceAndDestinations(source, destinations)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	var bEntry bytes.Buffer
	cacheKey := getCacheKey(source, destinations)
	enc := gob.NewEncoder(&bEntry)
	dec := gob.NewDecoder(&bEntry)

	entry, err := cache.Get(cacheKey)
	if err != nil {
		res, err := GetExchangeRatesFromSourceAndDestinations(source, destinations)
		if err != nil {
			return nil, err
		}
		if err := enc.Encode(res); err != nil {
			return nil, err
		}
		cache.Set(
			cacheKey,
			bEntry.Bytes(),
			viper.GetInt("currency_exchange_rate_cache_expiry_s"),
		)
	}

	if _, err := bEntry.Write(entry); err != nil {
		return nil, err
	}

	var res *RateResponse
	if err := dec.Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}
