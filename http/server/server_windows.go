// +build windows

package server

import (
	"fmt"
	"net/http"

	_reuseport "github.com/valyala/fasthttp/reuseport"
)

// graceful is not support in Windows. Using built-in package instead. This is for avoiding this package failed to run locally, rarely Windows used in server now.
func (s *Server) serve() error {
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     s.cors.Handler(s.handlers),
		IdleTimeout: s.idleTimeout,
	}

	// TODO add support for tls
	l, err := _reuseport.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(l)
}
