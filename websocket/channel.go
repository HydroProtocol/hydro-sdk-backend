package websocket

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"strings"
	"sync"
)

// Channel is a basic type implemented IChannel
type Channel struct {
	ID      string
	Clients map[string]*Client

	Subscribe   chan *Client
	Unsubscribe chan string
	Messages    chan *common.WebSocketMessage
}

func (c *Channel) GetID() string {
	return c.ID
}

func (c *Channel) AddSubscriber(client *Client) {
	c.Subscribe <- client
}

func (c *Channel) RemoveSubscriber(ID string) {
	c.Unsubscribe <- ID
}

func (c *Channel) AddMessage(msg *common.WebSocketMessage) {
	c.Messages <- msg
}

func (c *Channel) UnsubscribeChan() chan string {
	return c.Unsubscribe
}

func (c *Channel) SubScribeChan() chan *Client {
	return c.Subscribe
}

func (c *Channel) MessagesChan() chan *common.WebSocketMessage {
	return c.Messages
}

func (c *Channel) handleMessage(msg *common.WebSocketMessage) {
	for _, client := range c.Clients {
		err := client.Send(msg.Payload)

		if err != nil {
			utils.Debug("send message to client error: %v", err)
			c.handleUnsubscriber(client.ID)
		} else {
			utils.Debug("send message to client: channel: %s, payload: %s", msg.ChannelID, msg.Payload)
		}
	}
}

func (c *Channel) handleSubscriber(client *Client) {
	c.Clients[client.ID] = client

	utils.Debug("client(%s) joins channel(%s)", client.ID, c.ID)
}

func (c *Channel) handleUnsubscriber(ID string) {
	delete(c.Clients, ID)

	utils.Debug("client(%s) leaves channel(%s)", ID, c.ID)
}

func runChannel(c IChannel) {
	for {
		select {
		case msg := <-c.MessagesChan():
			c.handleMessage(msg)
		case client := <-c.SubScribeChan():
			c.handleSubscriber(client)
		case ID := <-c.UnsubscribeChan():
			c.handleUnsubscriber(ID)
		}
	}
}

var allChannels = make(map[string]IChannel, 10)
var allChannelsMutex = &sync.RWMutex{}

func findChannel(id string) IChannel {
	allChannelsMutex.RLock()
	defer allChannelsMutex.RUnlock()

	return allChannels[id]
}

func saveChannel(channel IChannel) {
	allChannelsMutex.Lock()
	defer allChannelsMutex.Unlock()

	allChannels[channel.GetID()] = channel
}

func createChannelByID(channelID string) IChannel {
	parts := strings.Split(channelID, "#")
	prefix := parts[0]

	var channel IChannel

	if creatorFunc := channelCreators[prefix]; creatorFunc != nil {
		channel = creatorFunc(channelID)
	} else {
		channel = createBaseChannel(channelID)
	}

	saveChannel(channel)
	go runChannel(channel)

	return channel
}

func createBaseChannel(channelID string) *Channel {
	return &Channel{
		ID:          channelID,
		Subscribe:   make(chan *Client),
		Unsubscribe: make(chan string),
		Messages:    make(chan *common.WebSocketMessage),
		Clients:     make(map[string]*Client),
	}
}
