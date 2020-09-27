package server

import (
	"fmt"
	"net/http"

	_http "github.com/mfathirirhas/godevkit/http/server"
)

type Router struct {
	handlers *_http.HTTP
}

func Init() *Router {
	cors := &_http.Cors{
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"foo.com"},
	}
	h := _http.New(&_http.Opts{
		Port: 8282,
		Cors: cors,
	})

	s := &Service{}
	h.GET("/get/:param1/:param2", s.GetData())
	h.POST("/set/:urlparam", s.SetData())

	return &Router{
		handlers: h,
	}
}

// Run blocking
func (r *Router) Run() {
	r.handlers.Run()
}

func (r *Router) Err() <-chan error {
	return r.handlers.ListenError()
}

type Service struct {
}

func (s *Service) GetData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p1 := r.URL.Query().Get("param1")
		p2 := r.URL.Query().Get("param2")
		fmt.Fprint(w, fmt.Sprintf("Param1:=%s , Param2:=%s", p1, p2))
	}
}

func (s *Service) SetData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Query().Get("urlparam")
		r.ParseForm()
		p1 := r.FormValue("Param1")
		p2 := r.FormValue("Param2")
		fmt.Fprint(w, fmt.Sprintf("Param1:=%s, Param2:=%s, URL Param:%s", p1, p2, urlParam))
	}
}
