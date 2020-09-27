package http

import (
	"fmt"
	nethttp "net/http"

	_grace "github.com/facebookgo/grace/gracehttp"
	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

// TODO add tls support

type HTTP struct {
	handlers *_router.Router
	errChan  chan error
	port     uint16
	cors     *_cors.Cors
}

type Opts struct {
	Port uint16

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

func New(opts *Opts) *HTTP {
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
	return &HTTP{
		handlers: h,
		port:     opts.Port,
		cors:     cors,
	}
}

// Run the server. Blocking. Execute it inside goroutine.
func (http *HTTP) Run() {
	// TODO add SO_REUSEPORT support
	http.errChan <- _grace.Serve(&nethttp.Server{
		Addr:    fmt.Sprintf(":%d", http.port),
		Handler: http.cors.Handler(http.handlers),
	})
}

func (http *HTTP) ListenError() <-chan error {
	return http.errChan
}

func f(h nethttp.HandlerFunc) _router.Handle {
	return func(w nethttp.ResponseWriter, r *nethttp.Request, ps _router.Params) {
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

func (http *HTTP) GET(path string, handler nethttp.HandlerFunc) {
	http.handlers.GET(path, f(handler))
}

func (http *HTTP) HEAD(path string, handler nethttp.HandlerFunc) {
	http.handlers.HEAD(path, f(handler))
}

func (http *HTTP) POST(path string, handler nethttp.HandlerFunc) {
	http.handlers.POST(path, f(handler))
}

func (http *HTTP) PUT(path string, handler nethttp.HandlerFunc) {
	http.handlers.POST(path, f(handler))
}

func (http *HTTP) DELETE(path string, handler nethttp.HandlerFunc) {
	http.handlers.DELETE(path, f(handler))
}

func (http *HTTP) PATCH(path string, handler nethttp.HandlerFunc) {
	http.handlers.PATCH(path, f(handler))
}

func (http *HTTP) OPTIONS(path string, handler nethttp.HandlerFunc) {
	http.handlers.OPTIONS(path, f(handler))
}
