package engine

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
)

// This queue is used to send message to ws servers

func orderUpdateMessage(order *common.MemoryOrder) common.WebSocketMessage {
	return accountMessage(order.Trader, &common.WebsocketOrderChangePayload{
		Type:  common.WsTypeOrderChange,
		Order: order,
	})
}

func accountMessage(address string, payload interface{}) common.WebSocketMessage {
	return common.WebSocketMessage{
		ChannelID: common.GetAccountChannelID(address),
		Payload:   payload,
	}
}

func MessagesForUpdateOrder(order *common.MemoryOrder) []common.WebSocketMessage {
	updateMsg := orderUpdateMessage(order)

	var balanceChangeMsg common.WebSocketMessage
	if order.Side == "buy" {
		balanceChangeMsg = lockedBalanceChangeMessage(order.Trader, order.QuoteTokenSymbol())
	} else {
		balanceChangeMsg = lockedBalanceChangeMessage(order.Trader, order.BaseTokenSymbol())
	}

	return []common.WebSocketMessage{updateMsg, balanceChangeMsg}
}

func lockedBalanceChangeMessage(address, symbol string) common.WebSocketMessage {
	return accountMessage(address, &common.WebsocketLockedBalanceChangePayload{
		Type:   common.WsTypeLockedBalanceChange,
		Symbol: symbol,
	})
}
