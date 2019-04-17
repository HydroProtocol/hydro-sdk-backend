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
		ID:     "fake-id",
		Market: "HOT-WETH",
		Price:  decimal.NewFromFloat(1.0),
		Amount: decimal.NewFromFloat(100.0),
		Side:   "sell",
		Type:   "limit",
	}

	matchRst, hasMatch := e.handleNewOrder(&order)

	s.False(hasMatch, "should have no match")
	s.True(len(matchRst.MatchItems) == 0, "should have no match")
}

func (s *engineTestSuite) TestNewEngineHandleOrders() {
	e := NewEngine(context.Background())

	orderSell := common.MemoryOrder{
		ID:     "fake-id1",
		Market: "HOT-WETH",
		Price:  decimal.NewFromFloat(1.0),
		Amount: decimal.NewFromFloat(100.0),
		Side:   "sell",
		Type:   "limit",
	}
	orderBuy := common.MemoryOrder{
		ID:     "fake-id2",
		Market: "HOT-WETH",
		Price:  decimal.NewFromFloat(1.0),
		Amount: decimal.NewFromFloat(100.0),
		Side:   "buy",
		Type:   "limit",
	}

	matchRst, hasMatch := e.handleNewOrder(&orderSell)
	matchRst2, hasMatch2 := e.handleNewOrder(&orderBuy)

	s.False(hasMatch, "should have no match")
	s.Equal(0, len(matchRst.MatchItems), "should have no match")

	s.True(hasMatch2, "should have match")
	s.True(len(matchRst2.MatchItems) > 0, "should have match")

	//todo
	//s.True(matchRst2.IsFullMatch)

	s.True(matchRst2.TakerOrderLeftAmount.IsZero())
	s.Equal(1, len(matchRst2.MatchItems))

	matchItem := matchRst2.MatchItems[0]
	s.Equal("fake-id1", matchItem.MakerOrder.ID)
	s.True(matchItem.MatchedAmount.Equal(decimal.NewFromFloat(100)))
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
		ID:     "fake-id",
		Market: "HOT-WETH",
		Price:  decimal.NewFromFloat(1.0),
		Amount: decimal.NewFromFloat(100.0),
		Side:   "sell",
		Type:   "limit",
	}

	matchRst, hasMatch := e.handleNewOrder(&order)

	s.False(hasMatch, "should have no match")
	s.Equal(0, len(matchRst.MatchItems), "should have no match")
}
