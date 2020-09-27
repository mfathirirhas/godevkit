package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8282/get/1/2", nil)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Err: ", err)
	}
	log.Println(string(res))
}
