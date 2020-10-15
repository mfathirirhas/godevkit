// +build darwin linux freebsd openbsd netbsd

package server

import (
	"fmt"
	"net/http"

	_grace "github.com/facebookgo/grace/gracehttp"
)

func (s *Server) serve() error {
	return _grace.Serve(&http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     s.cors.Handler(s.handlers),
		IdleTimeout: s.idleTimeout,
	})
}
