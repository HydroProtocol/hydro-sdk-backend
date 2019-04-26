package ethereum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/onrik/ethrpc"
	"math"
	"math/big"
	"os"
	"strconv"
	"unicode/utf8"
)

type IErc20 interface {
	Symbol(address string) (error, string)
	Decimals(address string) (error, int)
	Name(address string) (error, string)
	TotalSupply(address string) (error, *big.Int)
}

type Erc20Service struct {
	client *ethrpc.EthRPC
}

func NewErc20Service(client *ethrpc.EthRPC) IErc20 {
	if client == nil {
		blockChainNodeUrl := os.Getenv("HSK_BLOCKCHAIN_RPC_URL")
		if len(blockChainNodeUrl) == 0 {
			panic(errors.New("empty env HSK_BLOCKCHAIN_RPC_URL"))
		}
		client = ethrpc.New(blockChainNodeUrl)
	}

	return &Erc20Service{
		client: client,
	}
}

func (e *Erc20Service) TotalSupply(address string) (error, *big.Int) {
	result := callContract(e, address, ERC20TotalSupply)
	value := parseBigIntResult(result)

	if value.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("cannot find TotalSupply by address %s on chain", address), big.NewInt(-1)
	}

	return nil, value
}

func (e *Erc20Service) Symbol(address string) (error, string) {
	result := callContract(e, address, ERC20Symbol)
	retStr := parseStringResult(result)

	if retStr == "" {
		return fmt.Errorf("cannot find Symbol by address %s on chain", address), retStr
	}
	return nil, tuncate(retStr, 90)
}

func (e *Erc20Service) Decimals(address string) (error, int) {
	result := callContract(e, address, ERC20Decimals)
	value := parseIntResult(result)

	if value > math.MaxInt8 || value < 0 {
		return fmt.Errorf("cannot find Decimals by address %s on chain", address), -1
	}

	return nil, value
}

func (e *Erc20Service) Name(address string) (error, string) {
	result := callContract(e, address, ERC20Name)
	retStr := parseStringResult(result)

	if retStr == "" {
		return fmt.Errorf("cannot find Name by address %s on chain", address), retStr
	}
	return nil, tuncate(retStr, 250)
}

// sha3 result of erc20 method
const (
	ERC20TotalSupply = "0x18160ddd" // totalSupply()
	ERC20Symbol      = "0x95d89b41" // symbol()
	ERC20Name        = "0x06fdde03" // name()
	ERC20Decimals    = "0x313ce567" // decimals()
	//ERC20BalanceOf     = "0x70a08231"                                                         // balanceOf(address)
	//ERC20Allowance     = "0xdd62ed3e"                                                         // allowance(address,address)
	//ERC20Transfer      = "0xa9059cbb"                                                         // transfer(address,uint256)
	//ERC20Approve       = "0x095ea7b3"                                                         // approve(address,uint256)
	//ERC20TransferFrom  = "0x23b872dd"                                                         // transferFrom(address,address,uint256)
	//ERC20TransferEvent = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" // Transfer(address,address,uint256)
	//ERC20ApprovalEvent = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925" // Approval(address,address,uint256)
	//WETHWithdrawal     = "0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65" // Withdrawal(index_topic_1 address src, uint256 wad)
	//WETHDeposit        = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c" // Deposit(index_topic_1 address dst, uint256 wad)
)

// erc20 name, symbol
// about how to parse return value
// please read more: http://solidity.readthedocs.io/en/latest/abi-spec.html#examples
func parseStringResult(result string) (res string) {
	defer func() {
		if err := recover(); err != nil {
			res = ""
		}
	}()

	result = removeLeading0x(result)

	startPosition, err := strconv.ParseInt(result[:64], 16, 64)
	if err != nil {
		panic(err)
	}
	startPosition = startPosition * 2 // byte length to hex length, 32 bytes is 64 hex

	length, err := strconv.ParseInt(result[startPosition:startPosition+64], 16, 64)
	if err != nil {
		panic(err)
	}

	b, err := hex.DecodeString(result[startPosition+64 : startPosition+64+length*2])

	if err != nil {
		panic(err)
	}

	b = bytes.Replace(b, []byte{0}, nil, -1)

	str := string(b)

	if !utf8.ValidString(str) {
		panic(fmt.Errorf("invalid utf8 string %+v", str))
	}

	return str
}

func parseIntResult(result string) (res int) {
	defer func() {
		if err := recover(); err != nil {
			res = -1
		}
	}()

	result = removeLeading0x(result)

	decimals, err := strconv.ParseInt(result, 16, 64)

	if err != nil {
		panic(err)
	}

	return int(decimals)
}

func parseBigIntResult(result string) *big.Int {
	if len(removeLeading0x(result)) == 0 || len(result) > 66 {
		return big.NewInt(-1)
	}

	n := new(big.Int)
	n.SetString(result, 0)
	return n
}

func callContract(e *Erc20Service, address string, data string) string {
	result, err := e.client.EthCall(ethrpc.T{
		From: "0x0000000000000000000000000000000000000000",
		To:   address,
		Data: data,
	}, "latest")

	if err != nil {
		panic(err)
	}

	return result
}

func tuncate(str string, len int) string {
	if utf8.RuneCountInString(str) > len {
		return string(([]rune(str))[0:len])
	}

	return str
}

func removeLeading0x(s string) string {
	if s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
