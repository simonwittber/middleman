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

// Client is used by the MM server to hold information about client
// connections from trusted service providers and untrusted clients.
// +gen * set
type Client struct {
	Conn      *websocket.Conn
	Outbox    chan []byte
	IsTrusted bool
	Quit      chan bool
	GUID      string
	UID       string
	Header    textproto.MIMEHeader
}

// Message represents the protocol used for MM communication over the
// websocket.
type Message struct {
	Cmd    string
	Key    string
	Header textproto.MIMEHeader
	Body   []byte
	Client *Client
}

// Unmarshal turns a []byte into a *Message.
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

// Marshal turns a *Message into a []byte
func Marshal(message *Message) []byte {
	var b bytes.Buffer
	b.Write([]byte(message.Cmd))
	b.Write([]byte(" "))
	b.Write([]byte(message.Key))
	b.Write([]byte("\r\n"))
	if message.Header != nil {
		for k := range message.Header {
			for _, v := range message.Header[k] {
				b.Write([]byte(k))
				b.Write([]byte(": "))
				b.Write([]byte(v))
				b.Write([]byte("\r\n"))
			}
		}
	}

	b.Write([]byte("\r\n"))
	b.Write(message.Body)
	b.WriteString("\r\n.\r\n")
	return b.Bytes()
}
