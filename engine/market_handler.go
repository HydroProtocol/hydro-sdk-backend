package engine

import (
	"context"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

type MarketHandler struct {
	ctx                  context.Context
	market               string
	marketAmountDecimals int
	orderbook            *common.Orderbook
}

func (m MarketHandler) handleNewOrder(newOrder *common.MemoryOrder) (matchResult common.MatchResult, hasMatchOrder bool) {

	if m.orderbook.CanMatch(newOrder) {
		matchResult = *m.orderbook.ExecuteMatch(newOrder, m.marketAmountDecimals)

		if len(matchResult.MatchItems) == 0 {
			log.Errorf("No Match Items, %+v %+v", matchResult, newOrder)
			panic(fmt.Errorf("no match items"))
		}

		for i := range matchResult.MatchItems {
			item := matchResult.MatchItems[i]

			msgs := common.MessagesForUpdateOrder(item.MakerOrder)
			matchResult.OrderBookActivities = append(matchResult.OrderBookActivities, msgs...)

			newOrder.Amount = newOrder.Amount.Sub(item.MatchedAmount)
			utils.Debugf("  [Take Liquidity] price: %s amount: %s (%s) ", item.MakerOrder.Price.StringFixed(5), item.MatchedAmount.StringFixed(5), item.MakerOrder.ID)
		}

		hasMatchOrder = true
	}

	msgs := common.MessagesForUpdateOrder(newOrder)
	matchResult.OrderBookActivities = append(matchResult.OrderBookActivities, msgs...)

	// check if newOrder can be added to orderbook
	if common.TakerOrderShouldBeRemoved(newOrder) {
		matchResult.TakerOrderIsDone = true
	} else {
		// if matched, gasFee is paid
		if matchResult.BaseTokenTotalMatchedAmtWithoutCanceledMatch().IsPositive() {
			newOrder.GasFeeAmount = decimal.Zero
		}

		e := m.orderbook.InsertOrder(newOrder)
		msg := common.OrderBookChangeMessage(m.market, m.orderbook.Sequence, e.Side, e.Price, e.Amount)
		matchResult.OrderBookActivities = append(matchResult.OrderBookActivities, msg)

		utils.Debugf("  [Make Liquidity] price: %s amount: %s (%s)", newOrder.Price.StringFixed(5), newOrder.Amount.StringFixed(5), newOrder.ID)
	}

	return
}

func (m *MarketHandler) handleCancelOrder(bookOrder *common.MemoryOrder) *common.OrderbookEvent {
	return m.orderbook.RemoveOrder(bookOrder)
}

func NewMarketHandler(ctx context.Context, market string) (*MarketHandler, error) {
	marketOrderbook := common.NewOrderbook(market)

	marketOrderbook.UsePlugin(func(e *common.OrderbookEvent) {
		marketOrderbook.Sequence = marketOrderbook.Sequence + 1
	})

	marketHandler := MarketHandler{
		market:    market,
		ctx:       ctx,
		orderbook: marketOrderbook,
	}

	return &marketHandler, nil
}
