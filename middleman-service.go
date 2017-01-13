package middleman

import (
	"log"

	"github.com/gorilla/websocket"
)

// Service is the connection from a service provider to the MM server.
type Service struct {
	conn        *websocket.Conn
	pubHandlers *HandlerFuncAtomicMap
	reqHandlers *HandlerFuncAtomicMap
	outgoing    chan []byte
	quit        chan bool
}

// +gen atomicmap
type HandlerFunc func(*Message)

// Register a HandlerFunc to be called when REQ is received.
func (mmc Service) RegisterReqHandler(key string, fn HandlerFunc) {
	mmc.reqHandlers.Set(key, fn)
	mmc.outgoing <- Marshal(&Message{Cmd: "EREQ", Key: key})
}

// Register a HandlerFunc to be called when PUB is received.
func (mmc Service) RegisterPubHandler(key string, fn HandlerFunc) {
	mmc.pubHandlers.Set(key, fn)
	mmc.outgoing <- Marshal(&Message{Cmd: "EPUB", Key: key})
}

// Stop the service and disconnect,
func (mmc Service) Stop() {
	close(mmc.quit)
}

// Send a message from the service to the MM server.
func (mmc Service) Send(msg *Message) {
	mmc.outgoing <- Marshal(msg)
}

// NewService creates a connection to an MM server using a websocket
// URL, eg ws://localhost:8765/
func NewService(u string, key string) Service {
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	mmc := Service{conn: c, outgoing: make(chan []byte), quit: make(chan bool)}
	mmc.pubHandlers = NewHandlerFuncAtomicMap()
	mmc.reqHandlers = NewHandlerFuncAtomicMap()
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

func handleOutgoing(mmc Service) {
	for {
		select {
		case _, closed := <-mmc.quit:
			if closed {
				log.Println("Closing Ougoing")
				return
			}
		case msg := <-mmc.outgoing:
			mmc.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func handleIncoming(mmc Service) {
	defer mmc.conn.Close()
	defer close(mmc.quit)
	for {
		select {
		case _, closed := <-mmc.quit:
			if closed {
				log.Println("Closing Incoming")
				return
			}
		default:
			mt, payload, err := mmc.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if mt != websocket.TextMessage {
				log.Println("Not TextMessage:", mt)
				return
			}
			msg, err := Unmarshal(payload)
			if err != nil {
				log.Fatalln(err)
				return
			}

			if msg.Cmd == "PUB" {
				fn, ok := mmc.pubHandlers.Get(msg.Key)
				if ok {
					fn(msg)
				}
			}
			if msg.Cmd == "REQ" {
				fn, ok := mmc.reqHandlers.Get(msg.Key)
				if ok {
					fn(msg)
				}
			}
		}
	}
}
