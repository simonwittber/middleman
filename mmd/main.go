package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/rs/xid"
	"github.com/simonwittber/middleman"
)

var exposeMetrics = flag.String("metrics", "y", "expose metrics on /debug/metrics")
var addr = flag.String("addr", "0.0.0.0:8765", "service address")
var internalAddr = flag.String("internal", "127.0.0.1:8764", "internal address")
var trustedKey = flag.String("trustedkey", "xyzzy", "trusted client key")
var superKey = flag.String("superkey", "jabberwocky", "trusted super key, receives all messages")
var untrustedKey = flag.String("publickey", "plugh", "untrusted, public client key")

var upgrader = websocket.Upgrader{}

func handleOutgoing(client *middleman.Client) {
	for {
		select {
		case _, _ = <-client.Quit:
			log.Println("client.Quit")
			RemoveClient(client.GUID)
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
	client.GUID = xid.New().String()
	key := string(payload)
	client.IsTrusted = key == *trustedKey
	log.Println("New Client", client)
	if !client.IsTrusted && key != *untrustedKey {
		log.Println("Bad Key", client)
		return
	}
	client.Conn.WriteMessage(websocket.TextMessage, []byte("MM OK"))
	client.Conn.WriteMessage(websocket.TextMessage, []byte(client.GUID))
	AddClient(&client)
	client.Quit = make(chan bool)
	counter.Inc(1)
	defer counter.Inc(-1)
	go handleOutgoing(&client)
	subToPrivateChannel(&client)
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
		msg.Header.Set("cid", msg.Client.GUID)
		msg.Header.Set("uid", msg.Client.UID)
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
		case "INT":
			if client.IsTrusted {
				go handleInt(msg)
			} else {
				log.Println("Untrusted Internal msg", msg)
			}
		}
	}
	log.Println("Bye", client)
	handleClose(&client)
	close(client.Quit)
}

func internalService() {
	err := http.ListenAndServe(*internalAddr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	/*go influxdb.InfluxDB(
		metrics.DefaultRegistry, // metrics registry
		time.Second*10,          // interval
		"http://localhost:8086", // the InfluxDB url
		"mydbv",                 // your InfluxDB database
		"myuser",                // your InfluxDB user
		"mypassword",            // your InfluxDB password
	)*/
	http.HandleFunc("/", handleWebSocket)
	upgrader.CheckOrigin = func(request *http.Request) bool { return true }
	log.Println("Starting server on:", *addr)
	log.Println("Trusted key:", *trustedKey)
	log.Println("Untrusted key:", *untrustedKey)
	log.Println("Super key:", *superKey)
	if *exposeMetrics == "y" {
		//This exposes metrics on /debug/metrics
		exp.Exp(metrics.DefaultRegistry)
		log.Println("Exposing metrics: http://" + *addr + "/debug/metrics")
	}
	go internalService()
	err := http.ListenAndServeTLS(*addr, "server.crt", "server.key", nil)
	if err != nil {
		log.Fatalln(err)
	}
}


