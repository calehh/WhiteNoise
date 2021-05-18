package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"whitenoise/common/log"
)

var upgrader = websocket.Upgrader{}

type WsService struct {
	port string
	conn *websocket.Conn
}

func NewWsService(port string) *WsService {
	return &WsService{port: port}
}

func (service *WsService) Start() {
	http.HandleFunc("/ws", service.handleWS)
	err := http.ListenAndServe(service.port, nil)
	if err != nil {
		panic(err)
	}
}

func (ws *WsService) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
	}
	ws.conn = c
}
