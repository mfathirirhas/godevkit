package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_client "github.com/mfathirirhas/godevkit/example/grpc/client"
	_service "github.com/mfathirirhas/godevkit/example/grpc/service"
)

func main() {
	c, err := _client.Init()
	if err != nil {
		log.Println("Error Init: ", err)
		os.Exit(1)
	}
	getInput := &_service.GetInput{
		Number: 1,
		Text:   "this is text",
	}
	getOutput, err := c.Get(context.Background(), getInput)
	if err != nil {
		log.Println("error get: ", err)
		os.Exit(1)
	}
	fmt.Println(getOutput)

	setInput := &_service.SetInput{
		A: 1,
		B: "2",
	}
	setOutput, err := c.Set(context.Background(), setInput)
	if err != nil {
		log.Println("error set: ", err)
		os.Exit(1)
	}
	fmt.Println(setOutput)
}
