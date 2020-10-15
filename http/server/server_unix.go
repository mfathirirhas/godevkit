// +build darwin linux freebsd openbsd netbsd

package server

import (
	"fmt"
	nethttp "net/http"

	_grace "github.com/facebookgo/grace/gracehttp"
)

func (http *Server) serve() error {
	return _grace.Serve(&nethttp.Server{
		Addr:        fmt.Sprintf(":%d", http.port),
		Handler:     http.cors.Handler(http.handlers),
		IdleTimeout: http.idleTimeout,
	})
}
