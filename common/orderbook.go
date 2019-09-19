package common

import (
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/cevaris/ordered_map"
	"github.com/labstack/gommon/log"
	"github.com/petar/GoLLRB/llrb"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
)

type OrderbookEvent struct {
	Side    string
	OrderID string
	Price   decimal.Decimal
	Amount  decimal.Decimal
}

type OrderbookPlugin func(event *OrderbookEvent)

type IOrderBook interface {
	InsertOrder(*MemoryOrder)
	RemoveOrder(*MemoryOrder)
	ChangeOrder(*MemoryOrder, decimal.Decimal)

	UsePlugin(plugin OrderbookPlugin)

	SnapshotV2() *SnapshotV2
	CanMatch(*MemoryOrder) bool
	MatchOrder(*MemoryOrder, int) *MatchResult
	ExecuteMatch(*MemoryOrder, int) *MatchResult
}

type (
	MatchResult struct {
		TakerOrder           *MemoryOrder
		TakerOrderIsDone     bool
		MatchItems           []*MatchItem
		TakerOrderLeftAmount decimal.Decimal
		OrderBookActivities  []WebSocketMessage
	}

	MatchItem struct {
		MakerOrder            *MemoryOrder
		MakerOrderIsDone      bool
		MatchedAmount         decimal.Decimal
		MatchShouldBeCanceled bool
	}

	MemoryOrder struct {
		ID           string          `json:"id"`
		MarketID     string          `json:"marketID"`
		Price        decimal.Decimal `json:"price"`
		Amount       decimal.Decimal `json:"amount"`
		Side         string          `json:"side"`
		Type         string          `json:"type"`
		Trader       string          `json:"trader"`
		GasFeeAmount decimal.Decimal `json:"gasFeeAmount"`
		MakerFeeRate decimal.Decimal `json:"makerFeeRate"`
		TakerFeeRate decimal.Decimal `json:"takerFeeRate"`
	}

	SnapshotV2 struct {
		Sequence uint64      `json:"sequence"`
		Bids     [][2]string `json:"bids"`
		Asks     [][2]string `json:"asks"`
	}
)

func (order *MemoryOrder) QuoteTokenSymbol() string {
	parts := strings.Split(order.MarketID, "-")
	if len(parts) == 2 {
		return parts[1]
	} else {
		return "unknown"
	}
}

func (order *MemoryOrder) BaseTokenSymbol() string {
	parts := strings.Split(order.MarketID, "-")
	if len(parts) == 2 {
		return parts[0]
	} else {
		return "unknown"
	}
}

func (matchResult *MatchResult) QuoteTokenTotalMatchedAmt() decimal.Decimal {
	quoteTokenAmt := decimal.Zero
	for _, item := range matchResult.MatchItems {
		quoteTokenAmt = quoteTokenAmt.Add(item.MatchedAmount.Mul(item.MakerOrder.Price))
	}

	return quoteTokenAmt
}

func (matchResult *MatchResult) TakerTradeFeeInQuoteToken() decimal.Decimal {
	return matchResult.QuoteTokenTotalMatchedAmt().Mul(matchResult.TakerOrder.TakerFeeRate)
}

func (matchResult *MatchResult) MakerTradeFeeInQuoteToken() (sum decimal.Decimal) {
	for _, item := range matchResult.MatchItems {
		sum = sum.Add(item.MatchedAmount.Mul(item.MakerOrder.Price).Mul(item.MakerOrder.MakerFeeRate))
	}

	return
}

func (matchResult *MatchResult) BaseTokenTotalMatchedAmtWithoutCanceledMatch() decimal.Decimal {
	baseTokenAmt := decimal.Zero
	for _, item := range matchResult.MatchItems {
		if !item.MatchShouldBeCanceled {
			baseTokenAmt = baseTokenAmt.Add(item.MatchedAmount)
		}
	}

	return baseTokenAmt
}

func (matchResult *MatchResult) SumOfGasOfMakerOrders() decimal.Decimal {
	sum := decimal.Zero
	for _, item := range matchResult.MatchItems {
		sum = sum.Add(item.MakerOrder.GasFeeAmount)
	}

	return sum
}

func (matchResult MatchResult) ExistMatchToBeExecuted() bool {
	for _, match := range matchResult.MatchItems {
		if !match.MatchShouldBeCanceled {
			return true
		}
	}

	return false
}

type priceLevel struct {
	price       decimal.Decimal
	totalAmount decimal.Decimal
	orderMap    *ordered_map.OrderedMap
}

func newPriceLevel(price decimal.Decimal) *priceLevel {
	return &priceLevel{
		price:       price,
		totalAmount: decimal.Zero,
		orderMap:    ordered_map.NewOrderedMap(),
	}
}

func (p *priceLevel) Len() int {
	return p.orderMap.Len()
}

func (p *priceLevel) InsertOrder(order *MemoryOrder) {
	log.Debug("InsertOrder:", order.ID)

	if _, ok := p.orderMap.Get(order.ID); ok {
		panic(fmt.Errorf("can't add order which is already in this priceLevel. priceLevel: %s, orderID: %s", p.price.String(), order.ID))
	}

	p.orderMap.Set(order.ID, order)
	p.totalAmount = p.totalAmount.Add(order.Amount)
}

func (p *priceLevel) RemoveOrder(o *MemoryOrder) {
	orderItem, ok := p.orderMap.Get(o.ID)

	if !ok {
		panic(fmt.Errorf("can't remove order which is not in this priceLevel. priceLevel: %s", p.price.String()))
	}

	order := orderItem.(*MemoryOrder)
	p.orderMap.Delete(order.ID)
	p.totalAmount = p.totalAmount.Sub(order.Amount)
}

func (p *priceLevel) GetOrder(id string) (order *MemoryOrder, exist bool) {
	orderItem, exist := p.orderMap.Get(id)
	if !exist {
		return nil, exist
	}

	return orderItem.(*MemoryOrder), exist
}

func (p *priceLevel) ChangeOrder(o *MemoryOrder, changeAmount decimal.Decimal) {
	_, ok := p.orderMap.Get(o.ID)

	if !ok {
		panic(fmt.Errorf("can't remove order which is not in this priceLevel. priceLevel: %s", p.price.String()))
	}

	p.totalAmount = p.totalAmount.Add(changeAmount)
}

func (p *priceLevel) Less(item llrb.Item) bool {
	another := item.(*priceLevel)
	return p.price.LessThan(another.price)
}

// Orderbook ...
type Orderbook struct {
	market string

	plugins []OrderbookPlugin

	bidsTree *llrb.LLRB
	asksTree *llrb.LLRB

	lock sync.RWMutex

	Sequence uint64
}

// NewOrderbook return a new book
func NewOrderbook(market string) *Orderbook {
	book := &Orderbook{
		plugins:  make([]OrderbookPlugin, 0, 3),
		market:   market,
		bidsTree: llrb.New(),
		asksTree: llrb.New(),
	}

	return book
}

func (book *Orderbook) SnapshotV2() *SnapshotV2 {
	//startTime := time.Now()

	book.lock.RLock()
	defer book.lock.RUnlock()

	//utils.Debug("== cost in lock, Snapshot : %f", float64(time.Since(startTime))/1000000)

	bids := make([][2]string, 0, 0)
	asks := make([][2]string, 0, 0)

	asyncWaitGroup := sync.WaitGroup{}

	asyncWaitGroup.Add(1)
	go func() {
		book.asksTree.AscendGreaterOrEqual(newPriceLevel(decimal.Zero), func(i llrb.Item) bool {
			pl := i.(*priceLevel)
			asks = append(asks, [2]string{pl.price.String(), pl.totalAmount.String()})
			return true
		})
		asyncWaitGroup.Done()
	}()

	asyncWaitGroup.Add(1)
	go func() {
		book.bidsTree.DescendLessOrEqual(newPriceLevel(decimal.New(1, 99)), func(i llrb.Item) bool {
			pl := i.(*priceLevel)
			bids = append(bids, [2]string{pl.price.String(), pl.totalAmount.String()})
			return true
		})
		asyncWaitGroup.Done()
	}()

	asyncWaitGroup.Wait()

	res := &SnapshotV2{
		Bids: bids,
		Asks: asks,
	}

	//utils.Debugf("== cost in lock read, Snapshot : %f", float64(time.Since(startTime))/1000000)

	return res
}

func (book *Orderbook) InsertOrder(order *MemoryOrder) *OrderbookEvent {
	startTime := time.Now().UTC()
	book.lock.Lock()
	defer book.lock.Unlock()

	log.Debug("cost in lock, InsertOrder :", order.ID, float64(time.Since(startTime))/1000000)

	var tree *llrb.LLRB
	if order.Side == "sell" {
		tree = book.asksTree
	} else {
		tree = book.bidsTree
	}

	price := tree.Get(newPriceLevel(order.Price))

	if price == nil {
		price = newPriceLevel(order.Price)
		tree.InsertNoReplace(price)
	}

	price.(*priceLevel).InsertOrder(order)

	orderBookEvent := &OrderbookEvent{
		OrderID: order.ID,
		Side:    order.Side,
		Amount:  order.Amount,
		Price:   order.Price,
	}

	book.RunPlugins(orderBookEvent)

	return orderBookEvent
}

func (book *Orderbook) RemoveOrder(order *MemoryOrder) *OrderbookEvent {
	book.lock.Lock()
	defer book.lock.Unlock()

	var tree *llrb.LLRB
	if order.Side == "sell" {
		tree = book.asksTree
	} else {
		tree = book.bidsTree
	}

	// log
	plItem := tree.Get(newPriceLevel(order.Price))
	if plItem == nil {
		log.Infof("plItem is nil when RemoveOrder")
		return nil
	}

	price := plItem.(*priceLevel)

	if price == nil {
		panic(fmt.Sprintf("pl is nil when RemoveOrder, book: %s, order: %+v", book.market, order))
	}

	price.RemoveOrder(order)
	if price.Len() <= 0 {
		tree.Delete(price)
	}

	event := &OrderbookEvent{
		OrderID: order.ID,
		Side:    order.Side,
		Amount:  order.Amount.Mul(decimal.New(-1, 0)),
		Price:   order.Price,
	}

	book.RunPlugins(event)

	return event
}

func (book *Orderbook) ChangeOrder(order *MemoryOrder, changeAmount decimal.Decimal) *OrderbookEvent {
	book.lock.Lock()
	defer book.lock.Unlock()

	var tree *llrb.LLRB
	if order.Side == "sell" {
		tree = book.asksTree
	} else {
		tree = book.bidsTree
	}

	price := tree.Get(newPriceLevel(order.Price))

	if price == nil {
		fmt.Println("book snapshot:", book.SnapshotV2())
		panic(fmt.Sprintf("can't change order which is not in this orderbook. book: %s, order: %+v", book.market, order))
	}

	price.(*priceLevel).ChangeOrder(order, changeAmount)

	event := &OrderbookEvent{
		OrderID: order.ID,
		Side:    order.Side,
		Amount:  changeAmount,
		Price:   order.Price,
	}
	book.RunPlugins(event)

	return event
}

func (book *Orderbook) UsePlugin(plugin OrderbookPlugin) {
	book.plugins = append(book.plugins, plugin)
}

func (book *Orderbook) RunPlugins(event *OrderbookEvent) {
	for _, plugin := range book.plugins {
		plugin(event)
	}
}

func (book *Orderbook) GetOrder(id string, side string, price decimal.Decimal) (*MemoryOrder, bool) {
	book.lock.Lock()
	defer book.lock.Unlock()

	var tree *llrb.LLRB
	if side == "sell" {
		tree = book.asksTree
	} else {
		tree = book.bidsTree
	}

	pl := tree.Get(newPriceLevel(price))

	if pl == nil {
		return nil, false
	}

	return pl.(*priceLevel).GetOrder(id)
}

// MaxBid ...
func (book *Orderbook) MaxBid() *decimal.Decimal {
	book.lock.Lock()
	defer book.lock.Unlock()

	maxItem := book.bidsTree.Max()
	if maxItem != nil {
		return &maxItem.(*priceLevel).price
	}
	return nil
}

// MinAsk ...
func (book *Orderbook) MinAsk() *decimal.Decimal {
	book.lock.Lock()
	defer book.lock.Unlock()

	minItem := book.asksTree.Min()

	if minItem != nil {
		return &minItem.(*priceLevel).price
	}

	return nil
}

func (book *Orderbook) CanMatch(order *MemoryOrder) bool {
	if strings.EqualFold("buy", order.Side) {
		minItem := book.asksTree.Min()
		if minItem == nil {
			return false
		}

		if order.Price.GreaterThanOrEqual(minItem.(*priceLevel).price) {
			return true
		}

		return false
	} else {
		maxItem := book.bidsTree.Max()
		if maxItem == nil {
			return false
		}

		if order.Price.LessThanOrEqual(maxItem.(*priceLevel).price) {
			return true
		}

		return false
	}
}

// return matching orders in book
// will NOT modify the order book
//
// amt is quoteCurrency when order is MarketID Buy Order
// all other amount is baseCurrencyAmt
func (book *Orderbook) MatchOrder(takerOrder *MemoryOrder, marketAmountDecimals int) *MatchResult {
	book.lock.Lock()
	defer book.lock.Unlock()

	matchedResult := make([]*MatchItem, 0)

	totalMatchedAmount := decimal.NewFromFloat(0)
	leftAmount := takerOrder.Amount

	// This function will be called multi times
	// Return false to break the loop
	limitOrderIterator := func(i llrb.Item) bool {
		pl := i.(*priceLevel)

		if takerOrder.Side == "buy" && pl.price.GreaterThan(takerOrder.Price) {
			return false
		} else if takerOrder.Side == "sell" && pl.price.LessThan(takerOrder.Price) {
			return false
		}

		iter := pl.orderMap.IterFunc()
		for kv, ok := iter(); ok; kv, ok = iter() {
			if leftAmount.LessThanOrEqual(decimal.Zero) {
				break
			}

			bookOrder := kv.Value.(*MemoryOrder)

			if leftAmount.GreaterThanOrEqual(bookOrder.Amount) {
				matchedAmount := bookOrder.Amount

				matchedItem := &MatchItem{
					MatchedAmount: matchedAmount,
					MakerOrder:    bookOrder,
				}

				matchedResult = append(matchedResult, matchedItem)
				totalMatchedAmount = totalMatchedAmount.Add(matchedAmount)
				leftAmount = leftAmount.Sub(matchedAmount)
			} else {
				eatAmount := leftAmount
				matchedItem := &MatchItem{
					MatchedAmount: eatAmount,
					MakerOrder:    bookOrder,
				}

				matchedResult = append(matchedResult, matchedItem)
				totalMatchedAmount = totalMatchedAmount.Add(eatAmount)
				leftAmount = decimal.Zero
			}
		}

		return leftAmount.GreaterThan(decimal.Zero)
	}

	marketOrderIterator := func(i llrb.Item) bool {
		pl := i.(*priceLevel)

		iter := pl.orderMap.IterFunc()
		for kv, ok := iter(); ok; kv, ok = iter() {
			// break when no leftAmount
			if leftAmount.LessThanOrEqual(decimal.Zero) {
				return false
			}

			// for marketOrder with price limit
			if takerOrder.Price.GreaterThan(decimal.Zero) {
				if takerOrder.Side == "buy" && pl.price.GreaterThan(takerOrder.Price) {
					utils.Infof("market buy exit early for price bound: %s", takerOrder.Price)

					return false
				} else if takerOrder.Side == "sell" && pl.price.LessThan(takerOrder.Price) {
					utils.Infof("market sell exit early for price bound: %s", takerOrder.Price)

					return false
				}
			}

			memoryOrder := kv.Value.(*MemoryOrder)

			matchedItem := &MatchItem{
				MakerOrder: memoryOrder,
			}

			// for market order buy, leftAmount is quoteCurrencyAmount
			if takerOrder.Side == "buy" {
				//price = wethAmt / hotAmt
				makerQuoteCurrencyAmt := memoryOrder.Amount.Mul(memoryOrder.Price)

				if leftAmount.GreaterThanOrEqual(makerQuoteCurrencyAmt) {
					//can take this whole maker order
					matchedItem.MatchedAmount = memoryOrder.Amount
					leftAmount = leftAmount.Sub(makerQuoteCurrencyAmt)
				} else {
					// can take part of this order, round down with marketAmountDecimals
					eatBaseCurrencyAmt := leftAmount.DivRound(memoryOrder.Price, int32(marketAmountDecimals)+1).Truncate(int32(marketAmountDecimals))

					matchedItem.MatchedAmount = eatBaseCurrencyAmt
					leftAmount = decimal.Zero
				}
			} else {
				// for sell, leftAmount is baseCurrencyAmount
				if leftAmount.GreaterThanOrEqual(memoryOrder.Amount) {
					matchedItem.MatchedAmount = memoryOrder.Amount
					leftAmount = leftAmount.Sub(memoryOrder.Amount)
				} else {
					matchedItem.MatchedAmount = leftAmount
					leftAmount = decimal.Zero
				}
			}

			matchedResult = append(matchedResult, matchedItem)

			utils.Infof("matchedItem.MatchedAmount: %s", matchedItem.MatchedAmount)
			totalMatchedAmount = totalMatchedAmount.Add(matchedItem.MatchedAmount)
		}

		return leftAmount.GreaterThan(decimal.Zero)
	}

	// decide iterator
	var iterator llrb.ItemIterator
	if takerOrder.Type == "market" {
		iterator = marketOrderIterator
	} else {
		iterator = limitOrderIterator
	}

	if takerOrder.Side == "sell" {
		book.bidsTree.DescendLessOrEqual(newPriceLevel(decimal.New(1, 99)), iterator)
	} else {
		book.asksTree.AscendGreaterOrEqual(newPriceLevel(decimal.Zero), iterator)
	}

	return &MatchResult{
		MatchItems: matchedResult,
		TakerOrder: takerOrder,
	}
}

func (book *Orderbook) ExecuteMatch(takerOrder *MemoryOrder, marketAmountDecimals int) *MatchResult {
	result := book.MatchOrder(takerOrder, marketAmountDecimals)

	cancelSmallMatchesIfExist(result)

	for _, item := range result.MatchItems {
		var e *OrderbookEvent

		// after match, gasFee is paid
		if !item.MatchShouldBeCanceled && item.MatchedAmount.IsPositive() {
			item.MakerOrder.GasFeeAmount = decimal.Zero
		}

		if makerOrderShouldBeRemovedAfterMatch(takerOrder.GasFeeAmount, takerOrder.TakerFeeRate, item) {
			e = book.RemoveOrder(item.MakerOrder)
			item.MakerOrder.Amount = decimal.Zero

			item.MakerOrderIsDone = true
		} else {
			changeAmt := item.MatchedAmount

			e = book.ChangeOrder(item.MakerOrder, changeAmt.Mul(decimal.New(-1, 0)))
			item.MakerOrder.Amount = item.MakerOrder.Amount.Sub(changeAmt)
		}

		msg := OrderBookChangeMessage(book.market, book.Sequence, e.Side, e.Price, e.Amount)
		result.OrderBookActivities = append(result.OrderBookActivities, msg)
	}

	return result
}

// when makerOrder is sell
// one cases when maker order should be removed
// 1. all matched - no remaining amount left
//
// when makerOrder is buy
// two cases when maker order should be removed
// 1. all matched - no remaining amount left
// 2. remaining amount too small
func makerOrderShouldBeRemovedAfterMatch(assumedTakerOrderGasFee, assumedTakerOrderFeeRate decimal.Decimal, item *MatchItem) bool {
	remainingAmtInQuote := item.MakerOrder.Amount.Sub(item.MatchedAmount).Mul(item.MakerOrder.Price)

	return orderShouldBeRemoved(assumedTakerOrderGasFee, assumedTakerOrderFeeRate, remainingAmtInQuote, item.MakerOrder.Side)
}

func TakerOrderShouldBeRemoved(taker *MemoryOrder) bool {
	remainingAmtInQuote := taker.Amount.Mul(taker.Price)

	return orderShouldBeRemoved(taker.GasFeeAmount, taker.TakerFeeRate, remainingAmtInQuote, taker.Side)
}

// whether order should be removed depends on it's taker order,
// so we need assumedTakerGasFee & assumedTakerFeeRate
func orderShouldBeRemoved(assumedTakerGasFee, assumedTakerFeeRate, remainingAmtInQuote decimal.Decimal, orderSide string) bool {
	if orderSide == "sell" {
		return remainingAmtInQuote.LessThanOrEqual(decimal.Zero)
	} else {
		// take away taker's gas & tradeFee
		subtractAmtFromTaker := assumedTakerGasFee.Add(remainingAmtInQuote.Mul(assumedTakerFeeRate))

		return remainingAmtInQuote.LessThanOrEqual(decimal.Zero) || remainingAmtInQuote.Sub(subtractAmtFromTaker).IsNegative()
	}
}

// small matches should be canceled to avoid transaction revert
func cancelSmallMatchesIfExist(matchResult *MatchResult) (canceledAmtSum decimal.Decimal) {

	if matchResult.TakerOrder.Side == "buy" {
		// taker buy, every match > gas + tradeFee
		for _, matchItem := range matchResult.MatchItems {
			tradeAmt := matchItem.MatchedAmount.Mul(matchItem.MakerOrder.Price)
			// subtract taker's gas & tradeFee
			subtractAmt := matchItem.MakerOrder.GasFeeAmount.Add(matchItem.MakerOrder.MakerFeeRate.Mul(tradeAmt))

			if tradeAmt.LessThan(subtractAmt) {
				amt := cancelMatch(matchItem)
				canceledAmtSum = canceledAmtSum.Add(amt)
			}
		}
	} else {
		tradeAmtInQuoteToken := matchResult.QuoteTokenTotalMatchedAmt()
		// subtract taker's gas & tradeFee
		subtractAmt := matchResult.TakerOrder.GasFeeAmount.Add(matchResult.TakerTradeFeeInQuoteToken())

		if tradeAmtInQuoteToken.LessThan(subtractAmt) {
			canceledAmtSum = cancelAllMatches(matchResult)
		}
	}

	return
}

func cancelMatch(match *MatchItem) (cancelAmt decimal.Decimal) {
	match.MatchShouldBeCanceled = true

	return match.MatchedAmount
}

func cancelAllMatches(match *MatchResult) (sum decimal.Decimal) {
	for _, item := range match.MatchItems {
		amt := cancelMatch(item)

		sum = sum.Add(amt)
	}

	return sum
}
