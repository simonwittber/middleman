package main

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/simonwittber/go-string-set"
	"github.com/simonwittber/middleman"
	"github.com/tjgq/broadcast"
)

var safePubKeys = atomicstring.NewStringSet()
var safeSubKeys = atomicstring.NewStringSet()
var subscribers = make(map[string]*broadcast.Broadcaster)
var subscribersMutex sync.Mutex
var responders = make(map[string]chan []byte)
var respondersMutex sync.Mutex
var requestId uint64 = 0
var requests = make(map[uint64]*middleman.Client)
var requestMutex sync.Mutex

func getBroadcastChannel(key string) *broadcast.Broadcaster {
	subscribersMutex.Lock()
	defer subscribersMutex.Unlock()
	var bc *broadcast.Broadcaster
	bc, ok := subscribers[key]
	if !ok || bc == nil {
		bc = broadcast.New(8)
		subscribers[key] = bc
	}
	return bc
}

func handleSub(message *middleman.Message) {
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

func handlePub(message *middleman.Message) {
	if !messageIsTrusted(message, safeSubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	bc := getBroadcastChannel(message.Key)
	bc.Send(middleman.Marshal(message))
}

func messageIsTrusted(message *middleman.Message, safeKeys atomicstring.StringSet) bool {
	if message.Client.IsTrusted {
		return true
	}
	if safeKeys.Contains(message.Key) {
		return true
	}
	return false
}

func handleReq(message *middleman.Message) {
	if !messageIsTrusted(message, safePubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}
	respondersMutex.Lock()
	responder, ok := responders[message.Key]
	respondersMutex.Unlock()
	if !ok || responder == nil {
		sendError(message.Client, "No responder for: "+message.Key)
	} else {
		reqID := atomic.AddUint64(&requestId, 1)
		message.Header.Set("ReqID", string(reqID))
		requestMutex.Lock()
		requests[reqID] = message.Client
		requestMutex.Unlock()
		responder <- middleman.Marshal(message)
	}
}

func handleRes(message *middleman.Message) {
	if !messageIsTrusted(message, safePubKeys) {
		sendError(message.Client, "Not trusted:"+message.Key)
	}

	reqID, err := strconv.ParseUint(message.Header.Get("ReqID"), 10, 64)
	if err != nil {
		return
	}
	requestMutex.Lock()
	client, ok := requests[reqID]
	if ok {
		delete(requests, reqID)
	}
	requestMutex.Unlock()
	if !ok {
		sendError(message.Client, "No client for request ID ")
	} else {
		message.Header.Del("ReqID")
		client.Outbox <- middleman.Marshal(message)
	}
}

func handleEreq(message *middleman.Message) {
	if message.Client.IsTrusted {
		var c chan []byte
		respondersMutex.Lock()
		c, ok := responders[message.Key]
		if !ok || c == nil {
			c = make(chan []byte)
			responders[message.Key] = c
		}
		respondersMutex.Unlock()
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

func handleEpub(message *middleman.Message) {
	if message.Client.IsTrusted {
		safePubKeys.Add(message.Key)
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func handleEsub(message *middleman.Message) {
	if message.Client.IsTrusted {
		safeSubKeys.Add(message.Key)
	} else {
		sendError(message.Client, "Not trusted.")
	}
}

func sendError(client *middleman.Client, txt string) {
	msg := middleman.Message{Cmd: "PUB", Key: "ERROR", Body: []byte(txt)}
	client.Outbox <- middleman.Marshal(&msg)
}
