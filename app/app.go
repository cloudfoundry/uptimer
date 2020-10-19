package app

const Source = `package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	go periodicallyLog(1 * time.Second)

	http.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}

func hello(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")
	io.WriteString(res, "<strong>Hello!</strong>")
}

func periodicallyLog(fre time.Duration) {
	ticker := time.NewTicker(fre)
	for {
		select {
		case t := <-ticker.C:
			fmt.Println(t.Unix())
		}
	}
}
`
