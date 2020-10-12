package server

import (
	"encoding/json"
	"fmt"
	"log"
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
	h.GET("/get", s.GetData())
	h.POST("/post", s.SetData())
	h.POST("/post/json", s.SetJSON())
	h.POST("/post/multi", s.SetMulti())

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
		p1 := r.FormValue("param1")
		p2 := r.FormValue("param2")
		fmt.Fprint(w, fmt.Sprintf("Param1:=%s, Param2:=%s, URL Param:%s", p1, p2, urlParam))
	}
}

type Param struct {
	Param1 string `json:"param1"`
	Param2 string `json:"param2"`
	Param3 string `json:"param3"`
	Param4 string `json:"param4"`
}

func (s *Service) SetJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Param
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, fmt.Sprintf("JSON Param1:%s, Param2:%s, Param3:%s, Param4:%s", p.Param1, p.Param2, p.Param3, p.Param4))
	}
}

func (s *Service) SetMulti() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Header)
		// stores body max up to 52MB in memory
		err := r.ParseMultipartForm(50 << 20)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		// f := r.MultipartForm
		// ff := r.FormFile
	}
}
