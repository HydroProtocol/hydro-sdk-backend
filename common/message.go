package common

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// WebsocketMessage is message unit between engine and websocket
// Engine is producer
// Websocket is consumer

const WsTypeOrderChange = "orderChange"
const WsTypeTradeChange = "tradeChange"
const WsTypeLockedBalanceChange = "lockedBalanceChange"

const WsTypeNewMarketTrade = "newMarketTrade"

//const MessageTypeAccount = "account"
//const MessageTypeMarket = "market"

type WebSocketMessage struct {
	//MessageType string      `json:"message_type"`
	ChannelID string      `json:"channel_id"`
	Payload   interface{} `json:"payload"`
}

type WebsocketMarketOrderChangePayload struct {
	Side     string `json:"side"`
	Sequence uint64 `json:"sequence"`
	Price    string `json:"price"`
	Amount   string `json:"amount"`
}

type WebsocketLockedBalanceChangePayload struct {
	Type    string          `json:"type"`
	Symbol  string          `json:"symbol"`
	Balance decimal.Decimal `json:"balance"`
}

type WebsocketOrderChangePayload struct {
	Type  string      `json:"type"`
	Order interface{} `json:"order"`
}

type WebsocketTradeChangePayload struct {
	Type  string      `json:"type"`
	Trade interface{} `json:"trade"`
}

type WebsocketMarketNewMarketTradePayload struct {
	Type  string      `json:"type"`
	Trade interface{} `json:"trade"`
}

// engine event
const (
	EventNewOrder           = "EVENT/NEW_ORDER"
	EventCancelOrder        = "EVENT/EVENT_CANCEL_ORDER"
	EventRestartEngine      = "EVENT/EVENT_RESTART"
	EventConfirmTransaction = "EVENT/EVENT_CONFIRM_TRANSACTION"
	EventOpenMarket         = "EVENT/EVENT_OPEN_MARKET"
	EventCloseMarket        = "EVENT/EVENT_CLOSE_MARKET"
)

type Event struct {
	Type     string `json:"eventType"`
	MarketID string `json:"marketID"`
}

type NewOrderEvent struct {
	Event
	Order string `json:"order"`
}

type CancelOrderEvent struct {
	Event
	ID    string `json:"id"`
	Price string `json:"price"`
	Side  string `json:"side"`
}

type ConfirmTransactionEvent struct {
	Event
	Hash      string `json:"hash"`
	Status    string `json:"status"`
	Timestamp uint64 `json:"timestamp"`
}

// channel

const MarketChannelPrefix = "Market"
const AccountChannelPrefix = "TraderAddress"

func GetAccountChannelID(address string) string {
	return fmt.Sprintf("%s#%s", AccountChannelPrefix, address)
}

func GetMarketChannelID(marketID string) string {
	return fmt.Sprintf("%s#%s", MarketChannelPrefix, marketID)
}

func OrderBookChangeMessage(marketID string, sequence uint64, side string, price, amount decimal.Decimal) WebSocketMessage {
	payload := &WebsocketMarketOrderChangePayload{
		Sequence: sequence,
		Side:     side,
		Price:    price.String(),
		Amount:   amount.String(),
	}

	return marketChannelMessage(marketID, payload)
}

func marketChannelMessage(marketID string, payload interface{}) WebSocketMessage {
	return WebSocketMessage{
		//MessageType: MessageTypeMarket,
		ChannelID: GetMarketChannelID(marketID),
		Payload:   payload,
	}
}

func MessagesForUpdateOrder(order *MemoryOrder) []WebSocketMessage {
	updateMsg := orderUpdateMessage(order)

	var balanceChangeMsg WebSocketMessage
	if order.Side == "buy" {
		balanceChangeMsg = lockedBalanceChangeMessage(order.Trader, order.QuoteTokenSymbol())
	} else {
		balanceChangeMsg = lockedBalanceChangeMessage(order.Trader, order.BaseTokenSymbol())
	}

	return []WebSocketMessage{updateMsg, balanceChangeMsg}
}

func orderUpdateMessage(order *MemoryOrder) WebSocketMessage {
	return accountMessage(order.Trader, &WebsocketOrderChangePayload{
		Type:  WsTypeOrderChange,
		Order: order,
	})
}

func lockedBalanceChangeMessage(address, symbol string) WebSocketMessage {
	return accountMessage(address, &WebsocketLockedBalanceChangePayload{
		Type:   WsTypeLockedBalanceChange,
		Symbol: symbol,
	})
}

func accountMessage(address string, payload interface{}) WebSocketMessage {
	return WebSocketMessage{
		//MessageType: MessageTypeAccount,
		ChannelID: GetAccountChannelID(address),
		Payload:   payload,
	}
}
