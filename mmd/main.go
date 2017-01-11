package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/simonwittber/middleman"
	influxdb "github.com/vrischmann/go-metrics-influxdb"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8765", "service address")
var trustedKey = flag.String("trustedkey", "xyzzy", "trusted client key")
var superKey = flag.String("superkey", "jabberwocky", "trusted super key, receives all messages")
var untrustedKey = flag.String("publickey", "plugh", "untrusted, public client key")

var upgrader = websocket.Upgrader{}

func handleOutgoing(client *middleman.Client) {
	for {
		select {
		case _, _ = <-client.Quit:
			log.Println("client.Quit")
			return
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
	counter := metrics.GetOrRegisterCounter("Connections", nil)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	log.Println("New WebSocket Connection")
	defer c.Close()
	mt, payload, err := c.ReadMessage()
	if err != nil {
		return
	}
	if mt != websocket.TextMessage {
		log.Println("Not TextMessage")
		return
	}
	client := middleman.Client{}
	client.Conn = c
	client.Outbox = make(chan []byte)
	key := string(payload)
	client.IsTrusted = key == *trustedKey
	log.Println("New Client", client)
	if !client.IsTrusted && key != *untrustedKey {
		log.Println("Bad Key", client)
		return
	}
	client.Conn.WriteMessage(websocket.TextMessage, []byte("MM OK"))
	client.Quit = make(chan bool)
	counter.Inc(1)
	defer counter.Inc(-1)
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
		msg, err := middleman.Unmarshal(payload)
		if err != nil {
			log.Println("Unmarshal:", err)
			break
		}
		msg.Client = &client
		log.Println("MSG:", msg)
		switch msg.Cmd {
		case "PUB":
			go handlePub(msg)
		case "SUB":
			go handleSub(msg)
		case "REQ":
			go handleReq(msg)
		case "RES":
			go handleRes(msg)
		case "EPUB":
			go handleEpub(msg)
		case "ESUB":
			go handleEsub(msg)
		case "EREQ":
			go handleEreq(msg)
		}
	}
	close(client.Quit)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	go influxdb.InfluxDB(
		metrics.DefaultRegistry, // metrics registry
		time.Second*10,          // interval
		"http://localhost:8086", // the InfluxDB url
		"mydbv",                 // your InfluxDB database
		"myuser",                // your InfluxDB user
		"mypassword",            // your InfluxDB password
	)
	http.HandleFunc("/", handleWebSocket)
	upgrader.CheckOrigin = func(request *http.Request) bool { return true }
	log.Println("Starting server on:", *addr)
	log.Println("Trusted key:", *trustedKey)
	log.Println("Untrusted key:", *untrustedKey)
	log.Println("Super key:", *superKey)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
