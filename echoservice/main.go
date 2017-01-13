package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simonwittber/middleman"
)

var addr = flag.String("addr", "ws://localhost:8765", "service address")
var trustedKey = flag.String("trustedkey", "xyzzy", "trusted client key")

func echo(msg *middleman.Message) {
	msg.Client.Outbox <- middleman.Marshal(msg)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	var echoService = middleman.NewService(*addr, *trustedKey)
	echoService.RegisterPubHandler("ECHO", echo)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Println(sig)
}
