package server

import (
	context "context"

	_service "github.com/mfathirirhas/godevkit/example/grpc/service"
	_grpc "github.com/mfathirirhas/godevkit/grpc/server"
)

type server struct {
	srv *_grpc.Server
}

func Init() *server {
	port := 60001
	svc := &service{}
	srv := _grpc.New(&_grpc.Opts{
		Port: uint16(port),
	})
	_service.RegisterServiceServer(srv.Server(), svc)
	return &server{
		srv: srv,
	}
}

// Start start the server, blocking.
func (s *server) Start() {
	s.srv.Run()
}

func (s *server) Stop() {
	s.srv.Stop()
}

func (s *server) Err() <-chan error {
	return s.srv.ListenError()
}

type service struct {
}

func (s *service) Get(context.Context, *_service.GetInput) (*_service.GetOutput, error) {
	return &_service.GetOutput{
		Success: true,
		Message: "Success",
	}, nil
}

func (s *service) Set(context.Context, *_service.SetInput) (*_service.SetOutput, error) {
	return &_service.SetOutput{
		Float:   2.3,
		Message: "Success",
		Success: true,
	}, nil
}
