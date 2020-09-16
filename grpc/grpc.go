package grpc

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_reuseport "github.com/valyala/fasthttp/reuseport"
	_grpc "google.golang.org/grpc"
	_reflection "google.golang.org/grpc/reflection"
)

type Registerer func(*_grpc.Server, interface{})

type GRPC struct {
	server  *_grpc.Server
	service interface{}
	port    string
	errChan chan error
}

type Opts struct {
	Port    string
	Service interface{}
}

func New(opts *Opts) *GRPC {
	server := _grpc.NewServer()
	_reflection.Register(server)
	return &GRPC{
		server:  server,
		service: opts.Service,
		port:    strings.Replace(opts.Port, ":", "", -1),
	}
}

func (g *GRPC) Register(reg Registerer) {
	reg(g.server, g.service)
}

func (g *GRPC) Run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	lis, err := _reuseport.Listen("tcp4", g.port)
	if err != nil {
		g.errChan <- err
	}
	select {
	case g.errChan <- g.server.Serve(lis):
		log.Printf("grpc: error received from server.")
	case <-sig:
		log.Printf("grpc: terminate signal received, gracefully shut the server...")
		g.server.GracefulStop()
	}
}

func (g *GRPC) ListenError() <-chan error {
	return g.errChan
}
