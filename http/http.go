package http

import (
	"fmt"
	"net/http"
	nethttp "net/http"
	"strings"

	_grace "github.com/facebookgo/grace/gracehttp"
	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

type HTTP struct {
	handlers *_router.Router
	errChan  chan error
	port     string
	cors     *_cors.Cors
}

type Opts struct {
	Port string

	// Cors optional, can be nil, if nil then default will be set.
	Cors *Cors
}

// Cors corst options
type Cors struct {
	AllowedOrigins         []string
	AllowOriginFunc        func(origin string) bool
	AllowOriginRequestFunc func(r *http.Request, origin string) bool
	AllowedMethods         []string
	AllowedHeaders         []string
	ExposedHeaders         []string
	MaxAge                 int
	AllowCredentials       bool
	OptionsPassthrough     bool
}

func New(opts *Opts) *HTTP {
	h := _router.New()
	cors := _cors.Default()
	if opts.Cors != nil {
		cors = _cors.New(_cors.Options{
			AllowedOrigins:         opts.Cors.AllowedOrigins,
			AllowOriginFunc:        opts.Cors.AllowOriginFunc,
			AllowOriginRequestFunc: opts.Cors.AllowOriginRequestFunc,
			AllowedMethods:         opts.Cors.AllowedMethods,
			AllowedHeaders:         opts.Cors.AllowedHeaders,
			ExposedHeaders:         opts.Cors.ExposedHeaders,
			MaxAge:                 opts.Cors.MaxAge,
			AllowCredentials:       opts.Cors.AllowCredentials,
			OptionsPassthrough:     opts.Cors.OptionsPassthrough,
		})
	}
	return &HTTP{
		handlers: h,
		port:     strings.Replace(opts.Port, ":", "", -1),
		cors:     cors,
	}
}

func (http *HTTP) Run() {
	// TODO add SO_REUSEPORT support
	http.errChan <- _grace.Serve(&nethttp.Server{
		Addr:    fmt.Sprintf(":%s", http.port),
		Handler: http.cors.Handler(http.handlers),
	})
}

func (http *HTTP) ListenError() <-chan error {
	return http.errChan
}

func (http *HTTP) GET(path string, handler _router.Handle) {
	http.handlers.GET(path, handler)
}

func (http *HTTP) HEAD(path string, handler _router.Handle) {
	http.handlers.HEAD(path, handler)
}

func (http *HTTP) POST(path string, handler _router.Handle) {
	http.handlers.POST(path, handler)
}

func (http *HTTP) PUT(path string, handler _router.Handle) {
	http.handlers.POST(path, handler)
}

func (http *HTTP) DELETE(path string, handler _router.Handle) {
	http.handlers.DELETE(path, handler)
}

func (http *HTTP) PATCH(path string, handler _router.Handle) {
	http.handlers.PATCH(path, handler)
}

func (http *HTTP) OPTIONS(path string, handler _router.Handle) {
	http.handlers.OPTIONS(path, handler)
}
