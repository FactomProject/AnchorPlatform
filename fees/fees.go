package fees

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// ETHFees represents ETH fees received from external API
// Divide by 10 to get fees in gwei!
// https://docs.ethgasstation.info
type ETHFees struct {
	Fast    int // < 2m
	Fastest int // < 30s
	SafeLow int
	Average int // < 5m
}

// BTCFees represents BTC fees received from external API
// https://mempool.space/api
type BTCFees struct {
	FastestFee  int // next block
	HalfHourFee int
	HourFee     int
	MinimumFee  int
}

// GetETHFees returns real-time ETHFees
func GetETHFees() (*ETHFees, error) {

	api := "https://ethgasstation.info/api/ethgasAPI.json"
	fees := &ETHFees{}

	jsonResp, err := getJSON(api)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonResp, fees)
	if err != nil {
		return nil, err
	}

	return fees, nil

}

// GetBTCFees returns real-time BTCFees
func GetBTCFees() (*BTCFees, error) {

	api := "https://mempool.space/api/v1/fees/recommended"
	fees := &BTCFees{}

	jsonResp, err := getJSON(api)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonResp, fees)
	if err != nil {
		return nil, err
	}

	return fees, nil

}

func getJSON(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
