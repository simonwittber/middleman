package main

import (
	"bytes"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/antlinker/go-cmap"
	"github.com/simonwittber/go-string-set"
	"github.com/tjgq/broadcast"
)

var safePubKeys = atomicstring.NewStringSet()
var safeSubKeys = atomicstring.NewStringSet()
var subscribers = cmap.NewConcurrencyMap()
var responders = cmap.NewConcurrencyMap()
var requestId uint64 = 0
var requests = cmap.NewConcurrencyMap()

func getBroadcastChannel(key string) *broadcast.Broadcaster {
	var bc *broadcast.Broadcaster
	obj, err := subscribers.Get(key)
	if err != nil || obj == nil {
		bc = broadcast.New(8)
		subscribers.Set(key, bc)
	} else {
		bc = obj.(*broadcast.Broadcaster)
	}
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
	log.Println("X")
	bc := getBroadcastChannel(message.Key)
	bc.Send(marshalMessage(message))
}

func messageIsTrusted(message *Message, safeKeys atomicstring.StringSet) bool {
	if message.Client.IsTrusted {
		return true
	}
	if safeKeys.Contains(message.Key) {
		return true
	}
	return false
}

func handleReq(message *Message) {
	if !messageIsTrusted(message, safePubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	obj, err := responders.Get(message.Key)
	if err != nil || obj == nil {
		sendError(message.Client, "No responder for: "+message.Key)
	} else {
		reqID := atomic.AddUint64(&requestId, 1)
		message.Header.Set("ReqID", string(reqID))
		requests.Set(reqID, message.Client)
		responder := obj.(chan []byte)
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
	obj, err := requests.Remove(reqID)
	if err != nil {
		sendError(message.Client, "No client for request ID ")
	} else {
		message.Header.Del("ReqID")
		client := obj.(Client)
		client.Outbox <- marshalMessage(message)
	}
}

func handleEreq(message *Message) {
	if message.Client.IsTrusted {
		var c chan []byte
		obj, err := responders.Get(message.Key)
		if err != nil || obj == nil {
			c = make(chan []byte)
			responders.Set(message.Key, c)
		} else {
			c = obj.(chan []byte)
		}
		safePubKeys.Add(message.Key)
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
		safePubKeys.Add(message.Key)
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func handleEsub(message *Message) {
	if message.Client.IsTrusted {
		safeSubKeys.Add(message.Key)
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func sendError(client *Client, txt string) {
	msg := Message{Cmd: "PUB", Key: "ERROR", Body: []byte(txt)}
	client.Outbox <- marshalMessage(&msg)
}
