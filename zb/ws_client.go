package zb

import (
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"github.com/buger/jsonparser"
)

const WebSocketServerUrl = "wss://api.zb.com:9999/websocket"

type WebSocketClient struct {
	running   bool
	conn      *websocket.Conn
	decoders  map[string]func([]byte) interface{}
	callbacks map[string]func(interface{})
}

func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{running: false, decoders: make(map[string]func([]byte) interface{}), callbacks: make(map[string]func(interface{}))}
}

type eventMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
}

func (c *WebSocketClient) Connect() {
	if c.running {
		return
	}
	c.running = true

	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(WebSocketServerUrl, nil)
	c.conn = conn
	if err != nil {
		c.Disconnect()
		log.Fatalln("Fail to connect to " + WebSocketServerUrl + ", error: " + err.Error())
	}

	go func() {
		defer c.Disconnect()
		for {
			_, bytes, err := c.conn.ReadMessage()
			if err != nil {
				break
			}

			channel, _ := jsonparser.GetString(bytes, "channel")
			if decoder, ok := c.decoders[channel]; ok {
				value := decoder(bytes)
				if callback, ok := c.callbacks[channel]; ok {
					callback(value)
				}
			}
		}
	}()
}

func (c *WebSocketClient) Disconnect() {
	if !c.running {
		return
	}
	c.running = false

	c.conn.Close()
}

func (c *WebSocketClient) SubscribeQuote(symbol string, callback func(quote Quote)) {
	channel := strings.Replace(symbol, "_", "", 1) + "_ticker"
	c.register(channel, func(value []byte) interface{} {
		return marshalQuote(value)
	}, func(v interface{}) {
		callback(v.(Quote))
	})
	c.conn.WriteJSON(eventMessage{Event: "addChannel", Channel: channel})
}

func (c *WebSocketClient) register(channel string, decoder func(value []byte) interface{}, callback func(interface{})) {
	c.registerDecoder(channel, decoder)
	c.registerCallback(channel, callback)
}

func (c *WebSocketClient) registerDecoder(channel string, decoder func(value []byte) interface{}) {
	c.decoders[channel] = decoder
}

func (c *WebSocketClient) registerCallback(channel string, callback func(interface{})) {
	c.callbacks[channel] = callback
}
