package server

import (
	"fmt"

	_reuseport "github.com/valyala/fasthttp/reuseport"
	_grpc "google.golang.org/grpc"
	_reflection "google.golang.org/grpc/reflection"
)

type Server struct {
	srv     *_grpc.Server
	port    uint16
	errChan chan error
}

type Opts struct {
	Port uint16
}

func New(opts *Opts) *Server {
	srv := _grpc.NewServer()
	return &Server{
		srv:  srv,
		port: opts.Port,
	}
}

func (g *Server) Server() *_grpc.Server {
	return g.srv
}

func (g *Server) Run() {
	_reflection.Register(g.srv)
	lis, err := _reuseport.Listen("tcp4", fmt.Sprintf(":%d", g.port))
	if err != nil {
		g.errChan <- err
	}
	g.errChan <- g.srv.Serve(lis)
}

func (g *Server) Stop() {
	g.srv.GracefulStop()
}

func (g *Server) ListenError() <-chan error {
	return g.errChan
}
