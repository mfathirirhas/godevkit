package client

import (
	_service "github.com/mfathirirhas/godevkit/example/grpc/service"
	_grpc "github.com/mfathirirhas/godevkit/grpc/stub"
)

type Client struct {
	_service.ServiceClient
}

func Init() (*Client, error) {
	addr := "localhost:60001"
	stub, err := _grpc.New(&_grpc.Options{
		Address:             addr,
		MaxRetry:            20,
		WaitConnectionReady: false,
		ConnectionTimeOut:   3,
	})
	if err != nil {
		return nil, err
	}

	client := _service.NewServiceClient(stub.ClientConn())
	return &Client{client}, nil
}
