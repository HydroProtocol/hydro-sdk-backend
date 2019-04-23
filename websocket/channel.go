package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"strings"
	"sync"
)

type IChannel interface {
	GetID() string

	// Thread safe calls
	AddSubscriber(*Client)
	RemoveSubscriber(string)
	AddMessage(message *common.WebSocketMessage)

	handleMessage(*common.WebSocketMessage)
	handleSubscriber(*Client)
	handleUnsubscriber(string)

	UnsubscribeChan() chan string
	SubScribeChan() chan *Client
	MessagesChan() chan *common.WebSocketMessage

	Stop()
}

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

func (c *Channel) Stop() {
	//utils.Debug("Channel(%s) is closed", c.GetID())
	//
	//DeleteChannel(c.ID)
	//close(c.Subscribe)
	//close(c.Unsubscribe)
	//close(c.Messages)
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

// Special Implements of IChannel
type AddressChannel struct {
	*Channel
}

func (c *AddressChannel) handleMessage(msg *common.WebSocketMessage) {
	for _, client := range c.Clients {
		err := client.Send(msg.Payload)

		if err != nil {
			utils.Debug("send message to client error: %v", err)
			c.handleUnsubscriber(client.ID)
		}
	}
}

type MarketChannel struct {
	*Channel
	MarketID  string
	Orderbook *Orderbook
}

func (c *MarketChannel) handleSubscriber(client *Client) {
	c.Channel.handleSubscriber(client)
	snapshot := c.Orderbook.SnapshotV2()

	msg := NewOrderbookLevel2Snapshot(c.MarketID, snapshot.Bids, snapshot.Asks)

	err := client.Send(msg)

	if err != nil {
		utils.Debug("send message to client error: %v", err)
		c.handleUnsubscriber(client.ID)
	}
}

func (c *MarketChannel) handleMessage(msg *common.WebSocketMessage) {
	var commonPayload struct {
		Type string
	}

	bts, _ := json.Marshal(msg.Payload)
	_ = json.Unmarshal(bts, &commonPayload)

	var messageToBeSent interface{}

	switch commonPayload.Type {
	case common.WsTypeNewMarketTrade:
		var p common.WebsocketMarketNewMarketTradePayload
		_ = json.Unmarshal(bts, &p)
		messageToBeSent = &p
	default:
		var p common.WebsocketMarketOrderChangePayload
		_ = json.Unmarshal(bts, &p)

		// if current message is already aggregated in orderbook, skip it
		if p.Sequence <= c.Orderbook.Sequence {
			return
		}

		res := c.Orderbook.onMessage(&p)

		messageToBeSent = NewOrderbookLevel2Update(c.MarketID, res.Side, res.Price.String(), res.Amount.String())
	}

	for _, client := range c.Clients {
		err := client.Send(messageToBeSent)

		if err != nil {
			utils.Debug("send message to client error: %v", err)
			c.handleUnsubscriber(client.ID)
		}
	}
}

func CreateAddressChannel(channelID string) *AddressChannel {
	channel := &AddressChannel{
		Channel: &Channel{
			ID:          channelID,
			Subscribe:   make(chan *Client),
			Unsubscribe: make(chan string),
			Messages:    make(chan *common.WebSocketMessage),
			Clients:     make(map[string]*Client),
		},
	}

	allChannelsMutex.Lock()
	defer allChannelsMutex.Unlock()
	allChannels[channel.ID] = channel

	go RunChannel(channel)

	return channel
}

func RunChannel(c IChannel) {

	utils.Debug("Channel(%s) is running", c.GetID())

	defer c.Stop()

	for {
		select {
		case msg := <-c.MessagesChan():
			c.handleMessage(msg)
		case client := <-c.SubScribeChan():
			c.handleSubscriber(client)
		case ID := <-c.UnsubscribeChan():
			c.handleUnsubscriber(ID)

			//if c.hasNoSubscriber() {
			//	return
			//}
		}
	}
}

var allChannels = make(map[string]IChannel, 10)
var allChannelsMutex = &sync.RWMutex{}

func FindChannel(id string) IChannel {
	allChannelsMutex.RLock()
	defer allChannelsMutex.RUnlock()

	return allChannels[id]
}

func CreateMarketChannel(channelID, marketID string) *MarketChannel {
	channel := &MarketChannel{
		MarketID: marketID,
		Channel: &Channel{
			ID:          channelID,
			Subscribe:   make(chan *Client),
			Unsubscribe: make(chan string),
			Messages:    make(chan *common.WebSocketMessage),
			Clients:     make(map[string]*Client),
		},
	}

	if marketID != "" {
		channel.Orderbook = InitOrderbook(marketID)
	}

	allChannelsMutex.Lock()
	defer allChannelsMutex.Unlock()
	allChannels[channel.ID] = channel

	go RunChannel(channel)
	return channel
}

func DeleteChannel(id string) {
	allChannelsMutex.Lock()
	defer allChannelsMutex.Unlock()

	delete(allChannels, id)
}

func CreateChannelByID(ChannelID string) IChannel {
	channelType, ID := parseChannelID(ChannelID)

	if channelType == common.MarketChannelPrefix {
		return CreateMarketChannel(ChannelID, ID)
	} else if channelType == common.AccountChannelPrefix {
		return CreateAddressChannel(ChannelID)
	} else {
		return nil
	}
}

func parseChannelID(channelID string) (string, string) {
	if strings.HasPrefix(channelID, common.MarketChannelPrefix) {
		return common.MarketChannelPrefix, strings.Replace(channelID, fmt.Sprintf("%s#", common.MarketChannelPrefix), "", -1)
	} else if strings.HasPrefix(channelID, common.AccountChannelPrefix) {
		return common.AccountChannelPrefix, ""
	} else {
		return "", ""
	}
}
