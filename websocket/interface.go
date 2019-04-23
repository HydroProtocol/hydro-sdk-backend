package websocket

import "github.com/HydroProtocol/hydro-sdk-backend/common"

type IChannel interface {
	GetID() string

	// Thread safe calls
	AddSubscriber(*Client)
	RemoveSubscriber(string)
	AddMessage(message *common.WebSocketMessage)

	UnsubscribeChan() chan string
	SubScribeChan() chan *Client
	MessagesChan() chan *common.WebSocketMessage

	handleMessage(*common.WebSocketMessage)
	handleSubscriber(*Client)
	handleUnsubscriber(string)
}
