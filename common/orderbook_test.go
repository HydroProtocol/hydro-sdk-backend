package common

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"testing"
)

type orderbookTestSuite struct {
	suite.Suite
	book *Orderbook
}

func (s *orderbookTestSuite) SetupSuite() {
}

func (s *orderbookTestSuite) SetupTest() {
	s.book = NewOrderbook("test")
}

func (s *orderbookTestSuite) TearDownTest() {
}

func (s *orderbookTestSuite) TearDownSuite() {
}

func (s *orderbookTestSuite) TestSnapshot() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.3", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o3", "sell", "1.4", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o4", "sell", "1.5", "3.4"))

	s.Equal(&SnapshotV2{
		Bids: [][2]string{{"1.3", "3.4"}, {"1.2", "3.4"}},
		Asks: [][2]string{{"1.4", "3.4"}, {"1.5", "3.4"}},
	}, s.book.SnapshotV2())
}

func (s *orderbookTestSuite) TestNewOrderbok() {
	s.Equal(0, s.book.bidsTree.Len())
	s.Equal(0, s.book.asksTree.Len())
	s.Nil(s.book.MaxBid())
	s.Nil(s.book.MinAsk())
}

func (s *orderbookTestSuite) TestInsertAndRemoveOrder() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "sell", "1.8", "4"))

	s.Equal(1, s.book.bidsTree.Len())
	s.Equal(1, s.book.asksTree.Len())

	maxBidPriceLevel := s.book.bidsTree.Max().(*priceLevel)
	s.Equal(2, maxBidPriceLevel.Len())
	s.Equal("3", maxBidPriceLevel.totalAmount.String())

	s.book.RemoveOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.Equal(1, maxBidPriceLevel.Len())
	s.Equal("2", maxBidPriceLevel.totalAmount.String())
}

func (s *orderbookTestSuite) TestInsertAndChangeOrder() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "sell", "1.8", "4"))

	s.Equal(1, s.book.bidsTree.Len())
	s.Equal(1, s.book.asksTree.Len())

	maxBidPriceLevel := s.book.bidsTree.Max().(*priceLevel)
	s.Equal(2, maxBidPriceLevel.Len())
	s.Equal("3", maxBidPriceLevel.totalAmount.String())

	s.book.ChangeOrder(NewLimitOrder("o1", "buy", "1.2", "1"), decimal.NewFromFloat(0.9))
	s.Equal(2, maxBidPriceLevel.Len())
	s.Equal("3.9", maxBidPriceLevel.totalAmount.String())
}

var amtDecimals = 3

func (s *orderbookTestSuite) TestMatch() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "buy", "1.3", "2"))
	s.book.InsertOrder(NewLimitOrder("o4", "buy", "1.5", "2"))

	// no match
	result := s.book.MatchOrder(NewLimitOrder("o5", "sell", "2", "2"), amtDecimals)
	//s.True(result.NoMatch)
	//s.False(result.FullMatch)
	s.Equal("0", result.QuoteTokenTotalMatchedAmt().String())
	s.Equal(0, len(result.MatchItems))
}

func (s *orderbookTestSuite) TestMatch2() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "buy", "1.3", "2"))
	s.book.InsertOrder(NewLimitOrder("o4", "buy", "1.5", "2"))

	// full match
	result := s.book.MatchOrder(NewLimitOrder("o5", "sell", "1.5", "2"), amtDecimals)
	//s.False(result.NoMatch)
	//s.True(result.FullMatch)
	s.Equal("2", result.BaseTokenTotalMatchedAmtWithoutCanceledMatch().String())
	s.Equal(1, len(result.MatchItems))
	s.Equal("2", result.MatchItems[0].MatchedAmount.String())
	s.Equal("o4", result.MatchItems[0].MakerOrder.ID)
}

func (s *orderbookTestSuite) TestMatch3() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "buy", "1.3", "2"))
	s.book.InsertOrder(NewLimitOrder("o4", "buy", "1.5", "2"))

	// partial match
	result := s.book.MatchOrder(NewLimitOrder("o5", "sell", "1.5", "3"), amtDecimals)
	//s.False(result.NoMatch)
	//s.False(result.FullMatch)
	s.Equal("2", result.BaseTokenTotalMatchedAmtWithoutCanceledMatch().String())
	s.Equal(1, len(result.MatchItems))
	s.Equal("2", result.MatchItems[0].MatchedAmount.String())
	s.Equal("o4", result.MatchItems[0].MakerOrder.ID)
}

func (s *orderbookTestSuite) TestMatch4() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "buy", "1.3", "2"))
	s.book.InsertOrder(NewLimitOrder("o4", "buy", "1.5", "2"))

	// multi price full match
	result := s.book.MatchOrder(NewLimitOrder("o5", "sell", "1.2", "5"), amtDecimals)
	//s.False(result.NoMatch)
	//s.True(result.FullMatch)
	s.Equal("5", result.BaseTokenTotalMatchedAmtWithoutCanceledMatch().String())
	s.Equal(3, len(result.MatchItems))
	s.Equal("2", result.MatchItems[0].MatchedAmount.String())
	s.Equal("o4", result.MatchItems[0].MakerOrder.ID)
	s.Equal("2", result.MatchItems[1].MatchedAmount.String())
	s.Equal("o3", result.MatchItems[1].MakerOrder.ID)
	s.Equal("1", result.MatchItems[2].MatchedAmount.String())
	s.Equal("o1", result.MatchItems[2].MakerOrder.ID)

}

func (s *orderbookTestSuite) TestMatch5() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "1"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.2", "2"))
	s.book.InsertOrder(NewLimitOrder("o3", "buy", "1.3", "2"))
	s.book.InsertOrder(NewLimitOrder("o4", "buy", "1.5", "2"))

	// multi price partial match
	result := s.book.MatchOrder(NewLimitOrder("o5", "sell", "1.2", "8"), amtDecimals)
	//s.False(result.NoMatch)
	//s.False(result.FullMatch)
	s.Equal("7", result.BaseTokenTotalMatchedAmtWithoutCanceledMatch().String())
	s.Equal(4, len(result.MatchItems))
	s.Equal("2", result.MatchItems[0].MatchedAmount.String())
	s.Equal("o4", result.MatchItems[0].MakerOrder.ID)
	s.Equal("2", result.MatchItems[1].MatchedAmount.String())
	s.Equal("o3", result.MatchItems[1].MakerOrder.ID)
	s.Equal("1", result.MatchItems[2].MatchedAmount.String())
	s.Equal("o1", result.MatchItems[2].MakerOrder.ID)
	s.Equal("2", result.MatchItems[3].MatchedAmount.String())
	s.Equal("o2", result.MatchItems[3].MakerOrder.ID)
}

func (s *orderbookTestSuite) TestCanBeMatched() {
	s.book.InsertOrder(NewLimitOrder("o1", "buy", "1.2", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o2", "buy", "1.3", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o3", "sell", "1.4", "3.4"))
	s.book.InsertOrder(NewLimitOrder("o4", "sell", "1.5", "3.4"))

	canBeMatched0 := s.book.CanMatch(NewLimitOrder("1", "buy", "1.1", "1"))
	canBeMatched1 := s.book.CanMatch(NewLimitOrder("1", "sell", "1.3", "1"))
	canBeMatched2 := s.book.CanMatch(NewLimitOrder("1", "sell", "1.4", "1"))
	canBeMatched3 := s.book.CanMatch(NewLimitOrder("1", "buy", "1.6", "1"))
	canBeMatched4 := s.book.CanMatch(NewLimitOrder("1", "buy", "1.1", "1"))

	s.Equal(canBeMatched0, false)
	s.Equal(canBeMatched1, true)
	s.Equal(canBeMatched2, false)
	s.Equal(canBeMatched3, true)
	s.Equal(canBeMatched4, false)
}

func TestOrderbookTestSuite(t *testing.T) {
	suite.Run(t, new(orderbookTestSuite))
}

type orderTestSuite struct {
	suite.Suite
}

func (s *orderTestSuite) SetupSuite() {
}

func (s *orderTestSuite) TearDownSuite() {
}

func (s *orderTestSuite) TearDownTest() {
}

func (s *orderTestSuite) TestNewOrder() {
	s.Panics(func() { NewLimitOrder("", "buy", "1", "2") })     // no id
	s.Panics(func() { NewLimitOrder("123", "hehe", "1", "2") }) // wrong type
	s.Panics(func() { NewLimitOrder("123", "buy", "a", "2") })  // wrong price
	s.Panics(func() { NewLimitOrder("123", "buy", "1", "b") })  // wrong Amount
	s.NotPanics(func() { NewLimitOrder("123", "buy", "1", "2") })
	s.NotPanics(func() { NewLimitOrder("123", "sell", "1", "2") })
	s.NotPanics(func() { NewLimitOrder("123", "sell", "1.121423", "0.1241242") })
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(orderTestSuite))
}

// NewLimitOrder ...
func NewLimitOrder(id string, side string, price string, amount string) *MemoryOrder {
	return NewOrder(id, side, price, amount, "limit")
}

func NewOrder(id, side, price, amount, _type string) *MemoryOrder {
	if len(id) <= 0 {
		panic(fmt.Errorf("ID can't be blank"))
	}

	amountDecimal, err := decimal.NewFromString(amount)

	if side != "buy" && side != "sell" {
		panic(fmt.Errorf("side should be buy/sell. passed: %s", side))
	}

	if err != nil {
		panic(fmt.Errorf("amount decimal error, Amount: %s, error: %+v", amount, err))
	}

	priceDecimal, err := decimal.NewFromString(price)
	if err != nil {
		panic(fmt.Errorf("price decimal error, Price: %s, error: %+v", price, err))
	}

	return &MemoryOrder{
		ID:    id,
		Side:  side,
		Price: priceDecimal,

		Amount: amountDecimal,
		Type:   _type,
	}
}
