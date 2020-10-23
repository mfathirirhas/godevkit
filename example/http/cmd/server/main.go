package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_router "github.com/mfathirirhas/godevkit/example/http/server"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	r := _router.Init()
	go r.Run()
	select {
	case err := <-r.Err():
		log.Println("HTTP server got error: ", err)
		os.Exit(2)
	case <-sig:
		log.Println("interrupt signal received")
		os.Exit(0)
	}
}
