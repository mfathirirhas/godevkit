package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	_client "github.com/mfathirirhas/godevkit/http/client"
)

func main() {
	c := _client.New(&_client.Opts{
		MaxIdleConns:    15,
		IdleConnTimeout: 5 * time.Second,
		MaxRetry:        5,
	})

	Get(c)
	PostForm(c)
	PostJSON(c)

	GetInstant()
	PostFormInstant()
	PostJSONInstant()
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
		BaseURL:   getUrl,
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

type RespJSON struct {
	Param1 string `json:"param1"`
	Param2 string `json:"param2"`
}

func PostForm(c *_client.Client) {
	postUrl := "http://localhost:8282/post2"
	header := make(http.Header)
	m := make(map[string]string)
	m["param1"] = "123"
	m["param2"] = "456"
	ctx := context.Background()
	req := &_client.Request{
		BaseURL: postUrl,
		Header:  header,
		Body:    m,
	}
	r := RespJSON{}
	resp := c.PostForm(ctx, req)
	if resp.Err() != nil {
		fmt.Println("resp Error: ", resp.Err())
		return
	}
	str, err := resp.String()
	if err != nil {
		fmt.Println("string error: ", err)
		return
	}
	fmt.Printf("string: %s\n", str)
	if err = resp.Scan(&r); err != nil {
		fmt.Println("scan error: ", err)
		return
	}
	fmt.Println("scan: ", r)
}

func PostJSON(c *_client.Client) {
	postUrl := "http://localhost:8282/post/json"
	m := make(map[string]string)
	m["param1"] = "123"
	m["param2"] = "456"
	m["param3"] = "blabla"
	m["param4"] = "lkwmef"
	postJson := &_client.Request{
		BaseURL: postUrl,
		Body:    m,
	}
	ctx := context.Background()
	resp := c.PostJSON(ctx, postJson)
	if resp.Err() != nil {
		log.Println("Error: ", resp.Err())
		return
	}
	str, err := resp.String()
	if err != nil {
		log.Println("Errorr: ", err)
	}
	fmt.Println(str)
}

func GetInstant() {
	getUrl := "http://localhost:8282/get"

	header := make(http.Header)
	header.Set("request-id", "1")
	urlvalues := make(url.Values)
	urlvalues.Set("param1", "1")
	urlvalues.Set("param2", "2")
	urlvalues.Set("param3", "3")
	getRequest := &_client.Request{
		BaseURL:   getUrl,
		Header:    header,
		URLValues: urlvalues,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	resp := _client.Get(ctx, getRequest)
	fmt.Println("Error: ", resp.Err())
	str, err := resp.String()
	if err != nil {
		fmt.Println("wew: ", err)
	}
	fmt.Println("String: ", str)
}

func PostFormInstant() {
	postUrl := "http://localhost:8282/post2"
	header := make(http.Header)
	m := make(map[string]string)
	m["param1"] = "123"
	m["param2"] = "456"
	ctx := context.Background()
	req := &_client.Request{
		BaseURL: postUrl,
		Header:  header,
		Body:    m,
	}
	r := RespJSON{}
	resp := _client.PostForm(ctx, req)
	if resp.Err() != nil {
		fmt.Println("resp Error: ", resp.Err())
		return
	}
	str, err := resp.String()
	if err != nil {
		fmt.Println("string error: ", err)
		return
	}
	fmt.Printf("string: %s\n", str)
	if err = resp.Scan(&r); err != nil {
		fmt.Println("scan error: ", err)
		return
	}
	fmt.Println("scan: ", r)
}

func PostJSONInstant() {
	postUrl := "http://localhost:8282/post/json"
	m := make(map[string]string)
	m["param1"] = "123"
	m["param2"] = "456"
	m["param3"] = "blabla"
	m["param4"] = "lkwmef"
	postJson := &_client.Request{
		BaseURL: postUrl,
		Body:    m,
	}
	ctx := context.Background()
	resp := _client.PostJSON(ctx, postJson)
	if resp.Err() != nil {
		log.Println("Error: ", resp.Err())
		return
	}
	str, err := resp.String()
	if err != nil {
		log.Println("Errorr: ", err)
	}
	fmt.Println(str)
}
