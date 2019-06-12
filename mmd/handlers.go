package main

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/simonwittber/go-string-set"
	"github.com/simonwittber/middleman"
)

type SubscriptionMap map[*middleman.Client]atomicstring.StringSet

var safePubKeys = atomicstring.NewStringSet()
var safeSubKeys = atomicstring.NewStringSet()
var subscribers = middleman.NewClientSetAtomicMap()
var responders = make(map[string]chan []byte)
var respondersMutex sync.Mutex
var requestId uint64 = 0
var requests = make(map[uint64]*middleman.Client)
var requestMutex sync.Mutex
var subMutex sync.Mutex
var ClientMap = make(map[string]*middleman.Client)
var subscriptions = make(SubscriptionMap)

func RemoveClient(guid string) {
	var _, ok = ClientMap[guid]
	if ok {
		delete(ClientMap, guid)
	}
}

func AddClient(client *middleman.Client) {
	ClientMap[client.GUID] = client
}

func getBroadcastChannel(key string) middleman.ClientSet {
	bc, ok := subscribers.Get(key)
	if !ok || bc == nil {
		bc = middleman.NewClientSet()
		subscribers.Set(key, bc)
	}
	return bc
}

func addSubscription(client *middleman.Client, key string) bool {
	subMutex.Lock()
	defer subMutex.Unlock()
	keys, ok := subscriptions[client]
	if !ok {
		keys = atomicstring.NewStringSet()
		subscriptions[client] = keys
	}
	if !keys.Contains(key) {
		keys.Add(key)
	}
	return true
}

func subToPrivateChannel(client *middleman.Client) {
	var key = "MSG:" + client.GUID
	if addSubscription(client, key) {
		bc := getBroadcastChannel(key)
		log.Println("Subscribing to", key)
		bc.Add(client)
	}
}

func handleSub(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.SUB", nil)
	timer.Time(func() {
		if !messageIsTrusted(message, safeSubKeys) {
			sendError(message.Client, "Not trusted:", message)
		}
		if addSubscription(message.Client, message.Key) {
			bc := getBroadcastChannel(message.Key)
			log.Println("Subscribing to", message.Key)
			bc.Add(message.Client)
		}
	})
}

func handleInt(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.INT", nil)
	timer.Time(func() {
		log.Println("INT", message.Header.Get("setuid"))
		switch message.Key {
		case "UID":
			var cid = message.Header.Get("forcid")
			var uid = message.Header.Get("setuid")
			ClientMap[cid].UID = uid
		}
	})
}

func handlePub(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.PUB", nil)
	timer.Time(func() {
		if !messageIsTrusted(message, safePubKeys) {
			sendError(message.Client, "Not trusted", message)
		}
		bc := getBroadcastChannel(message.Key)
		bytes := middleman.Marshal(message)
		for c := range bc.Iter() {
			c.Outbox <- bytes
		}
	})
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
	timer := metrics.GetOrRegisterTimer("handlers.REQ", nil)
	timer.Time(func() {
		if !messageIsTrusted(message, safePubKeys) {
			sendError(message.Client, "Not trusted", message)
		}
		respondersMutex.Lock()
		responder, ok := responders[message.Key]
		respondersMutex.Unlock()
		if !ok || responder == nil {
			sendError(message.Client, "No responder for", message)
		} else {
			reqID := atomic.AddUint64(&requestId, 1)
			message.Header.Set("rid", strconv.FormatUint(reqID, 10))
			requestMutex.Lock()
			requests[reqID] = message.Client
			requestMutex.Unlock()
			responder <- middleman.Marshal(message)
		}
	})
}

func handleRes(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.RES", nil)
	timer.Time(func() {
		if !messageIsTrusted(message, safePubKeys) {
			sendError(message.Client, "Not trusted", message)
		}
		reqID, err := strconv.ParseUint(message.Header.Get("rid"), 10, 64)
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
			sendError(message.Client, "No client for request ID ", message)
		} else {
			message.Header.Del("rid")
			client.Outbox <- middleman.Marshal(message)
		}
	})
}

func handleEreq(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.EREQ", nil)
	timer.Time(func() {
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
			sendError(message.Client, "Not trusted", message)
		}
	})
}

func handleEpub(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.EPUB", nil)
	timer.Time(func() {
		if message.Client.IsTrusted {
			log.Println("Enabling Publish for Key: " + message.Key)
			safePubKeys.Add(message.Key)
		} else {
			sendError(message.Client, "Not trusted", message)
		}
	})
}

func handleEsub(message *middleman.Message) {
	timer := metrics.GetOrRegisterTimer("handlers.ESUB", nil)
	timer.Time(func() {
		if message.Client.IsTrusted {
			safeSubKeys.Add(message.Key)
		} else {
			sendError(message.Client, "Not trusted", message)
		}
	})
}

func sendError(client *middleman.Client, txt string, msg *middleman.Message) {
	m := middleman.Message{Cmd: "PUB", Key: "ERROR", Body: []byte(txt + " " + msg.Cmd + " " + msg.Key)}
	client.Outbox <- middleman.Marshal(&m)
}

func handleClose(client *middleman.Client) {
	subMutex.Lock()
	defer subMutex.Unlock()
	keys, ok := subscriptions[client]
	log.Println("Disconnect", subscriptions, client, keys, ok)
	if ok {
		for k := range keys.Iter() {
			c, ok := subscribers.Get(k)
			if ok {
				c.Remove(client)
			}
		}
	}
	delete(subscriptions, client)
}
