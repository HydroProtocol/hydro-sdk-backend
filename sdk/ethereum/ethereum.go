package ethereum

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/rlp"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/signer"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/labstack/gommon/log"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
	"os"
	"strconv"
	"strings"
)

var EIP712_DOMAIN_TYPEHASH []byte
var EIP712_ORDER_TYPE []byte

// compile time interface check
var _ sdk.BlockChain = &Ethereum{}
var _ sdk.HydroProtocol = &EthereumHydroProtocol{}
var _ sdk.Hydro = &EthereumHydro{}

func init() {
	EIP712_DOMAIN_TYPEHASH = crypto.Keccak256([]byte(`EIP712Domain(string name)`))
	EIP712_ORDER_TYPE = crypto.Keccak256([]byte(`Order(address trader,address relayer,address baseToken,address quoteToken,uint256 baseTokenAmount,uint256 quoteTokenAmount,uint256 gasTokenAmount,bytes32 data)`))
}

type EthereumBlock struct {
	*ethrpc.Block
}

func (block *EthereumBlock) GetTransactions() []sdk.Transaction {
	txs := make([]sdk.Transaction, 0, 20)

	for i := range block.Block.Transactions {
		tx := block.Block.Transactions[i]
		txs = append(txs, &EthereumTransaction{&tx})
	}

	return txs
}

func (block *EthereumBlock) Number() uint64 {
	return uint64(block.Block.Number)
}

func (block *EthereumBlock) Timestamp() uint64 {
	return uint64(block.Block.Timestamp)
}

type EthereumTransaction struct {
	*ethrpc.Transaction
}

func (t *EthereumTransaction) GetHash() string {
	return t.Hash
}

type EthereumTransactionReceipt struct {
	*ethrpc.TransactionReceipt
}

func (r *EthereumTransactionReceipt) GetResult() bool {
	res, err := strconv.ParseInt(r.Status, 0, 64)

	if err != nil {
		panic(err)
	}

	return res == 1
}

func (r *EthereumTransactionReceipt) GetBlockNumber() uint64 {
	return uint64(r.BlockNumber)
}

type Ethereum struct {
	client       *ethrpc.EthRPC
	hybridExAddr string
}

func (e *Ethereum) EnableDebug(b bool) {
	e.client.Debug = b
}

func (e *Ethereum) GetBlockByNumber(number uint64) (sdk.Block, error) {

	block, err := e.client.EthGetBlockByNumber(int(number), true)

	if err != nil {
		log.Errorf("get Block by Number failed %+v", err)
		return nil, err
	}

	if block == nil {
		log.Errorf("get Block by Number returns nil block for num: %d", number)
		return nil, errors.New("get Block by Number returns nil block for num: " + strconv.Itoa(int(number)))
	}

	return &EthereumBlock{block}, nil
}

func (e *Ethereum) GetBlockNumber() (uint64, error) {
	number, err := e.client.EthBlockNumber()

	if err != nil {
		log.Errorf("GetBlockNumber failed, %v", err)
		return 0, err
	}

	return uint64(number), nil
}

func (e *Ethereum) GetTransaction(ID string) (sdk.Transaction, error) {
	tx, err := e.client.EthGetTransactionByHash(ID)

	if err != nil {
		log.Errorf("GetTransaction failed, %v", err)
		return nil, err
	}

	return &EthereumTransaction{tx}, nil
}

func signTransaction(tx *types.Transaction, pkHex string) string {
	privateKey, _ := crypto.NewPrivateKeyByHex(pkHex)
	signTx, err := signer.SignTx(tx, privateKey)

	if err != nil {
		panic(err)
	}

	rlpBytes := rlp.Encode([]interface{}{
		rlp.EncodeUint64ToBytes(signTx.Nonce),
		signTx.GasPrice.Bytes(),
		rlp.EncodeUint64ToBytes(signTx.Nonce),
		utils.Hex2Bytes(signTx.To[2:]),
		signTx.Value.Bytes(),
		signTx.Data,
		signTx.Signature[64:],
		signTx.Signature[0:32],
		signTx.Signature[32:64],
	})

	return utils.Bytes2HexP(rlpBytes)
}

func (e *Ethereum) SendTransaction(txAttributes map[string]interface{}, privateKey []byte) (transactionHash string, err error) {
	tx := types.NewTransaction(
		txAttributes["nonce"].(uint64),
		txAttributes["to"].(string),
		utils.DecimalToBigInt(txAttributes["value"].(decimal.Decimal)),
		txAttributes["gasLimit"].(uint64),
		utils.DecimalToBigInt(txAttributes["gasPrice"].(decimal.Decimal)),
		txAttributes["data"].([]byte),
	)

	pkHex := hex.EncodeToString(privateKey)
	rawTransactionString := signTransaction(tx, pkHex)

	return e.client.EthSendRawTransaction(rawTransactionString)
}

func (e *Ethereum) GetTransactionReceipt(ID string) (sdk.TransactionReceipt, error) {
	txReceipt, err := e.client.EthGetTransactionReceipt(ID)

	if err != nil {
		log.Errorf("GetTransactionReceipt failed, %v", err)
		return nil, err
	}

	return &EthereumTransactionReceipt{txReceipt}, nil
}

func (e *Ethereum) GetTransactionAndReceipt(ID string) (sdk.Transaction, sdk.TransactionReceipt, error) {
	txReceiptChannel := make(chan sdk.TransactionReceipt)

	go func() {
		rec, _ := e.GetTransactionReceipt(ID)
		txReceiptChannel <- rec
	}()

	txInfoChannel := make(chan sdk.Transaction)
	go func() {
		tx, _ := e.GetTransaction(ID)
		txInfoChannel <- tx
	}()

	return <-txInfoChannel, <-txReceiptChannel, nil
}

func (e *Ethereum) GetTokenBalance(tokenAddress, address string) decimal.Decimal {
	res, err := e.client.EthCall(ethrpc.T{
		To:   tokenAddress,
		From: address,
		Data: fmt.Sprintf("0x70a08231000000000000000000000000%s", without0xPrefix(address)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res)
}

func without0xPrefix(address string) string {
	if address[:2] == "0x" {
		address = address[2:]
	}

	return address
}

func (e *Ethereum) GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal {
	res, err := e.client.EthCall(ethrpc.T{
		To:   tokenAddress,
		From: address,
		Data: fmt.Sprintf("0xdd62ed3e000000000000000000000000%s000000000000000000000000%s", without0xPrefix(address), without0xPrefix(proxyAddress)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res)
}

func (e *Ethereum) GetHotFeeDiscount(address string) decimal.Decimal {
	if address == "" {
		return decimal.New(1, 0)
	}

	from := address

	res, err := e.client.EthCall(ethrpc.T{
		To:   e.hybridExAddr,
		From: from,
		Data: fmt.Sprintf("0x4376abf1000000000000000000000000%s", without0xPrefix(address)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res).Div(decimal.New(1, 2))
}

func (e *Ethereum) IsValidSignature(address string, message string, signature string) (bool, error) {
	if len(address) != 42 {
		return false, errors.New("address must be 42 size long")
	}

	if len(signature) != 132 {
		return false, errors.New("signature must be 132 size long")
	}

	var hashBytes []byte
	if strings.HasPrefix(message, "0x") {
		hashBytes = utils.Hex2Bytes(message[2:])
	} else {
		hashBytes = []byte(message)
	}

	signatureByte := utils.Hex2Bytes(signature[2:])
	pk, err := crypto.PersonalEcRecover(hashBytes, signatureByte)

	if err != nil {
		return false, err
	}

	return "0x"+strings.ToLower(pk) == strings.ToLower(address), nil
}

func (e *Ethereum) SendRawTransaction(tx interface{}) (string, error) {
	rawTransaction := tx.(string)
	return e.client.EthSendRawTransaction(rawTransaction)
}

func (e *Ethereum) GetTransactionCount(address string) (int, error) {
	return e.client.EthGetTransactionCount(address, "latest")
}

func NewEthereum(rpcUrl string, hybridExAddr string) *Ethereum {
	if hybridExAddr == "" {
		hybridExAddr = os.Getenv("HSK_HYBRID_EXCHANGE_ADDRESS")
	}

	if hybridExAddr == "" {
		panic(fmt.Errorf("NewEthereum need argument hybridExAddr"))
	}

	return &Ethereum{
		client:       ethrpc.New(rpcUrl),
		hybridExAddr: hybridExAddr,
	}
}

func IsValidSignature(address string, message string, signature string) (bool, error) {
	return new(Ethereum).IsValidSignature(address, message, signature)
}

func PersonalSign(message []byte, privateKey string) ([]byte, error) {
	return crypto.PersonalSign(message, privateKey)
}
