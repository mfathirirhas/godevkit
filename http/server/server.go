package server

import (
	"log"
	"net/http"
	"time"

	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

// TODO add tls support

type Server struct {
	handlers    *_router.Router
	errChan     chan error
	port        uint16
	idleTimeout time.Duration
	cors        *_cors.Cors
}

type Opts struct {
	Port uint16

	// IdleTimeout keep-alive timeout while waiting for the next request coming. If empty then no timeout.
	IdleTimeout time.Duration

	// ErrorLog Optional. Logging error happening in connection, handlers, or filesystem.
	ErrorLog *log.Logger

	// Cors optional, can be nil, if nil then default will be set.
	Cors *Cors
}

// Cors corst options
type Cors struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	MaxAge             int
	AllowCredentials   bool
	OptionsPassthrough bool
	IsDebug            bool
}

func New(opts *Opts) *Server {
	h := _router.New()
	cors := _cors.Default()
	if opts.Cors != nil {
		cors = _cors.New(_cors.Options{
			AllowedOrigins:     opts.Cors.AllowedOrigins,
			AllowedMethods:     opts.Cors.AllowedMethods,
			AllowedHeaders:     opts.Cors.AllowedHeaders,
			ExposedHeaders:     opts.Cors.ExposedHeaders,
			MaxAge:             opts.Cors.MaxAge,
			AllowCredentials:   opts.Cors.AllowCredentials,
			OptionsPassthrough: opts.Cors.OptionsPassthrough,
			Debug:              opts.Cors.IsDebug,
		})
	}
	return &Server{
		handlers:    h,
		port:        opts.Port,
		idleTimeout: opts.IdleTimeout,
		cors:        cors,
	}
}

// Run the server. Blocking. Execute it inside goroutine.
func (s *Server) Run() {
	// TODO add SO_REUSEPORT support
	s.errChan <- s.serve()
}

func (s *Server) ListenError() <-chan error {
	return s.errChan
}

func f(h http.HandlerFunc) _router.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps _router.Params) {
		if len(ps) > 0 {
			urlValues := r.URL.Query()
			for i := range ps {
				urlValues.Add(ps[i].Key, ps[i].Value)
			}
			r.URL.RawQuery = urlValues.Encode()
		}
		h(w, r)
	}
}

func (s *Server) GET(path string, handler http.HandlerFunc) {
	s.handlers.GET(path, f(handler))
}

func (s *Server) HEAD(path string, handler http.HandlerFunc) {
	s.handlers.HEAD(path, f(handler))
}

func (s *Server) POST(path string, handler http.HandlerFunc) {
	s.handlers.POST(path, f(handler))
}

func (s *Server) PUT(path string, handler http.HandlerFunc) {
	s.handlers.POST(path, f(handler))
}

func (s *Server) DELETE(path string, handler http.HandlerFunc) {
	s.handlers.DELETE(path, f(handler))
}

func (s *Server) PATCH(path string, handler http.HandlerFunc) {
	s.handlers.PATCH(path, f(handler))
}

func (s *Server) OPTIONS(path string, handler http.HandlerFunc) {
	s.handlers.OPTIONS(path, f(handler))
}
