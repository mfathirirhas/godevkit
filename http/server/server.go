package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	_uuid "github.com/google/uuid"
	_router "github.com/julienschmidt/httprouter"
	_cors "github.com/rs/cors"
)

// TODO add tls support

type Server struct {
	handlers     *_router.Router
	errChan      chan error
	port         uint16
	idleTimeout  time.Duration
	enableLogger bool
	logger       *log.Logger
	cors         *_cors.Cors
}

type Opts struct {
	Port uint16

	// EnableLogger enable logging for incoming requests
	EnableLogger bool

	// IdleTimeout keep-alive timeout while waiting for the next request coming. If empty then no timeout.
	IdleTimeout time.Duration

	// Cors optional, can be nil, if nil then default will be set.
	Cors *Cors
}

// Cors corst options
type Cors struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
	IsDebug          bool
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
			OptionsPassthrough: true,
			Debug:              opts.Cors.IsDebug,
		})
	}
	logger := log.New(os.Stderr, "", 0)
	return &Server{
		handlers:     h,
		port:         opts.Port,
		idleTimeout:  opts.IdleTimeout,
		enableLogger: opts.EnableLogger,
		logger:       logger,
		cors:         cors,
		errChan:      make(chan error),
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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// default if not set is 200
	return &responseWriter{w, http.StatusOK}
}

func f(next http.HandlerFunc) _router.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps _router.Params) {
		if r.Header.Get("Request-Id") == "" && r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("Request-Id", _uuid.New().String())
		}
		if r.Header.Get("Request-Id") == "" && r.Header.Get("X-Request-Id") != "" {
			r.Header.Set("Request-Id", r.Header.Get("X-Request-Id"))
		}
		if len(ps) > 0 {
			urlValues := r.URL.Query()
			for i := range ps {
				urlValues.Add(ps[i].Key, ps[i].Value)
			}
			r.URL.RawQuery = urlValues.Encode()
		}
		rw := newResponseWriter(w)
		next(rw, r)
	}
}

func (s *Server) log(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		elapsed := time.Since(start)
		var statusCode int
		rw, ok := w.(*responseWriter)
		if !ok { // impossible...!!! but let be safe.
			statusCode = http.StatusOK // default http.ResponseWriter status code
		} else {
			statusCode = rw.statusCode
		}
		if s.enableLogger {
			if statusCode >= 400 {
				s.logger.Printf("%s | httpserver | %s | %d | %s | %v | %s\n", time.Now().Format(time.RFC3339), r.Method, statusCode, r.URL.Path, elapsed, r.Header.Get("Request-Id"))
			} else {
				fmt.Printf("%s | httpserver | %s | %d | %s | %v | %s\n", time.Now().Format(time.RFC3339), r.Method, statusCode, r.URL.Path, elapsed, r.Header.Get("Request-Id"))
			}
		}
	}
}

func (s *Server) recoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ResponseString(w, r, http.StatusInternalServerError, "httpserver got panic")
				s.logger.Printf("%s | httpserver | %s | %s | %s | %s\n", time.Now().Format(time.RFC3339), "PANIC", r.Method, r.URL.Path, r.Header.Get("Request-Id"))
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC START (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				debug.PrintStack()
				s.logger.Printf("☠️ ☠️ ☠️ ☠️ ☠️ ☠️  PANIC END (%s) ☠️ ☠️ ☠️ ☠️ ☠️ ☠️", r.Header.Get("Request-Id"))
				return
			}
		}()
		next(w, r)
	}
}

func responseHeader(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Request-Id", r.Header.Get("Request-Id"))
	w.Header().Set(fmt.Sprintf("X-Req-Id_%s-Status_code", r.Header.Get("Request-Id")), strconv.Itoa(statusCode))
	w.WriteHeader(statusCode)
}

// Response response by writing the body to http.ResponseWriter.
// Call at the end line of your handler.
func Response(w http.ResponseWriter, r *http.Request, statusCode int, body []byte) {
	responseHeader(w, r, statusCode)
	w.Write(body)
}

// ResponseJSON response by writing body with json encoder into http.ResponseWriter.
// Body must be either struct or map[string]interface{}. Otherwise would result in incorrect parsing at client side.
// If you have []byte as response body, then use Response function instead.
// Call at the end line of your handler.
func ResponseJSON(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) error {
	responseHeader(w, r, statusCode)
	return json.NewEncoder(w).Encode(body)
}

// ResponseString response in form of string whatever passed into body param.
// Call at the end line of your handler.
func ResponseString(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) {
	responseHeader(w, r, statusCode)
	fmt.Fprintf(w, "%v", body)
}

func (s *Server) GET(path string, handler http.HandlerFunc) {
	s.handlers.GET(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) HEAD(path string, handler http.HandlerFunc) {
	s.handlers.HEAD(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) POST(path string, handler http.HandlerFunc) {
	s.handlers.POST(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) PUT(path string, handler http.HandlerFunc) {
	s.handlers.POST(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) DELETE(path string, handler http.HandlerFunc) {
	s.handlers.DELETE(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) PATCH(path string, handler http.HandlerFunc) {
	s.handlers.PATCH(path, f(s.recoverPanic(s.log(handler))))
}

func (s *Server) OPTIONS(path string, handler http.HandlerFunc) {
	s.handlers.OPTIONS(path, f(s.recoverPanic(s.log(handler))))
}
