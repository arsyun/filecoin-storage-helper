package main

import (
	"context"
	"fmt"
	"go-filecoin-storage-helper/api"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/filecoin-project/lotus/lib/jsonrpc"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"golang.org/x/xerrors"
)

func serveRPC(pn api.Proxy, addr multiaddr.Multiaddr) error {
	rpcServer := jsonrpc.NewServer()
	rpcServer.Register("proxy", pn)

	http.Handle("/rpc/v0", rpcServer)

	lst, err := manet.Listen(addr)
	if err != nil {
		return xerrors.Errorf("could not listen: %w", err)
	}

	srv := &http.Server{Handler: http.DefaultServeMux}

	sigChan := make(chan os.Signal, 2)
	go func() {
		<-sigChan
		if err := srv.Shutdown(context.TODO()); err != nil {
			fmt.Printf("shutting down RPC server failed: %s", err)
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	return srv.Serve(manet.NetListener(lst))
}
