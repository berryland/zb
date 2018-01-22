package zb

import (
	"github.com/gorilla/websocket"
	"log"
)

const WebSocketServerUrl = "wss://api.zb.com:9999/websocket"

type WebSocketClient struct {
	conn *websocket.Conn
}

func NewWebSocketClient() *WebSocketClient {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(WebSocketServerUrl, nil)
	if err != nil {
		log.Fatalln("Fail to connect to " + WebSocketServerUrl + ", error: " + err.Error())
	}

	go func() {
		for {
			_, bytes, _ := conn.ReadMessage()
			println(string(bytes))
		}
	}()

	return &WebSocketClient{conn: conn}
}

type eventMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
}

func (c *WebSocketClient) SubscribeQuote() {
	c.conn.WriteJSON(eventMessage{Event: "addChannel", Channel: "btcusdt_ticker"})
}
