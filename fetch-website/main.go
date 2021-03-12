package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)


func main() {
	log.Println("Start fetching website...")
	cli := &http.Client {
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", "http://wendu.cn", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("test.html", buf, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done")
}
