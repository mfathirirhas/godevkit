// +build windows

package server

import (
	"fmt"
	nethttp "net/http"

	_ "github.com/valyala/fasthttp/reuseport"
)

// graceful is not support in Windows. Using built-in package instead. This is for avoiding this package failed to run locally, rarely Windows used in server now.
func (http *HTTP) serve() error {
	srv := &nethttp.Server{
		Addr:        fmt.Sprintf(":%d", http.port),
		Handler:     http.cors.Handler(http.handlers),
		IdleTimeout: http.idleTimeout,
	}

	// TODO add support for tls
	return srv.ListenAndServe()
}
