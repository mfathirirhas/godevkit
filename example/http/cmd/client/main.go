package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_client "github.com/mfathirirhas/godevkit/http/client"
)

func main() {
	c := _client.New(&_client.Options{
		MaxIdleConns:    15,
		IdleConnTimeout: 5 * time.Second,
		MaxRetry:        5,
	})

	Get(c)
}

func Get(c *_client.Client) {
	getUrl := "http://localhost:8282/get"

	header := make(http.Header)
	header.Set("request-id", "1")
	urlvalues := make(url.Values)
	urlvalues.Set("param1", "1")
	urlvalues.Set("param2", "2")
	urlvalues.Set("param3", "3")
	getRequest := &_client.Request{
		URL:       getUrl,
		Header:    header,
		URLValues: urlvalues,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	resp := c.Get(ctx, getRequest)
	fmt.Println("Error: ", resp.Err())
	str, err := resp.String()
	if err != nil {
		fmt.Println("wew: ", err)
	}
	fmt.Println("String: ", str)
}
