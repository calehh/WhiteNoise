package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"whitenoise/sdk"
)

type HosterService interface {
	GetRoomName() string
}

type HosterConfig struct {
	port string
}

type BasicHoster struct {
	Name string
	node sdk.Client
	port string
}

func (h *BasicHoster) start() {
	http.HandleFunc("/", HelloHandler)
	err := http.ListenAndServe(h.port, nil)
	if err != nil {
		panic(err)
	}
}

func NewBasicHoster(cfg *HosterConfig) BasicHoster {
	return BasicHoster{port: cfg.port}
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func main() {
	r := bufio.NewReader(os.Stdin)
	go func() {
		for true {
			l,err:= r.ReadString('\n')
			if err != nil{
				if err == io.EOF{
					continue
				}
				panic(err)
			}
			fmt.Println(l)
		}
	}()
	select {
	}
}