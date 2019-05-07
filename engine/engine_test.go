package engine

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type engineTestSuite struct {
	suite.Suite
}

func TestEngineTestSuite(t *testing.T) {
	suite.Run(t, new(engineTestSuite))
}

func (s *engineTestSuite) TestNewEngine() {
	e := NewEngine(context.Background())

	order := common.MemoryOrder{
		ID:       "fake-id",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(100.0),
		Side:     "sell",
		Type:     "limit",
	}

	matchRst, hasMatch := e.HandleNewOrder(&order)

	s.False(hasMatch, "should have no match")
	s.True(len(matchRst.MatchItems) == 0, "should have no match")
}

func (s *engineTestSuite) TestNewEngineHandleOrders() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:       "fake-id1",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(100.0),
		Side:     "sell",
		Type:     "limit",
	}
	orderBuy := common.MemoryOrder{
		ID:       "fake-id2",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(100.0),
		Side:     "buy",
		Type:     "limit",
	}

	matchRst, hasMatch := e.HandleNewOrder(&orderSell)
	matchRst2, hasMatch2 := e.HandleNewOrder(&orderBuy)

	s.False(hasMatch, "should have no match")
	s.Equal(0, len(matchRst.MatchItems), "should have no match")

	s.True(hasMatch2, "should have match")
	s.True(len(matchRst2.MatchItems) > 0, "should have match")

	s.True(matchRst2.TakerOrderLeftAmount.IsZero())
	s.Equal(1, len(matchRst2.MatchItems))

	matchItem := matchRst2.MatchItems[0]
	s.Equal("fake-id1", matchItem.MakerOrder.ID)
	s.True(matchItem.MatchedAmount.Equal(decimal.NewFromFloat(100)))
}

func (s *engineTestSuite) TestNewEngineHandleOrders2() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:       "fake-id1",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(101.0),
		Side:     "sell",
		Type:     "limit",
	}
	orderBuy := common.MemoryOrder{
		ID:       "fake-id2",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(100.0),
		Side:     "buy",
		Type:     "limit",
	}

	matchRst, hasMatch := e.HandleNewOrder(&orderSell)
	matchRst2, hasMatch2 := e.HandleNewOrder(&orderBuy)

	s.False(hasMatch, "should have no match")
	s.Equal(0, len(matchRst.MatchItems), "should have no match")

	s.True(hasMatch2, "should have match")
	s.True(len(matchRst2.MatchItems) > 0, "should have match")

	s.True(matchRst2.TakerOrderLeftAmount.IsZero())
	s.Equal(1, len(matchRst2.MatchItems))

	matchItem := matchRst2.MatchItems[0]
	s.Equal("fake-id1", matchItem.MakerOrder.ID)
	s.True(matchItem.MatchedAmount.Equal(decimal.NewFromFloat(100)))

	handler, _ := e.marketHandlerMap["HOT-WETH"]

	sellOrder, _ := handler.orderbook.GetOrder("fake-id1", "sell", decimal.NewFromFloat(1))
	s.True(sellOrder.Amount.Equal(decimal.NewFromFloat(1)))
	s.True(sellOrder.GasFeeAmount.Equal(decimal.Zero))

	s.Nil(handler.orderbook.MaxBid())
}

func (s *engineTestSuite) TestHandleOrdersAfterMatchSellOrderCanBeSmall() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:           "fake-id1",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.01),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
	}
	orderBuy := common.MemoryOrder{
		ID:           "fake-id2",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.0),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
	}

	e.HandleNewOrder(&orderSell)
	e.HandleNewOrder(&orderBuy)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MinAsk())

	sellOrder, _ := handler.orderbook.GetOrder("fake-id1", "sell", decimal.NewFromFloat(1))
	s.True(sellOrder.Amount.Equal(decimal.NewFromFloat(0.01)))
	s.True(sellOrder.GasFeeAmount.Equal(decimal.Zero))
}

// a small buy(quoteAmt < taker gas + takerTradeFee) will not be accept
func (s *engineTestSuite) TestSmallBuyBeIgnored() {
	e := NewEngine(context.Background())

	// quoteAmt = 0.1
	// gas + tradeFee = 0.1 + 0.1*0.003 (assume taker's gas & fee same as this order)
	smallBuy := common.MemoryOrder{
		ID:           "fake-id",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(0.1),
		Amount:       decimal.NewFromFloat(1.0),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	e.HandleNewOrder(&smallBuy)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.Nil(handler.orderbook.MaxBid())
}
func (s *engineTestSuite) TestBigBuyWillNotBeIgnored() {
	e := NewEngine(context.Background())

	// quoteAmt = 0.1
	// gas + tradeFee = 0.0997 + 0.1*0.003 (assume taker's gas & fee same as this order)
	smallBuy := common.MemoryOrder{
		ID:           "fake-id",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(0.1),
		Amount:       decimal.NewFromFloat(1.0),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.0997),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	e.HandleNewOrder(&smallBuy)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MaxBid())
}

// a small sell (quoteAmt < its gas + its tradeFee) will not be accept
func (s *engineTestSuite) TestSmallSellBeIgnored() {
	e := NewEngine(context.Background())

	// quoteAmt = 0.1
	// gas + tradeFee = 0.1 + 0.1*0.003 (assume taker's gas & fee same as this order)
	smallSell := common.MemoryOrder{
		ID:           "fake-id",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(0.1),
		Amount:       decimal.NewFromFloat(1.0),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	_, hasMatch := e.HandleNewOrder(&smallSell)
	s.False(hasMatch)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.Nil(handler.orderbook.MaxBid())
}

func (s *engineTestSuite) TestBigSellWillNotBeIgnored() {
	e := NewEngine(context.Background())

	// quoteAmt = 0.1
	// gas + tradeFee = 0.0997 + 0.1*0.003 (assume taker's gas & fee same as this order)
	bigSell := common.MemoryOrder{
		ID:           "fake-id",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(0.1),
		Amount:       decimal.NewFromFloat(1.0),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.0997),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	_, hasMatch := e.HandleNewOrder(&bigSell)
	s.False(hasMatch)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MinAsk())
}

// after match, maker Sell will stay on orderbook however small it is
func (s *engineTestSuite) TestSmallSellCanStayOnBookAfterMatch() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:           "fake-id1",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.01),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}
	orderBuy := common.MemoryOrder{
		ID:           "fake-id2",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.0),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	e.HandleNewOrder(&orderSell)
	e.HandleNewOrder(&orderBuy)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MinAsk())

	sellOrder, _ := handler.orderbook.GetOrder("fake-id1", "sell", decimal.NewFromFloat(1))
	s.True(sellOrder.Amount.Equal(decimal.NewFromFloat(0.01)))
	s.True(sellOrder.GasFeeAmount.Equal(decimal.Zero))
}

// after match, maker Buy won't stay on orderbook if its small (quoteAmt < takerGas + takerTradeFee)
func (s *engineTestSuite) TestSmallBuyCanNotStayOnBookAfterMatch() {
	e := NewEngine(context.Background())

	orderBuy := common.MemoryOrder{
		ID:           "fake-id2",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.1),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}
	orderSell := common.MemoryOrder{
		ID:           "fake-id1",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	e.HandleNewOrder(&orderBuy)
	e.HandleNewOrder(&orderSell)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.Nil(handler.orderbook.MaxBid())
	s.Nil(handler.orderbook.MinAsk())
}
func (s *engineTestSuite) TestBigBuyCanStayOnBookAfterMatch() {
	e := NewEngine(context.Background())

	orderBuy := common.MemoryOrder{
		ID:           "fake-id2",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.1),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}
	orderSell := common.MemoryOrder{
		ID:           "fake-id1",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.0997),
		MakerFeeRate: decimal.NewFromFloat(0.001),
		TakerFeeRate: decimal.NewFromFloat(0.003),
	}

	// after match
	// quote: 0.1*1
	// taker gas + tradeFee = 0.0997 + 0.1*0.003

	e.HandleNewOrder(&orderBuy)
	e.HandleNewOrder(&orderSell)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MaxBid())
	s.Nil(handler.orderbook.MinAsk())

	buyOrder, _ := handler.orderbook.GetOrder("fake-id2", "buy", decimal.NewFromFloat(1))
	s.True(buyOrder.Amount.Equal(decimal.NewFromFloat(0.1)))
	s.True(buyOrder.GasFeeAmount.Equal(decimal.Zero))
}

func (s *engineTestSuite) TestHandleOrdersKeepBigRemainingOrder() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:           "fake-id1",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.1),
		Side:         "sell",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
	}
	orderBuy := common.MemoryOrder{
		ID:           "fake-id2",
		MarketID:     "HOT-WETH",
		Price:        decimal.NewFromFloat(1.0),
		Amount:       decimal.NewFromFloat(100.0),
		Side:         "buy",
		Type:         "limit",
		GasFeeAmount: decimal.NewFromFloat(0.1),
	}

	e.HandleNewOrder(&orderSell)
	e.HandleNewOrder(&orderBuy)

	handler, _ := e.marketHandlerMap["HOT-WETH"]
	s.NotNil(handler.orderbook.MinAsk())
}

type FakeDBHandler struct {
}

func (handler FakeDBHandler) Update(matchRst common.MatchResult) sync.WaitGroup {
	log.Info("Update called of fake db handler")
	return sync.WaitGroup{}
}

func (s *engineTestSuite) TestNewEngineWithDBHandler() {
	h := FakeDBHandler{}

	e := NewEngine(context.Background())
	e.RegisterDBHandler(h)

	order := common.MemoryOrder{
		ID:       "fake-id",
		MarketID: "HOT-WETH",
		Price:    decimal.NewFromFloat(1.0),
		Amount:   decimal.NewFromFloat(100.0),
		Side:     "sell",
		Type:     "limit",
	}

	matchRst, hasMatch := e.HandleNewOrder(&order)

	s.False(hasMatch, "should have no match")
	s.Equal(0, len(matchRst.MatchItems), "should have no match")
}
