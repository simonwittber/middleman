PACKAGE DOCUMENTATION

package middleman
    import "."


FUNCTIONS

func Marshal(message *Message) []byte
    Marshal turns a *Message into a []byte

TYPES

type Client struct {
    Conn      *websocket.Conn
    Outbox    chan []byte
    IsTrusted bool
    Quit      chan bool
}
    Client is used by the MM server to hold information about client
    connections from trusted service providers and untrusted clients.

type HandlerFunc func(*Message)
    +gen atomicmap

type HandlerFuncAtomicMap struct {
    // contains filtered or unexported fields
}
    HandlerFuncAtomicMap is a copy-on-write thread-safe map of HandlerFunc

func NewHandlerFuncAtomicMap() *HandlerFuncAtomicMap
    NewHandlerFuncAtomicMap returns a new initialized HandlerFuncAtomicMap

func (am *HandlerFuncAtomicMap) Delete(key string)
    Delete removes the HandlerFunc under key from the map

func (am *HandlerFuncAtomicMap) Get(key string) (value HandlerFunc, ok bool)
    Get returns a HandlerFunc for a given key

func (am *HandlerFuncAtomicMap) GetAll() map[string]HandlerFunc
    GetAll returns the underlying map of HandlerFunc this map must NOT be
    modified, to change the map safely use the Set and Delete functions and
    Get the value again

func (am *HandlerFuncAtomicMap) Len() int
    Len returns the number of elements in the map

func (am *HandlerFuncAtomicMap) Set(key string, value HandlerFunc)
    Set inserts in the map a HandlerFunc under a given key

type Message struct {
    Cmd    string
    Key    string
    Header textproto.MIMEHeader
    Body   []byte
    Client *Client
}
    Message represents the protocol used for MM communication over the
    websocket.

func Unmarshal(payload []byte) (*Message, error)
    Unmarshal turns a []byte into a *Message.

type Service struct {
    // contains filtered or unexported fields
}
    Service is the connection from a service provider to the MM server.

func NewService(u string, key string) Service
    NewService creates a connection to an MM server using a websocket URL,
    eg ws://localhost:8765/

func (mmc Service) RegisterPubHandler(key string, fn HandlerFunc)
    Register a HandlerFunc to be called when PUB is received.

func (mmc Service) RegisterReqHandler(key string, fn HandlerFunc)
    Register a HandlerFunc to be called when REQ is received.

func (mmc Service) Send(msg *Message)
    Send a message from the service to the MM server.

func (mmc Service) Stop()
    Stop the service and disconnect,

SUBDIRECTORIES

	clients
	mmd

