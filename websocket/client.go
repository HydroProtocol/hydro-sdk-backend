package websocket

import (
	"github.com/satori/go.uuid"
	"net"
	"sync"
)

// For Mock Test
type clientConn interface {
	WriteJSON(interface{}) error
	ReadJSON(interface{}) error
	RemoteAddr() net.Addr
}

type Client struct {
	ID       string
	Conn     clientConn
	Channels map[string]*Channel
	mu       sync.Mutex
}

func (c *Client) sendData(data interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Conn.WriteJSON(data)
}

func (c *Client) Send(data interface{}) error {
	err := c.sendData(data)

	if err != nil {
		return err
	}

	return nil
}

func NewClient() *Client {
	return &Client{
		ID:       uuid.NewV4().String(),
		Channels: make(map[string]*Channel),
	}
}
