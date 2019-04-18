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

type WebSocketMessage struct {
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
