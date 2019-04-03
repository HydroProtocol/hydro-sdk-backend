package websocket

import (
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
)

type Orderbook struct {
	Sequence uint64
	*common.Orderbook
}

type OnMessageResult struct {
	Price  decimal.Decimal
	Side   string
	Amount decimal.Decimal
}

func InitOrderbook(marketID string) *Orderbook {

	snapshot := GetMarketOrderbookSnapshotV2(nil, marketID)

	orderbook := &Orderbook{
		Orderbook: common.NewOrderbook(marketID),
		Sequence:  snapshot.Sequence,
	}

	for _, v := range snapshot.Bids {
		price, _ := decimal.NewFromString(v[0])
		amount, _ := decimal.NewFromString(v[1])
		order := &common.MemoryOrder{
			Side:   "buy",
			ID:     fmt.Sprintf("buy-%s", v[0]),
			Amount: amount,
			Price:  price,
		}
		orderbook.Orderbook.InsertOrder(order)
	}

	for _, v := range snapshot.Asks {
		price, _ := decimal.NewFromString(v[0])
		amount, _ := decimal.NewFromString(v[1])
		order := &common.MemoryOrder{
			Side:   "sell",
			ID:     fmt.Sprintf("sell-%s", v[0]),
			Amount: amount,
			Price:  price,
		}
		orderbook.Orderbook.InsertOrder(order)
	}

	return orderbook
}

func (o *Orderbook) onMessage(payload *common.WebsocketMarketOrderChangePayload) *OnMessageResult {

	res := &OnMessageResult{
		Side:  payload.Side,
		Price: utils.StringToDecimal(payload.Price),
	}

	orderID := fmt.Sprintf("%s-%s", payload.Side, payload.Price)

	if order, ok := o.Orderbook.GetOrder(orderID, payload.Side, utils.StringToDecimal(payload.Price)); ok {
		changedAmount := utils.StringToDecimal(payload.Amount)
		order.Amount = order.Amount.Add(changedAmount)
		priceLevelAmountAfterChange := order.Amount
		res.Amount = priceLevelAmountAfterChange

		if priceLevelAmountAfterChange.LessThanOrEqual(decimal.Zero) {
			o.Orderbook.RemoveOrder(order)
		} else {
			o.Orderbook.ChangeOrder(order, changedAmount)
		}
	} else {
		//s := o.SnapshotV2()
		//s.Sequence = o.Sequence

		o.Orderbook.InsertOrder(&common.MemoryOrder{
			ID:     orderID,
			Price:  utils.StringToDecimal(payload.Price),
			Amount: utils.StringToDecimal(payload.Amount),
			Side:   payload.Side,
		})

		if utils.StringToDecimal(payload.Amount).LessThan(decimal.Zero) {
			panic(fmt.Errorf("Can't find order in orderbook, change payload is %v", payload))
		}

		res.Amount = utils.StringToDecimal(payload.Amount)
	}

	o.Sequence = payload.Sequence
	return res
}
