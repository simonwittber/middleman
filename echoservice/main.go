package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/simonwittber/middleman"
)

var addr = flag.String("addr", "ws://localhost:8765", "service address")
var trustedKey = flag.String("trustedkey", "xyzzy", "trusted client key")
var reconnect = flag.Bool("reconnect", true, "reconnect on server error")

func echo(msg *middleman.Message) {
	msg.Client.Outbox <- middleman.Marshal(msg)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	echoService, err := middleman.NewService(*addr, *trustedKey)
	if err != nil {
		log.Fatal(err)
		return
	}
	echoService.RegisterPubHandler("ECHO", echo)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-sigs:
			log.Println(sig)
			return
		case closed := <-echoService.Closed:
			if closed && *reconnect {
				for {
					err := echoService.Connect()
					if err == nil {
						break
					} else {
						time.Sleep(time.Second)
					}
				}
			} else {
				return
			}
		}
	}

}
