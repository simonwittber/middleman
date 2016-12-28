package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	Outbox    chan []byte
	IsTrusted bool
	Quit      chan bool
}

type Message struct {
	Cmd    string
	Key    string
	Header textproto.MIMEHeader
	Body   []byte
	Client *Client
}

var addr = flag.String("addr", "localhost:8080", "service address")
var trustedKey = flag.String("key", "xyzzy", "trusted client key")

var upgrader = websocket.Upgrader{}

func handleOutgoing(client *Client) {
	for {
		select {
		case _, _ = <-client.Quit:
			break
		case msg, ok := <-client.Outbox:
			if !ok {
				return
			}
			err := client.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("WriteMessage: ", err)
				return
			}
		}
	}
	log.Println("Sender has finished.")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	mt, payload, err := c.ReadMessage()
	if err != nil {
		return
	}
	if mt != websocket.TextMessage {
		return
	}
	client := Client{}
	client.Conn = c
	client.Outbox = make(chan []byte)
	client.IsTrusted = string(payload) == *trustedKey
	client.Quit = make(chan bool)
	go handleOutgoing(&client)

	for {
		mt, payload, err := c.ReadMessage()
		if err != nil {
			log.Println("ReadMessage:", err)
			break
		}
		if mt != websocket.TextMessage {
			log.Println("!  TextMessage")
			break
		}
		br := bytes.NewReader(payload)
		rd := bufio.NewReader(br)
		tr := textproto.NewReader(rd)
		msg := Message{}
		line, err := tr.ReadLine()
		if err != nil {
			log.Println("ReadLine:", err)
			break
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			log.Println("SplitN != 2")
			break
		}
		msg.Cmd = parts[0]
		msg.Key = parts[1]
		msg.Header, err = tr.ReadMIMEHeader()
		if err != nil {
			log.Println("ReadMIMEHeader", err)
			break
		}
		msg.Body, err = tr.ReadDotBytes()
		if err != nil {
			log.Println("ReadDotBytes", err)
			break
		}
		msg.Client = &client
		log.Println("MSG:", msg)
		go handlers[msg.Cmd](&msg)
	}
	close(client.Quit)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", handleWebSocket)
	upgrader.CheckOrigin = func(request *http.Request) bool { return true }
	log.Println("Starting server on", *addr, "with trusted key", *trustedKey)
	connectHandlerFunctions()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
