package websocket

type orderbookLevel2Snapshot struct {
	Type     string      `json:"type"`
	MarketID string      `json:"marketID"`
	Bids     [][2]string `json:"bids"`
	Asks     [][2]string `json:"asks"`
}

func newOrderbookLevel2Snapshot(marketID string, bids, asks [][2]string) *orderbookLevel2Snapshot {
	return &orderbookLevel2Snapshot{
		Bids:     bids,
		Asks:     asks,
		MarketID: marketID,
		Type:     "level2OrderbookSnapshot",
	}
}

type orderbookLevel2Update struct {
	Type     string `json:"type"`
	MarketID string `json:"marketID"`
	Price    string `json:"price"`
	Side     string `json:"side"`
	Amount   string `json:"amount"`
}

func newOrderbookLevel2Update(marketID string, side, price, amount string) *orderbookLevel2Update {
	return &orderbookLevel2Update{
		Type:     "level2OrderbookUpdate",
		MarketID: marketID,
		Side:     side,
		Price:    price,
		Amount:   amount,
	}
}
