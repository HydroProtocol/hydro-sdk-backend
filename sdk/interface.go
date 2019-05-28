package sdk

import (
	"math/big"

	"github.com/shopspring/decimal"
)

type BlockChain interface {
	GetTokenBalance(tokenAddress, address string) decimal.Decimal
	GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal
	GetHotFeeDiscount(address string) decimal.Decimal

	GetBlockNumber() (uint64, error)
	GetBlockByNumber(blockNumber uint64) (Block, error)

	GetTransaction(ID string) (Transaction, error)
	GetTransactionReceipt(ID string) (TransactionReceipt, error)
	GetTransactionAndReceipt(ID string) (Transaction, TransactionReceipt, error)

	IsValidSignature(address string, message string, signature string) (bool, error)
	//GetTransactionCount(address string) (int, error)

	SendTransaction(txAttributes map[string]interface{}, privateKey []byte) (transactionHash string, err error)
	SendRawTransaction(tx interface{}) (string, error)
}

type HydroProtocol interface {
	GenerateOrderData(version, expiredAtSeconds, salt int64, asMakerFeeRate, asTakerFeeRate, makerRebateRate decimal.Decimal, isSell, isMarket, isMakerOnly bool) string
	GetOrderHash(*Order) []byte
	GetMatchOrderCallData(*Order, []*Order, []*big.Int) []byte

	IsValidOrderSignature(address string, orderID string, signature string) bool
}

type Hydro interface {
	HydroProtocol
	BlockChain
}

type Block interface {
	Number() uint64
	Timestamp() uint64
	GetTransactions() []Transaction

	Hash() string
	ParentHash() string
}

type Transaction interface {
	GetBlockHash() string
	GetBlockNumber() uint64
	GetFrom() string
	GetGas() int
	GetGasPrice() big.Int
	GetHash() string
	GetTo() string
	GetValue() big.Int
}

type TransactionReceipt interface {
	GetResult() bool
	GetBlockNumber() uint64

	GetBlockHash() string
	GetTxHash() string
	GetTxIndex() int

	GetLogs() []IReceiptLog
}

type IReceiptLog interface {
	GetRemoved() bool
	GetLogIndex() int
	GetTransactionIndex() int
	GetTransactionHash() string
	GetBlockNum() int
	GetBlockHash() string
	GetAddress() string
	GetData() string
	GetTopics() []string
}

type (
	OrderSignature struct {
		Config [32]byte
		R      [32]byte
		S      [32]byte
	}

	Order struct {
		Trader           string
		BaseTokenAmount  *big.Int
		QuoteTokenAmount *big.Int
		GasTokenAmount   *big.Int
		Data             string
		Signature        string

		Relayer           string
		BaseTokenAddress  string
		QuoteTokenAddress string
	}

	OrderParam struct {
		Trader           string          `json:"trader"`
		BaseTokenAmount  *big.Int        `json:"base_token_amount"`
		QuoteTokenAmount *big.Int        `json:"quote_token_amount"`
		GasTokenAmount   *big.Int        `json:"gas_token_amount"`
		Data             string          `json:"data"`
		Signature        *OrderSignature `json:"signature"`
	}

	OrderAddressSet struct {
		BaseToken  string `json:"baseToken"`
		QuoteToken string `json:"quoteToken"`
		Relayer    string `json:"relayer"`
	}
)

func NewOrderWithData(
	trader, relayer, baseTokenAddress, quoteTokenAddress string,
	baseTokenAmount, quoteTokenAmount, gasTokenAddress *big.Int,
	data string,
	signature string,
) *Order {
	return &Order{
		Trader:            trader,
		Relayer:           relayer,
		BaseTokenAmount:   baseTokenAmount,
		QuoteTokenAmount:  quoteTokenAmount,
		BaseTokenAddress:  baseTokenAddress,
		QuoteTokenAddress: quoteTokenAddress,
		GasTokenAmount:    gasTokenAddress,
		Data:              data,
		Signature:         signature,
	}
}
