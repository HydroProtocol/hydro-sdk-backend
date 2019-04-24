package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"strings"
)

type marketChannel struct {
	*Channel
	MarketID  string
	Orderbook *Orderbook
}

func (c *marketChannel) handleSubscriber(client *Client) {
	c.Channel.handleSubscriber(client)
	snapshot := c.Orderbook.SnapshotV2()

	msg := newOrderbookLevel2Snapshot(c.MarketID, snapshot.Bids, snapshot.Asks)

	err := client.Send(msg)

	if err != nil {
		utils.Debug("send message to client error: %v", err)
		c.handleUnsubscriber(client.ID)
	}
}

func (c *marketChannel) handleMessage(msg *common.WebSocketMessage) {
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

		messageToBeSent = newOrderbookLevel2Update(c.MarketID, res.Side, res.Price.String(), res.Amount.String())
	}

	for _, client := range c.Clients {
		err := client.Send(messageToBeSent)

		if err != nil {
			utils.Debug("send message to client error: %v", err)
			c.handleUnsubscriber(client.ID)
		}
	}
}

func NewMarketChannelCreator(fetcher SnapshotFetcher) func(channelID string) IChannel {
	return func(channelID string) IChannel {
		marketID := strings.Replace(channelID, fmt.Sprintf("%s#", common.MarketChannelPrefix), "", -1)

		channel := &marketChannel{
			MarketID: marketID,
			Channel:  createBaseChannel(channelID),
		}

		snapshot := fetcher.GetV2(marketID)
		channel.Orderbook = initOrderbook(marketID, snapshot)

		return channel
	}
}
