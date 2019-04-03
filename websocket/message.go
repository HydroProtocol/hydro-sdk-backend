package websocket

type OrderbookLevel2Snapshot struct {
	Type     string      `json:"type"`
	MarketID string      `json:"marketID"`
	Bids     [][2]string `json:"bids"`
	Asks     [][2]string `json:"asks"`
}

func NewOrderbookLevel2Snapshot(marketID string, bids, asks [][2]string) *OrderbookLevel2Snapshot {
	return &OrderbookLevel2Snapshot{
		Bids:     bids,
		Asks:     asks,
		MarketID: marketID,
		Type:     "level2OrderbookSnapshot",
	}
}

type OrderbookLevel2Update struct {
	Type     string `json:"type"`
	MarketID string `json:"marketID"`
	Price    string `json:"price"`
	Side     string `json:"side"`
	Amount   string `json:"amount"`
}

func NewOrderbookLevel2Update(marketID string, side, price, amount string) *OrderbookLevel2Update {
	return &OrderbookLevel2Update{
		Type:     "level2OrderbookUpdate",
		MarketID: marketID,
		Side:     side,
		Price:    price,
		Amount:   amount,
	}
}
