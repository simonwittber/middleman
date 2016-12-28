package main

import (
	"bytes"
	"log"
	"strconv"
	"sync"

	"github.com/tjgq/broadcast"
)

var safePubKeys = make(map[string]bool)
var safeSubKeys = make(map[string]bool)
var reqMux, idMux, resMux, subMux, keyMux sync.Mutex
var handlers = make(map[string]func(*Message))
var subscribers = make(map[string]*broadcast.Broadcaster)
var responders = make(map[string]chan []byte)
var requestId uint64 = 0
var requests = make(map[uint64]*Client)

func getBroadcastChannel(key string) *broadcast.Broadcaster {
	subMux.Lock()
	bc, exists := subscribers[key]
	if !exists {
		bc = broadcast.New(8)
		subscribers[key] = bc
	}
	subMux.Unlock()
	return bc
}

func handleSub(message *Message) {
	if !messageIsTrusted(message, safeSubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	bc := getBroadcastChannel(message.Key).Listen()
	log.Println("Subscribing to", message.Key)
	for {
		select {
		case m, ok := <-bc.Ch:
			if !ok {
				return
			}
			log.Println(string(m.([]byte)))
			message.Client.Outbox <- m.([]byte)
		case _, _ = <-message.Client.Quit:
			return
		}
	}
}

func marshalMessage(message *Message) []byte {
	var b bytes.Buffer
	b.Write([]byte(message.Cmd))
	b.Write([]byte(" "))
	b.Write([]byte(message.Key))
	b.Write([]byte("\r\n"))
	for k := range message.Header {
		for _, v := range message.Header[k] {
			b.Write([]byte(k))
			b.Write([]byte(": "))
			b.Write([]byte(v))
			b.Write([]byte("\r\n"))
		}
	}
	b.Write([]byte("\r\n"))
	b.Write(message.Body)
	return b.Bytes()
}

func handlePub(message *Message) {
	if !messageIsTrusted(message, safeSubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	bc := getBroadcastChannel(message.Key)
	bc.Send(marshalMessage(message))
}

func messageIsTrusted(message *Message, safeKeys map[string]bool) bool {
	if message.Client.IsTrusted {
		return true
	}
	keyMux.Lock()
	defer keyMux.Unlock()
	if val, ok := safeKeys[message.Key]; val && ok {
		return true
	}
	return false
}

func handleReq(message *Message) {
	if !messageIsTrusted(message, safePubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	resMux.Lock()
	responder, ok := responders[message.Key]
	resMux.Unlock()
	if !ok {
		sendError(message.Client, "No responder for: "+message.Key)
	} else {
		reqMux.Lock()
		message.Header.Set("ReqID", string(requestId))
		requests[requestId] = message.Client
		requestId += 1
		reqMux.Unlock()
		responder <- marshalMessage(message)
	}
}

func handleRes(message *Message) {
	if !messageIsTrusted(message, safePubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}

	reqID, err := strconv.ParseUint(message.Header.Get("ReqID"), 10, 64)
	if err != nil {
		return
	}
	reqMux.Lock()
	client, ok := requests[reqID]
	if !ok {
		sendError(message.Client, "No client for request ID ")
	} else {
		message.Header.Del("ReqID")
		client.Outbox <- marshalMessage(message)
	}
	reqMux.Unlock()
}

func handleEreq(message *Message) {
	if message.Client.IsTrusted {
		resMux.Lock()
		c, ok := responders[message.Key]
		if !ok {
			c = make(chan []byte)
			responders[message.Key] = c
		}
		resMux.Unlock()
		for {
			select {
			case m, ok := <-c:
				if !ok {
					return
				}
				message.Client.Outbox <- m
			case _, _ = <-message.Client.Quit:
				return
			}
		}
	} else {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
}

func handleEpub(message *Message) {
	if message.Client.IsTrusted {
		keyMux.Lock()
		safePubKeys[message.Key] = true
		keyMux.Unlock()
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func handleEsub(message *Message) {
	if message.Client.IsTrusted {
		keyMux.Lock()
		safeSubKeys[message.Key] = true
		keyMux.Unlock()
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func sendError(client *Client, txt string) {
	msg := Message{Cmd: "PUB", Key: "ERROR", Body: []byte(txt)}
	client.Outbox <- marshalMessage(&msg)
}

func connectHandlerFunctions() {
	handlers["PUB"] = handlePub
	handlers["SUB"] = handleSub
	handlers["REQ"] = handleReq
	handlers["RES"] = handleRes
	handlers["EPUB"] = handleEpub
	handlers["ESUB"] = handleEsub
	handlers["EREQ"] = handleEreq
}
