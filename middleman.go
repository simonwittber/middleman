package middleman

import (
	"bufio"
	"bytes"
	"errors"
	"log"
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

func Unmarshal(payload []byte) (*Message, error) {
	br := bytes.NewReader(payload)
	rd := bufio.NewReader(br)
	tr := textproto.NewReader(rd)
	msg := Message{}
	line, err := tr.ReadLine()
	if err != nil {
		log.Println("ReadLine:", err)
		return nil, err
	}
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		log.Println("SplitN != 2")
		return nil, errors.New("Parts in top line must be 2")
	}
	msg.Cmd = parts[0]
	msg.Key = parts[1]
	msg.Header, err = tr.ReadMIMEHeader()
	if err != nil {
		log.Println("ReadMIMEHeader", err)
		return nil, err
	}
	msg.Body, err = tr.ReadDotBytes()
	if err != nil {
		log.Println("ReadDotBytes", err)
		return nil, err
	}
	return &msg, nil
}

func Marshal(message *Message) []byte {
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
