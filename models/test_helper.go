package models

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/config"
	uuid2 "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

type MMarketDao struct {
	mock.Mock
}

func (m *MMarketDao) FindAllMarkets() []*Market {
	args := m.Called()
	return args.Get(0).([]*Market)
}

func (m *MMarketDao) FindMarketByID(marketID string) *Market {
	args := m.Called(marketID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*Market)
}

func (m *MMarketDao) InsertMarket(market *Market) error {
	args := m.Called(market)
	return args.Get(0).(error)
}

type MTradeDao struct {
	mock.Mock
}

func (m *MTradeDao) FindTradesByMarket(pair string, startTime time.Time, endTime time.Time) []*Trade {
	args := m.Called(pair, startTime, endTime)
	return args.Get(0).([]*Trade)
}

func (m *MTradeDao) FindAllTrades(marketID string) (int64, []*Trade) {
	args := m.Called(marketID)
	return args.Get(0).(int64), args.Get(1).([]*Trade)
}

func (m *MTradeDao) FindTradesByHash(hash string) []*Trade {
	args := m.Called(hash)
	return args.Get(0).([]*Trade)
}

func (m *MTradeDao) FindTradeByID(id int64) *Trade {
	args := m.Called(id)
	return args.Get(0).(*Trade)
}

func (m *MTradeDao) FindAccountMarketTrades(account, marketID, status string, limit, offset int) (int64, []*Trade) {
	args := m.Called(account, status, limit, offset)
	return args.Get(0).(int64), args.Get(1).([]*Trade)
}

func (m *MTradeDao) InsertTrade(trade *Trade) error {
	args := m.Called(trade)
	return args.Error(0)
}

func (m *MTradeDao) UpdateTrade(trade *Trade) error {
	args := m.Called(trade)
	return args.Get(0).(error)
}

func (m *MTradeDao) Count() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MTradeDao) FindTradeByTransactionID(transactionID int64) []*Trade {
	args := m.Called(transactionID)
	return args.Get(0).([]*Trade)
}

type MErc20 struct {
	mock.Mock
}

func (m *MErc20) BalanceOf(contract, owner string) (decimal.Decimal, error) {
	args := m.Called(contract, owner)
	if args.Get(1) == nil {
		return args.Get(0).(decimal.Decimal), nil
	}
	return args.Get(0).(decimal.Decimal), args.Get(1).(error)
}

func (m *MErc20) Allowance(contract, owner, spender string) (decimal.Decimal, error) {
	args := m.Called(contract, owner, spender)
	if args.Get(1) == nil {
		return args.Get(0).(decimal.Decimal), nil
	}
	return args.Get(0).(decimal.Decimal), args.Get(1).(error)
}

type MLockedBalanceDao struct {
	mock.Mock
}

func (m *MLockedBalanceDao) GetByAccountAndSymbol(account, tokenSymbol string, decimals int) decimal.Decimal {
	args := m.Called(account, tokenSymbol, decimals)
	return args.Get(0).(decimal.Decimal)
}

type MCache struct {
	mock.Mock
}

func (m *MCache) Set(key string, value string, expire time.Duration) error {
	args := m.Called(key, value, expire)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MCache) Get(key string) (string, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return args.Get(0).(string), nil
	}
	return args.Get(0).(string), args.Get(1).(error)
}

func (m *MCache) Push(key []byte) error {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MCache) Pop() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), args.Get(1).(error)
}

// default models
func MarketHotDai() *Market {
	marketHotDai := &Market{
		ID:                 "HOT-DAI",
		BaseTokenSymbol:    "HOT",
		BaseTokenAddress:   config.Getenv("HSK_WETH_TOKEN_ADDRESS"),
		BaseTokenDecimals:  18,
		QuoteTokenSymbol:   "DAI",
		QuoteTokenAddress:  config.Getenv("HSK_USD_TOKEN_ADDRESS"),
		QuoteTokenDecimals: 18,
		MinOrderSize:       decimal.NewFromFloat(0.1),
		PricePrecision:     5,
		PriceDecimals:      5,
		AmountDecimals:     5,
		MakerFeeRate:       decimal.NewFromFloat(0.001),
		TakerFeeRate:       decimal.NewFromFloat(0.003),
		GasUsedEstimation:  250000,
	}

	return marketHotDai
}

func MockMarketDao() {
	marketDao := &MMarketDao{}

	MarketDao = marketDao
	var markets []*Market

	marketWethDai := &Market{
		ID:                 "WETH-DAI",
		BaseTokenSymbol:    "WETH",
		BaseTokenAddress:   config.Getenv("HSK_WETH_TOKEN_ADDRESS"),
		BaseTokenDecimals:  18,
		QuoteTokenSymbol:   "DAI",
		QuoteTokenAddress:  config.Getenv("HSK_USD_TOKEN_ADDRESS"),
		QuoteTokenDecimals: 18,
		MinOrderSize:       decimal.NewFromFloat(0.1),
		PricePrecision:     5,
		PriceDecimals:      5,
		AmountDecimals:     5,
		MakerFeeRate:       decimal.NewFromFloat(0.001),
		TakerFeeRate:       decimal.NewFromFloat(0.003),
		GasUsedEstimation:  250000,
	}

	marketHotDai := &Market{
		ID:                 "HOT-DAI",
		BaseTokenSymbol:    "HOT",
		BaseTokenAddress:   config.Getenv("HSK_WETH_TOKEN_ADDRESS"),
		BaseTokenDecimals:  18,
		QuoteTokenSymbol:   "DAI",
		QuoteTokenAddress:  config.Getenv("HSK_USD_TOKEN_ADDRESS"),
		QuoteTokenDecimals: 18,
		MinOrderSize:       decimal.NewFromFloat(0.1),
		PricePrecision:     5,
		PriceDecimals:      5,
		AmountDecimals:     5,
		MakerFeeRate:       decimal.NewFromFloat(0.001),
		TakerFeeRate:       decimal.NewFromFloat(0.003),
		GasUsedEstimation:  250000,
	}
	markets = append(markets, marketWethDai)
	markets = append(markets, marketHotDai)

	marketDao.On("FindAllMarkets").Return(markets).Once()
	marketDao.On("FindMarketByID", mock.MatchedBy(func(marketID string) bool { return marketID == "WETH-DAI" })).Return(marketWethDai)
	marketDao.On("FindMarketByID", "HOT-DAI").Times(10).Return(marketHotDai)
	marketDao.On("FindMarketByID", mock.AnythingOfType("string")).Times(10).Return(nil)
}

func MockTradeDao() {
	tradeDao := &MTradeDao{}
	TradeDao = tradeDao
	var tradesWethDai []*Trade
	var tradesHotDai []*Trade

	trade1 := getMockTradeWithTime("WETH-DAI", true, time.Now().Add(-time.Hour*1))
	trade2 := getMockTradeWithTime("WETH-DAI", true, time.Now().Add(-time.Hour*2))
	trade3 := getMockTradeWithTime("WETH-DAI", false, time.Now().Add(-time.Hour*3))

	trade4 := getMockTradeWithTime("HOT-DAI", true, time.Now().Add(-time.Hour*1))
	trade5 := getMockTradeWithTime("HOT-DAI", true, time.Now().Add(-time.Hour*2))
	trade6 := getMockTradeWithTime("HOT-DAI", false, time.Now().Add(-time.Hour*3))

	tradesWethDai = append(tradesWethDai, trade1)
	tradesWethDai = append(tradesWethDai, trade2)
	tradesWethDai = append(tradesWethDai, trade3)

	tradesHotDai = append(tradesHotDai, trade4)
	tradesHotDai = append(tradesHotDai, trade5)
	tradesHotDai = append(tradesHotDai, trade6)

	tradeDao.On("FindTradesByMarket", "WETH-DAI", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(tradesWethDai).Once()
	tradeDao.On("FindTradesByMarket", "HOT-DAI", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(tradesWethDai).Once()
}

func getMockTradeWithTime(marketID string, success bool, time time.Time) *Trade {
	status := common.STATUS_SUCCESSFUL

	if !success {
		status = common.STATUS_FAILED
	}

	trade := Trade{
		ID:              rand.Int63(),
		TransactionHash: "0x17e16163f030936110cc4b548ac53fd96e963f17437f99d28df8137a2a680378",
		Status:          status,
		MarketID:        marketID,
		Maker:           "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
		Taker:           "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
		TakerSide:       "buy",
		MakerOrderID:    uuid2.NewV4().String(),
		TakerOrderID:    uuid2.NewV4().String(),
		Sequence:        0,
		Amount:          decimal.NewFromFloat(123.456),
		Price:           decimal.NewFromFloat(0.789),
		ExecutedAt:      time,
		CreatedAt:       time,
	}

	return &trade
}
func InitTestDB() {
	ConnectDatabase("postgres", os.Getenv("HSK_DATABASE_URL"))
	bts, err := ioutil.ReadFile("../db/migrations/0001-init.up.sql")

	if err != nil {
		panic(err)
	}

	_, _ = DB.Exec(string(bts))
}
