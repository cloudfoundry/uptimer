package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("logging...")
	go log(1 * time.Second)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "<strong>Hello!</strong>")
}

func log(fre time.Duration) {
	for {
		time.Sleep(fre)
		fmt.Printf("%d\n", time.Now().Unix())
	}
}
