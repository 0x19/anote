package currency

import (
	"testing"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"
)

type exchangeArgs struct {
	Source       Currency
	Destinations []Currency
}

func TestExchangeRateTestUrl(t *testing.T) {
	lAssert := assert.New(t)
	testCases := []struct {
		Name     string
		Args     exchangeArgs
		Expected string
		Error    bool
	}{
		{
			Name: "EUR to WAVES, BTC and ETH",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{WAVES, BTC, ETH},
			},
			Expected: "https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=WAVES%2CBTC%2CETH",
			Error:    false,
		},
		{
			Name: "EUR to WAVES",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{WAVES},
			},
			Expected: "https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=WAVES",
			Error:    false,
		},
		{
			Name: "EUR to NONE should raise error",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{},
			},
			Expected: "https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=WAVES",
			Error:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, err := buildExchangeRateRequestURL(tc.Args.Source, tc.Args.Destinations)
			if !tc.Error {
				lAssert.NoError(err)
				lAssert.Equal(req.String(), tc.Expected)
			} else {
				lAssert.Error(err)
			}
		})
	}
}

func TestGetExchangeRatesFromSourceAndDestinations(t *testing.T) {
	lAssert := assert.New(t)
	testCases := []struct {
		Name  string
		Args  exchangeArgs
		Error bool
	}{
		{
			Name: "EUR to WAVES, BTC and ETH",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{WAVES, BTC, ETH},
			},
			Error: false,
		},
		{
			Name: "EUR to no destinations should result in error",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{},
			},
			Error: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := GetExchangeRatesFromSourceAndDestinations(tc.Args.Source, tc.Args.Destinations)
			if !tc.Error {
				lAssert.NoError(err)
				lAssert.NotZero(resp.Waves)
				lAssert.NotZero(resp.Btc)
				lAssert.NotZero(resp.Eth)
			} else {
				lAssert.Error(err)
			}
		})
	}
}

func TestGetExchangeRatesFromSourceAndDestinationsWithCache(t *testing.T) {
	lAssert := assert.New(t)

	// 1MB cache
	cache := freecache.NewCache(1024 * 1024)

	testCases := []struct {
		Name  string
		Args  exchangeArgs
		Error bool
	}{
		{
			Name: "EUR to WAVES, BTC and ETH",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{WAVES, BTC, ETH},
			},
			Error: false,
		},
		{
			Name: "EUR to no destinations should result in error",
			Args: exchangeArgs{
				Source:       EUR,
				Destinations: []Currency{},
			},
			Error: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := GetExchangeRatesFromSourceAndDestinationsWithCache(cache, tc.Args.Source, tc.Args.Destinations)
			if !tc.Error {
				lAssert.NoError(err)
				lAssert.NotZero(resp.Waves)
				lAssert.NotZero(resp.Btc)
				lAssert.NotZero(resp.Eth)
			} else {
				lAssert.Error(err)
			}
		})
	}
}
