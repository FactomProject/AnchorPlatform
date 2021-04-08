package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/AdamSLevy/jsonrpc2/v14"
	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/AnchorPlatform/fees"
)

type API struct {
	conf *config.Config
}

type FeesResponse struct {
	BTC int `json:"btc"`
	ETH int `json:"eth"`
}

func NewAPI(conf *config.Config) *API {

	api := API{conf: conf}

	http.HandleFunc("/", UIHandler)

	methods := jsonrpc2.MethodMap{
		"fees": getFees,
	}

	apihandler := jsonrpc2.HTTPRequestHandler(methods, log.New(os.Stdout, "", 0))
	http.HandleFunc("/v2", apihandler)

	return &api

}

func (api *API) Start() error {
	fmt.Printf("Starting JSON-RPC API at http://localhost:%d\n", api.conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(api.conf.HTTPPort), nil))

	return nil
}

func getFees(_ context.Context, _ json.RawMessage) interface{} {

	btcf, _ := fees.GetBTCFees()
	ethf, _ := fees.GetETHFees()

	resp := &FeesResponse{}
	resp.BTC = btcf.FastestFee
	resp.ETH = ethf.Fast / 10

	return resp
}

func UIHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UI there\n")
}
