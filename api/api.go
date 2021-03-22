package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc/jsonrpc"
	"strconv"

	"github.com/FactomProject/AnchorPlatform/config"
)

type API struct {
	conf *config.Config
}

func NewAPI(conf *config.Config) *API {

	api := API{conf: conf}

	http.HandleFunc("/v2", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		res := NewRPCRequest(req.Body).Call()
		io.Copy(w, res)
	})

	return &api

}

func (api *API) Start() error {
	fmt.Printf("Starting JSON-RPC API at http://localhost:%d\n", api.conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(api.conf.HTTPPort), nil))

	return nil
}

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r    io.Reader     // holds the JSON formated RPC request
	rw   io.ReadWriter // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// NewRPCRequest returns a new rpcRequest.
func NewRPCRequest(r io.Reader) *rpcRequest {
	var buf bytes.Buffer
	done := make(chan bool)
	return &rpcRequest{r, &buf, done}
}

// Read implements the io.ReadWriteCloser Read method.
func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Write implements the io.ReadWriteCloser Write method.
func (r *rpcRequest) Write(p []byte) (n int, err error) {
	return r.rw.Write(p)
}

// Close implements the io.ReadWriteCloser Close method.
func (r *rpcRequest) Close() error {
	r.done <- true
	return nil
}

// Call invokes the RPC request, waits for it to complete, and returns the results.
func (r *rpcRequest) Call() io.Reader {
	go jsonrpc.ServeConn(r)
	<-r.done
	return r.rw
}
