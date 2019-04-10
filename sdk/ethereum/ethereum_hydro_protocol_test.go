package ethereum

import (
	"math/big"
	"strings"
	"testing"

	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/test"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type hydroTestSuite struct {
	suite.Suite
}

func (suite *hydroTestSuite) SetupSuite() {
}

func (suite *hydroTestSuite) TearDownSuite() {
}

func (suite *hydroTestSuite) TearDownTest() {
}

func (suite *hydroTestSuite) TestDomainHash() {
	suite.Equal(
		"0xb2178a58fb1eefb359ecfdd57bb19c0bdd0f4e6eed8547f46600e500ed111af3",
		utils.Bytes2HexP(EIP712_DOMAIN_TYPEHASH),
	)
}

func (suite *hydroTestSuite) TestDomainSeparator() {
	suite.Equal(
		"0x097976fcea7606c3ff7a3beb3e4d47c93030165478ea6a99683bb493608d36bc",
		utils.Bytes2HexP(getDomainSeparator()),
	)
}

func (suite *hydroTestSuite) TestGetEIP712MessageHash() {
	message := utils.Hex2Bytes("ea83cdcdd06bf61e414054115a551e23133711d0507dcbc07a4bab7dc4581935")
	suite.Equal(
		"0xf77f07f3ec21820e65cf13028dc5deaa9fbab93e3cba2bc7acdab12813004459",
		utils.Bytes2HexP(getEIP712MessageHash(message)),
	)
}

func (suite *hydroTestSuite) TestIsValidSignature() {
	taker := "0x3870b6f2c0b723f4855d8ad53ab7599b02d4df84"
	orderHash := "0x4e418856ca6935c955df6afd9cb780f84be72e3256ade452811d8ebbc8ea42e1"

	sig := append(
		utils.Hex2Bytes("19cef14892021d56d31b8f6ca6ed99ab89ac918a2d8e2e9d034b14ccf1dfa17f"),
		utils.Hex2Bytes("27601ed6b0d1a7fd64f6e62c2fea27580f36521d2609baaf2f9921ecd6cb761b")...)

	sig = append(sig, (byte)(int('\x1c')-27))

	pub, err := crypto.SigToPub(utils.Hex2Bytes(orderHash), sig)
	if err != nil {
		panic(err)
	}

	realAdx := strings.ToLower(crypto.PubKey2Address(*pub))

	suite.Equal(taker, realAdx)
}

func (suite *hydroTestSuite) TestIsValidOrderSignature() {
	taker := "0xe269e891a2ec8585a378882ffa531141205e92e9"
	orderHash := "0x4e418856ca6935c955df6afd9cb780f84be72e3256ade452811d8ebbc8ea42e1"

	signature := "0x1b00000000000000000000000000000000000000000000000000000000000000" +
		"3966c24be7df61e273ae732f825140ffbe157c10a9ff6e2665e671b92d64d01a" +
		"5e8f0e9bb00962812ab7abca3467ab47c72a980c7acc18280f34496464176c70"

	suite.Equal(true, new(EthereumHydroProtocol).IsValidOrderSignature(taker, orderHash, signature))
}

var version = uint64(2)

// data component:
// ╔════════════════════╤═══════════════════════════════════════════════════════════╗
// ║                    │ length(bytes)   desc                                      ║
// ╟────────────────────┼───────────────────────────────────────────────────────────╢
// ║ version            │ 1               order version                             ║
// ║ side               │ 1               0: buy, 1: sell                           ║
// ║ isMarketOrder      │ 1               0: limitOrder, 1: marketOrder             ║
// ║ expiredAt          │ 5               second of this order expiration time      ║
// ║ asMakerFeeRate     │ 2               maker fee rate base 100,000               ║
// ║ asTakerFeeRate     │ 2               taker fee rate base 100,000               ║
// ║ makerRebateRate    │ 2               rebate rate for maker base 1,000,000      ║
// ║ salt               │ 8               salt                                      ║
// ║ isMakerOnly        │ 1               0: not isMakerOnly, 1: isMakerOnly        ║
// ║                    │ 9               reserved                                  ║
// ╚════════════════════╧═══════════════════════════════════════════════════════════╝
// bytes32 data;

func (suite hydroTestSuite) TestGetOrderData() {
	res := GetOrderData(version, true, true, 1539247438, 10000, 50000, 10000, 488701836, false)

	// 1539247438    hex             00 5b bf 0d 4e
	// 10000         hex                      27 10
	// 50000         hex                      c3 50
	// 488701836     hex    00 00 00 00 1d 20 ff 8c
	suite.Equal("020101005bbf0d4e2710c3502710000000001d20ff8c00000000000000000000", res)

	res = GetOrderData(version, true, true, 0, 0, 1, 1, 1, false)
	suite.Equal("0201010000000000000000010001000000000000000100000000000000000000", res)
}

func (suite *hydroTestSuite) TestGetHash() {
	taker := "0x3870b6f2c0b723f4855d8ad53ab7599b02d4df84"
	relayer := "0xd4a1963e645244c7fb4fe8efab12e4bc02c5fad3"

	baseCurrency := "0xfe1e07852eb0fa0df66843e84a41da212b455e98"
	quoteCurrency := "0x9712e6cadf82d1902088ef858502ca17261bb893"

	takerOrder := NewOrder(
		taker,
		relayer,
		baseCurrency,
		quoteCurrency,
		utils.Hex2BigInt("0x1bc16d674ec80000"),
		big.NewInt(0),
		utils.Hex2BigInt("0x572255eb17edfc4"),
		true,
		true,
		version,
		9999999999,
		100,
		200,
		100,
		520496,
		"0x"+"1c01000000000000000000000000000000000000000000000000000000000000"+"19cef14892021d56d31b8f6ca6ed99ab89ac918a2d8e2e9d034b14ccf1dfa17f"+"27601ed6b0d1a7fd64f6e62c2fea27580f36521d2609baaf2f9921ecd6cb761b",
	)

	computedOrderHash := GetHash(takerOrder)
	suite.Equal("6ae837ed30cba8174b589c644e351166c5e7dfa1ffdbd1a1333287daf63c9b43", utils.Bytes2Hex(computedOrderHash))
}

func (suite *hydroTestSuite) TestGetMatchOrdersDataHex() {
	test.PreTest()

	taker := "0x3870b6f2c0b723f4855d8ad53ab7599b02d4df84"
	maker := "0x85cf54dd216997bcf324c72aa1c845be2f059299"
	relayer := "0x93388b4efe13b9b18ed480783c05462409851547"

	baseCurrency := "0xfe1e07852eb0fa0df66843e84a41da212b455e98"
	quoteCurrency := "0x9712e6cadf82d1902088ef858502ca17261bb893"

	takerOrder := NewOrder(
		taker,
		relayer,
		baseCurrency,
		quoteCurrency,
		utils.Hex2BigInt("0x1bc16d674ec80000"),
		big.NewInt(0),
		utils.Hex2BigInt("0x572255eb17edfc4"),
		true,
		true,
		version,
		9999999999,
		100,
		200,
		100,
		520496,
		"0x"+
			"1c01000000000000000000000000000000000000000000000000000000000000"+
			"19cef14892021d56d31b8f6ca6ed99ab89ac918a2d8e2e9d034b14ccf1dfa17f"+
			"27601ed6b0d1a7fd64f6e62c2fea27580f36521d2609baaf2f9921ecd6cb761b",
	)

	makerOrders := []*sdk.Order{
		NewOrder(
			maker,
			relayer,
			baseCurrency,
			quoteCurrency,
			utils.Hex2BigInt("0xde0b6b3a7640000"),
			utils.Hex2BigInt("0xc7d713b49da0000"),
			utils.Hex2BigInt("0x572255eb17edfc4"),
			false,
			false,
			version,
			9999999999,
			100,
			200,
			100,
			183426,
			"0x"+
				"1b01000000000000000000000000000000000000000000000000000000000000"+
				"a096a43f547bd361d79f965723604965cf45fb9e67754c67accb100eef344804"+
				"571c23cc135c501eb72d137eeb6f430bb3a62da2122610223e3d80b99cecdff9",
		),
		NewOrder(
			maker,
			relayer,
			baseCurrency,
			quoteCurrency,
			utils.Hex2BigInt("0xde0b6b3a7640000"),
			utils.Hex2BigInt("0xb1a2bc2ec500000"),
			utils.Hex2BigInt("0x572255eb17edfc4"),
			false,
			false,
			version,
			9999999999,
			100,
			200,
			100,
			821893,
			"0x"+
				"1b01000000000000000000000000000000000000000000000000000000000000"+
				"583d5bece6c5b0ef5b1807f681270c864e987ffeed66578d39343ab6996851c5"+
				"371013591a920621cd32c8bfd3f7d89d8aef36dfed7c636e08a0b0871491e765",
		),
	}

	expectedResult := "0x884dad2e0000000000000000000000003870b6f2c0b723f4855d8ad53ab7599b02d4df840000000000000000000000000000000000000000000000001bc16d674ec8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000572255eb17edfc402010102540be3ff006400c80064000000000007f130000000000000000000001c0100000000000000000000000000000000000000000000000000000000000019cef14892021d56d31b8f6ca6ed99ab89ac918a2d8e2e9d034b14ccf1dfa17f27601ed6b0d1a7fd64f6e62c2fea27580f36521d2609baaf2f9921ecd6cb761b00000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000003c0000000000000000000000000fe1e07852eb0fa0df66843e84a41da212b455e980000000000000000000000009712e6cadf82d1902088ef858502ca17261bb89300000000000000000000000093388b4efe13b9b18ed480783c0546240985154700000000000000000000000004f67e8b7c39a25e100847cb167460d715215feb000000000000000000000000000000000000000000000000000000000000000200000000000000000000000085cf54dd216997bcf324c72aa1c845be2f0592990000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000c7d713b49da00000000000000000000000000000000000000000000000000000572255eb17edfc402000002540be3ff006400c80064000000000002cc82000000000000000000001b01000000000000000000000000000000000000000000000000000000000000a096a43f547bd361d79f965723604965cf45fb9e67754c67accb100eef344804571c23cc135c501eb72d137eeb6f430bb3a62da2122610223e3d80b99cecdff900000000000000000000000085cf54dd216997bcf324c72aa1c845be2f0592990000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000b1a2bc2ec5000000000000000000000000000000000000000000000000000000572255eb17edfc402000002540be3ff006400c8006400000000000c8a85000000000000000000001b01000000000000000000000000000000000000000000000000000000000000583d5bece6c5b0ef5b1807f681270c864e987ffeed66578d39343ab6996851c5371013591a920621cd32c8bfd3f7d89d8aef36dfed7c636e08a0b0871491e76500000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000de0b6b3a7640000"

	var baseTokenFilledAmt []*big.Int
	baseTokenFilledAmt = append(baseTokenFilledAmt, utils.DecimalToBigInt(decimal.New(1, 18)))
	baseTokenFilledAmt = append(baseTokenFilledAmt, utils.DecimalToBigInt(decimal.New(1, 18)))

	ep := &EthereumHydroProtocol{}
	res := utils.Bytes2HexP(ep.GetMatchOrderCallData(takerOrder, makerOrders, baseTokenFilledAmt))
	suite.Equal(expectedResult, res)
}

func (suite *hydroTestSuite) TestGetAsTakerFeeRateFromOrderData() {
	data := "01010102540be3ff006400c8006400000000000df8f400000000000000000000"
	takerFee := GetRawTakerFeeRateFromOrderData(data)
	makerFee := GetRawMakerFeeRateFromOrderData(data)
	rebate := GetRawMakerRebateRateFromOrderData(data)

	suite.Equal(uint16(200), takerFee)
	suite.Equal(uint16(100), makerFee)
	suite.Equal(uint16(100), rebate)
}

func (suite *hydroTestSuite) TestGetAsTakerFeeRateFromOrderData2() {
	data := "01010102540be3ff006400c8006400000000000df8f400000000000000000000"

	asTakerFeeRate := decimal.New(int64(GetRawTakerFeeRateFromOrderData(data)), 0)

	suite.True(decimal.NewFromFloat(200).Equal(asTakerFeeRate))

	considerFee := asTakerFeeRate.Div(decimal.New(1, 5)).Add(decimal.New(1, 0))
	suite.True(considerFee.Equal(decimal.NewFromFloat(1.002)))
}

func (suite *hydroTestSuite) TestA() {
	config := "0x1234"

	var configBytes [32]byte
	copy(configBytes[:], utils.RightPadBytes(utils.Hex2Bytes(config[2:]), 32))
}

func (suite *hydroTestSuite) TestExpireTs() {
	data := "0x01010002540be3ff006400c8006400000000000d119600000000000000000000"

	ts := GetOrderExpireTsFromOrderData(data)

	suite.Equal(ts, uint64(9999999999))
}

func NewOrder(
	trader, relayer, baseCurrency, quoteCurrency string,
	baseCurrencyHugeAmount, quoteCurrencyHugeAmount, gasTokenHugeAmount *big.Int,
	isSell, isMarketOrder bool,
	version, expiredAt, rawMakerFeeRate, rawTakerFeeRate, rawMakerRebateRate, salt uint64,
	signature string,
) *sdk.Order {
	return &sdk.Order{
		Trader:            trader,
		Relayer:           relayer,
		BaseTokenAmount:   baseCurrencyHugeAmount,
		QuoteTokenAmount:  quoteCurrencyHugeAmount,
		BaseTokenAddress:  baseCurrency,
		QuoteTokenAddress: quoteCurrency,
		GasTokenAmount:    gasTokenHugeAmount,
		Data:              GetOrderData(version, isSell, isMarketOrder, expiredAt, rawMakerFeeRate, rawTakerFeeRate, rawMakerRebateRate, salt, false),
		Signature:         signature,
	}
}

func TestHydroTestSuite(t *testing.T) {
	suite.Run(t, new(hydroTestSuite))
}
