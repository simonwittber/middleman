package middleman

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/simonwittber/middleman"
)

type MiddlemanClient struct {
	Conn        *websocket.Conn
	PubHandlers *HandlerFuncAtomicMap
	ReqHandlers *HandlerFuncAtomicMap
	Outgoing    chan []byte
	Quit        chan bool
}

// +gen atomicmap
type HandlerFunc func(*middleman.Message)

func (mmc MiddlemanClient) RegisterReqHandler(key string, fn HandlerFunc) {
	mmc.ReqHandlers.Set(key, fn)
}

func (mmc MiddlemanClient) RegisterPubHandler(key string, fn HandlerFunc) {
	mmc.PubHandlers.Set(key, fn)
}

func NewMiddlemanClient(u string, key string) MiddlemanClient {
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	mmc := MiddlemanClient{Conn: c, Outgoing: make(chan []byte), Quit: make(chan bool)}
	mmc.PubHandlers = NewHandlerFuncAtomicMap()
	mmc.ReqHandlers = NewHandlerFuncAtomicMap()
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
	go handleIncoming(mmc)
	go handleOutgoing(mmc)
	return mmc
}

func handleOutgoing(mmc MiddlemanClient) {
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

func handleIncoming(mmc MiddlemanClient) {
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
			mt, payload, err := mmc.Conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if mt != websocket.TextMessage {
				log.Println("Not TextMessage:", mt)
				return
			}
			msg, err := middleman.Unmarshal(payload)
			if err != nil {
				log.Fatalln(err)
				return
			}

			if msg.Cmd == "PUB" {
				fn, ok := mmc.PubHandlers.Get(msg.Key)
				if ok {
					fn(msg)
				}
			}
			if msg.Cmd == "REQ" {
				fn, ok := mmc.ReqHandlers.Get(msg.Key)
				if ok {
					fn(msg)
				}
			}
		}
	}
}
