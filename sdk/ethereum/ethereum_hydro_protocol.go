package ethereum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/HydroProtocol/hydro-sdk-backend/config"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
)

type EthereumHydroProtocol struct{}

func (*EthereumHydroProtocol) GenerateOrderData(version, expiredAtSeconds, salt int64, asMakerFeeRate, asTakerFeeRate, makerRebateRate decimal.Decimal, isSell, isMarket, isMakerOnly bool) string {
	data := strings.Builder{}
	data.WriteString("0x")
	data.WriteString(addLeadingZero(strconv.FormatInt(version, 10), 2))
	if isSell {
		data.WriteString("01")
	} else {
		data.WriteString("00")
	}

	if isMarket {
		data.WriteString("01")
	} else {
		data.WriteString("00")
	}

	data.WriteString(addTailingZero(fmt.Sprintf("%x", expiredAtSeconds), 5*2))
	data.WriteString(addLeadingZero(utils.Bytes2Hex(utils.DecimalToBigInt(asMakerFeeRate.Mul(decimal.New(FeeRateBase, 0))).Bytes()), 2*2))
	data.WriteString(addLeadingZero(utils.Bytes2Hex(utils.DecimalToBigInt(asTakerFeeRate.Mul(decimal.New(FeeRateBase, 0))).Bytes()), 2*2))
	data.WriteString(addLeadingZero(utils.Bytes2Hex(utils.DecimalToBigInt(makerRebateRate.Mul(decimal.New(FeeRateBase, 0))).Bytes()), 2*2))
	data.WriteString(addLeadingZero(fmt.Sprintf("%x", salt), 8*2))

	if isMakerOnly {
		data.WriteString("01")
	} else {
		data.WriteString("00")
	}

	return addTailingZero(data.String(), 66)
}

func (*EthereumHydroProtocol) GetOrderHash(order *sdk.Order) []byte {
	return getEIP712MessageHash(
		crypto.Keccak256(
			EIP712_ORDER_TYPE,
			types.HexToHash(order.Trader).Bytes(),
			types.HexToHash(order.Relayer).Bytes(),
			types.HexToHash(order.BaseTokenAddress).Bytes(),
			types.HexToHash(order.QuoteTokenAddress).Bytes(),
			types.BytesToHash(order.BaseTokenAmount.Bytes()).Bytes(),
			types.BytesToHash(order.QuoteTokenAmount.Bytes()).Bytes(),
			types.BytesToHash(order.GasTokenAmount.Bytes()).Bytes(),
			types.HexToHash(order.Data).Bytes(),
		),
	)
}
func (*EthereumHydroProtocol) GetMatchOrderCallData(takerOrder *sdk.Order, makerOrders []*sdk.Order, baseTokenFilledAmounts []*big.Int) []byte {
	var buf bytes.Buffer

	//buf.Write([]byte{'\x8d', '\x10', '\x88', '\x3d'}) // function id v1.0
	buf.Write([]byte{'\x88', '\x4d', '\xad', '\x2e'}) // function id v1.1
	buf.Write(getLightOrderBytesFromOrder(takerOrder))

	// offset of makerOrders
	buf.Write(uint64ToPaddingBytes(uint64(13*32), 32))
	// offset of fillAmounts
	buf.Write(uint64ToPaddingBytes(uint64((14+len(makerOrders)*8)*32), 32))

	buf.Write(types.HexToHash(takerOrder.BaseTokenAddress).Bytes())
	buf.Write(types.HexToHash(takerOrder.QuoteTokenAddress).Bytes())

	relayerAdx := types.HexToAddress(strings.Trim(takerOrder.Relayer, "0x"))
	buf.Write(utils.LeftPadBytes(relayerAdx.Bytes(), 32))

	proxyAdx := types.HexToAddress(strings.Trim(config.Getenv("HSK_PROXY_ADDRESS"), "0x"))
	buf.Write(utils.LeftPadBytes(proxyAdx.Bytes(), 32))

	// makerCount
	buf.Write(uint64ToPaddingBytes(uint64(len(makerOrders)), 32))
	// makerLightOrders
	for _, makerOrder := range makerOrders {
		buf.Write(getLightOrderBytesFromOrder(makerOrder))
	}

	// baseTokenFilledAmount count
	buf.Write(uint64ToPaddingBytes(uint64(len(baseTokenFilledAmounts)), 32))
	// baseTokenFilledAmounts
	for _, baseTokenFilledAmount := range baseTokenFilledAmounts {
		buf.Write(types.BigToHash(baseTokenFilledAmount).Bytes())
	}

	return buf.Bytes()
}

func (*EthereumHydroProtocol) IsValidOrderSignature(address string, orderID string, signature string) bool {
	// ethereum signature config: [:32] r[32:64] s[64:]
	// first byte of config is v
	sigBytes := utils.Hex2Bytes(signature)

	if len(sigBytes) != 96 {
		panic(fmt.Errorf("order signature for ethereum should have 96 bytes. %s", signature))
	}

	ethSig := make([]byte, 65)
	copy(ethSig[:64], sigBytes[32:])
	ethSig[64] = sigBytes[0]

	res, _ := IsValidSignature(address, orderID, utils.Bytes2HexP(ethSig))

	return res
}

func getDomainSeparator() []byte {
	return crypto.Keccak256(
		EIP712_DOMAIN_TYPEHASH,
		crypto.Keccak256([]byte("Hydro Protocol")),
	)
}

func getEIP712MessageHash(message []byte) []byte {
	return crypto.Keccak256(
		[]byte{'\x19', '\x01'},
		getDomainSeparator(),
		message,
	)
}

func uint64ToPaddingBytes(num uint64, bytesLength int) []byte {
	numStr := strconv.FormatUint(num, 16)
	if len(numStr)&1 == 1 {
		numStr = fmt.Sprintf("0%s", numStr)
	}
	return utils.LeftPadBytes(utils.Hex2Bytes(numStr), bytesLength)
}

func GetOrderData(
	version uint64,
	isSell bool,
	isMarketOrder bool,
	expiredAt, rawMakerFeeRate, rawTakerFeeRate, rawMakerRebateRate uint64,
	salt uint64,
	isMakerOnly bool,
) string {
	var buf bytes.Buffer

	buf.WriteByte(uint64ToPaddingBytes(version, 1)[0])

	if isSell {
		buf.WriteByte('\x01')
	} else {
		buf.WriteByte('\x00')
	}

	if isMarketOrder {
		buf.WriteByte('\x01')
	} else {
		buf.WriteByte('\x00')
	}

	buf.Write(uint64ToPaddingBytes(expiredAt, 5))

	buf.Write(uint64ToPaddingBytes(rawMakerFeeRate, 2))
	buf.Write(uint64ToPaddingBytes(rawTakerFeeRate, 2))
	buf.Write(uint64ToPaddingBytes(rawMakerRebateRate, 2))
	buf.Write(uint64ToPaddingBytes(salt, 8))

	if isMakerOnly {
		buf.WriteByte('\x01')
	} else {
		buf.WriteByte('\x00')
	}

	rst := utils.Bytes2Hex(utils.RightPadBytes(buf.Bytes(), 32))

	return rst
}

func GetHash(order *sdk.Order) []byte {
	return getEIP712MessageHash(
		crypto.Keccak256(
			EIP712_ORDER_TYPE,
			types.HexToHash(order.Trader).Bytes(),
			types.HexToHash(order.Relayer).Bytes(),
			types.HexToHash(order.BaseTokenAddress).Bytes(),
			types.HexToHash(order.QuoteTokenAddress).Bytes(),
			types.BytesToHash(order.BaseTokenAmount.Bytes()).Bytes(),
			types.BytesToHash(order.QuoteTokenAmount.Bytes()).Bytes(),
			types.BytesToHash(order.GasTokenAmount.Bytes()).Bytes(),
			types.HexToHash(order.Data).Bytes(),
		),
	)
}

func GetRawMakerFeeRateFromOrderData(data string) uint16 {
	return binary.BigEndian.Uint16(types.HexToHash(data).Bytes()[8:10])
}
func GetRawTakerFeeRateFromOrderData(data string) uint16 {
	return binary.BigEndian.Uint16(types.HexToHash(data).Bytes()[10:12])
}
func GetRawMakerRebateRateFromOrderData(data string) uint16 {
	return binary.BigEndian.Uint16(types.HexToHash(data).Bytes()[12:14])
}
func GetIsMakerOnlyFromOrderData(data string) bool {
	return int(types.HexToHash(data).Bytes()[22]) >= 1
}

func GetOrderExpireTsFromOrderData(data string) uint64 {
	bytes := types.HexToHash(data).Bytes()[3:8]
	paddedBytes := utils.LeftPadBytes(bytes[:], 8)

	return binary.BigEndian.Uint64(paddedBytes)
}

func getLightOrderBytesFromOrder(order *sdk.Order) []byte {
	var buf bytes.Buffer

	buf.Write(types.HexToHash(order.Trader).Bytes())
	buf.Write(types.BytesToHash(order.BaseTokenAmount.Bytes()).Bytes())
	buf.Write(types.BytesToHash(order.QuoteTokenAmount.Bytes()).Bytes())
	buf.Write(types.BytesToHash(order.GasTokenAmount.Bytes()).Bytes())
	buf.Write(types.HexToHash(order.Data).Bytes())

	buf.Write(utils.Hex2Bytes(order.Signature))

	return buf.Bytes()
}

const FeeRateBase = 100000

func addTailingZero(data string, length int) string {
	return data + strings.Repeat("0", length-len(data))
}

func addLeadingZero(data string, length int) string {
	return strings.Repeat("0", length-len(data)) + data
}
