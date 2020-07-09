package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"sync"
	"time"
)

const (
	TextMessage = websocket.TextMessage

	BinaryMessage = websocket.BinaryMessage

	CloseMessage = websocket.CloseMessage

	PingMessage = websocket.PingMessage

	PongMessage = websocket.PongMessage
)

type out struct {
	messageType int
	data        []byte
}

type Connection struct {
	conn      *websocket.Conn
	inChan    chan []byte
	outChan   chan *out
	closeChan chan struct{}
	once      *sync.Once
}

func dial(api string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(api, nil)
	return conn, err
}

func Dial(api string) (*Connection, error) {
	conn, err := dial(api)
	if err != nil {
		return nil, err
	}

	connection := &Connection{
		conn:      conn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan *out, 1000),
		closeChan: make(chan struct{}),
		once:      &sync.Once{},
	}
	go connection.listen()
	return connection, nil
}

// Handles websocket requests from peers
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*Connection, error) {
	// TODO config
	//upgrader.ReadBufferSize = ?
	//upgrader.WriteBufferSize = ?

	conn, err := upgrader.Upgrade(w, r, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		log.Warn("[WebsocketUpgrade] not a websocket handshake")
		return nil, ecode.ErrBadRequest
	} else if err != nil {
		log.Error("[WebsocketUpgrade] failed to upgrade ", "error", err)
		return nil, ecode.ErrInternalServer
	}

	connection := &Connection{
		conn:      conn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan *out, 1000),
		closeChan: make(chan struct{}),
		once:      &sync.Once{},
	}
	go connection.listen()
	return connection, nil
}

func (c *Connection) listen() {
	// 启动读协程
	go c.readLoop()
	// 启动写协程
	c.writeLoop()
}

func (c *Connection) SetReadLimit(limit int64) {
	c.conn.SetReadLimit(limit)
}

func (c *Connection) SetReadDeadline(t time.Time) {
	_ = c.conn.SetReadDeadline(t)
}

func (c *Connection) SetPongHandler(h func(appData string) error) {
	c.conn.SetPongHandler(h)
}

func (c *Connection) ReadMessage() ([]byte, error) {
	select {
	case data := <-c.inChan:
		return data, nil
	case <-c.closeChan:
		return nil, errors.New("connection is closed")
	}
}

func (c *Connection) WriteTextMessage(data []byte) error {
	return c.WriteMessage(TextMessage, data)
}

func (c *Connection) WriteBinaryMessage(data []byte) error {
	return c.WriteMessage(BinaryMessage, data)
}

func (c *Connection) WriteMessage(messageType int, data []byte) error {
	select {
	case c.outChan <- &out{messageType: messageType, data: data}:
		return nil
	case <-c.closeChan:
		return errors.New("connection is closed")
	}
}

func (c *Connection) Close() {
	c.once.Do(func() {
		_ = c.conn.Close()
		close(c.closeChan)
	})
}

func (c *Connection) readLoop() {
	defer c.Close()

	for {
		_, in, err := c.conn.ReadMessage()
		if err != nil {
			log.Error("[ReadLoop] websocket read message", "err", err)
			return
		}

		select {
		case c.inChan <- in:
		case <-c.closeChan:
			return
		}
	}
}

func (c *Connection) writeLoop() {
	defer c.Close()

	for {
		select {
		case out := <-c.outChan:
			if err := c.conn.WriteMessage(out.messageType, out.data); err != nil {
				log.Error("[WriteLoop] websocket write message", "err", err)
				return
			}
		case <-c.closeChan:
			return
		}
	}
}
