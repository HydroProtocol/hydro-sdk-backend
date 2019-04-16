package engine

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
)

type Orderbook struct {
	*common.Orderbook
	Sequence uint64
}

func NewOrderbook(marketID string) *Orderbook {
	originalOrderbook := common.NewOrderbook(marketID)

	orderbook := &Orderbook{
		Orderbook: originalOrderbook,
		Sequence:  uint64(0),
	}

	return orderbook
}
