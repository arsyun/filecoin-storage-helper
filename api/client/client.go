package client

import (
	"go-filecoin-storage-helper/api"
	"go-filecoin-storage-helper/api/apistruct"
	"net/http"

	"github.com/filecoin-project/lotus/lib/jsonrpc"
)

func NewProxyRPC(addr string, requestHeader http.Header) (api.Proxy, jsonrpc.ClientCloser, error) {
	var res apistruct.ProxyStruct
	closer, err := jsonrpc.NewMergeClient(addr, "proxy",
		[]interface{}{
			&res.Internal,
		}, requestHeader)

	return &res, closer, err
}
