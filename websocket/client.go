package websocket

import (
	"github.com/satori/go.uuid"
	"net"
)

// For Mock Test
type ClientConn interface {
	WriteJSON(interface{}) error
	ReadJSON(interface{}) error
	RemoteAddr() net.Addr
}

type Client struct {
	ID       string
	Conn     ClientConn
	Channels map[string]*Channel
}

func (c *Client) Send(data interface{}) error {
	err := c.Conn.WriteJSON(data)

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
