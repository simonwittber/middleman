package main

import (
	"bytes"
	"flag"
	"log"
	"net/textproto"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "ws://localhost:8765/", "middleman service address")
var key = flag.String("key", "xyzzy", "middleman API key")

type MiddlemanClient struct {
	Conn     *websocket.Conn
	Incoming chan []byte
	Outgoing chan []byte
	Quit     chan bool
}

func (mmc MiddlemanClient) Send(cmd string, key string, header textproto.MIMEHeader, body []byte) {
	var buf = bytes.Buffer{}
	top := cmd + " " + key + "\r\n"
	buf.WriteString(top)
	if header != nil {
		for k, vs := range header {
			for _, v := range vs {
				buf.WriteString(k + ": " + v + "\r\n")
			}
		}
	}
	buf.WriteString("\r\n")
	if body != nil {
		buf.Write(body)
	}
	buf.WriteString("\r\n.\r\n")
	mmc.Outgoing <- buf.Bytes()
}

func NewMiddlemanClient(u string, key string) MiddlemanClient {
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	mmc := MiddlemanClient{Conn: c, Incoming: make(chan []byte), Outgoing: make(chan []byte), Quit: make(chan bool)}
	c.WriteMessage(websocket.TextMessage, []byte(key))
	_, msg, err := c.ReadMessage()
	if err != nil {
		log.Println("read: ", err)
	}
	if string(msg) != "MM OK" {
		log.Println("Key not accepted.")
	} else {
		log.Println("Key accepted.")
	}
	go HandleIncoming(mmc)
	go HandleOutgoing(mmc)
	return mmc
}

func HandleOutgoing(mmc MiddlemanClient) {
	for {
		select {
		case _, closed := <-mmc.Quit:
			if closed {
				log.Println("Closing Ougoing")
				return
			}
		case msg := <-mmc.Outgoing:
			mmc.Conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func HandleIncoming(mmc MiddlemanClient) {
	defer mmc.Conn.Close()
	defer close(mmc.Quit)
	for {
		select {
		case _, closed := <-mmc.Quit:
			if closed {
				log.Println("Closing Incoming")
				return
			}
		default:
			mt, message, err := mmc.Conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if mt != websocket.TextMessage {
				log.Println("Not TextMessage:", mt)
				return
			}
			mmc.Incoming <- message
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	mmc := NewMiddlemanClient(*addr, *key)
	mmc.Send("PUB", "BOOYA", nil, nil)
	log.Println(mmc)
	msg := <-mmc.Incoming
	log.Println(msg)
	return
	log.Println("Bye.")

}
