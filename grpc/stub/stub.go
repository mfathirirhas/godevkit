package stub

import (
	"context"
	"log"
	"time"

	_grpc "google.golang.org/grpc"
	_conn "google.golang.org/grpc/connectivity"
)

type Stub struct {
	client *_grpc.ClientConn
}

type Options struct {
	// Address of grpc server, i.e. 10.10.10.10:234434
	Address string

	// IsTLSEnabled if true then authentication to server needed with cert.
	IsTLSEnabled bool

	// WaitConnectionReady wait till connection ready. Blocking.
	// If true then will block for as long as the ConnectionTimeOut.
	WaitConnectionReady bool

	// MaxRetry if WaitConnectionReady is true then will be ignored.
	// if WaitConnectionReady is false then will be used in background for retrying mechanism.
	MaxRetry int

	// ConnectionTimeOut(seconds) If WaitConnectionReady is true then will be used as time to wait till connected.
	// if WaitConnectionReady is false then will be used as time to wait in WaitForStateChange while retrying.
	ConnectionTimeOut int
}

func New(o *Options) (*Stub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(int64(o.ConnectionTimeOut))*time.Second)
	defer cancel()

	// setting dial options
	dialOpts := []_grpc.DialOption{}
	if o.IsTLSEnabled {
		// TODO add tls support
	} else {
		dialOpts = append(dialOpts, _grpc.WithInsecure())
	}
	if o.WaitConnectionReady {
		dialOpts = append(dialOpts, _grpc.WithBlock())
	}

	conn, err := _grpc.DialContext(ctx, o.Address, dialOpts...)
	if err != nil {
		if conn != nil {
			defer conn.Close()
		}
		return nil, err
	}

	stub := &Stub{
		client: conn,
	}

	retryCount := 0
	if !o.WaitConnectionReady {
		go func(con *_grpc.ClientConn) {
			for retryCount < o.MaxRetry {
				if retry(ctx, con) {
					return
				}
				retryCount++
			}
			log.Println("grpc: exceed max retry, failed to reconnect.")
		}(conn)
	}

	return stub, nil
}

func retry(ctx context.Context, con *_grpc.ClientConn) bool {
	state := con.GetState()
	if state != _conn.Ready {
		log.Println("grpc: connection is not ready yet, waiting for state change")
		if !con.WaitForStateChange(ctx, state) {
			log.Println("grpc: reconnecting is failed, retrying...")
			return false
		} else {
			log.Println("grpc: connection is ready")
			return true
		}
	} else {
		log.Println("grpc: connection is ready")
		return true
	}
}

func (s *Stub) GetClient() *_grpc.ClientConn {
	return s.client
}

func (s *Stub) IsClientReady() bool {
	return s.client.GetState() == _conn.Ready
}

func (s *Stub) GetServerAddress() string {
	return s.client.Target()
}

func (s *Stub) Close() error {
	return s.client.Close()
}
