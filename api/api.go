package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/AdamSLevy/jsonrpc2"
	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/AnchorPlatform/fees"
)

type API struct {
	conf *config.Config
}

type FeesResponse struct {
	BTC int
	ETH int
}

func NewAPI(conf *config.Config) *API {

	api := API{conf: conf}

	jsonrpc2.RegisterMethod("fees", getFees)

	http.HandleFunc("/", UIHandler)
	http.HandleFunc("/v2", jsonrpc2.HTTPRequestHandler)

	return &api

}

func (api *API) Start() error {
	fmt.Printf("Starting JSON-RPC API at http://localhost:%d\n", api.conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(api.conf.HTTPPort), nil))

	return nil
}

func getFees(params interface{}) jsonrpc2.Response {

	btcf, _ := fees.GetBTCFees()
	ethf, _ := fees.GetETHFees()

	resp := &FeesResponse{}
	resp.BTC = btcf.FastestFee
	resp.ETH = ethf.Fast / 10

	return jsonrpc2.NewResponse(resp)
}

func UIHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UI there\n")
}
