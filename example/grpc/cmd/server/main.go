package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_server "github.com/mfathirirhas/godevkit/example/grpc/server"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	server := _server.Init()
	go server.Start()
	select {
	case <-sig:
		log.Println("Interup signal received...")
		server.Stop()
		os.Exit(0)
	case err := <-server.Err():
		log.Println("server error: ", err)
		os.Exit(1)
	}
}
