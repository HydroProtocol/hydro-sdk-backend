package ethereum

import (
	"encoding/hex"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ethereumTestSuite struct {
	suite.Suite
	blockchain sdk.BlockChain
}

func (s *ethereumTestSuite) SetupSuite() {
	s.blockchain = NewEthereum("http://localhost:8545", "foo")
}

func (s *ethereumTestSuite) TearDownSuite() {
}

func (s *ethereumTestSuite) TearDownTest() {
}

func (s *ethereumTestSuite) TestGetTransactions() {
	_, err := s.blockchain.GetBlockNumber()
	s.Nil(err)
}

func (s *ethereumTestSuite) TestGetBlockByNumber() {
	blockNumber, err := s.blockchain.GetBlockNumber()
	s.Nil(err)

	_, err = s.blockchain.GetBlockByNumber(blockNumber)
	s.Nil(err)
}

func (s *ethereumTestSuite) TestGetTransaction() {
	blockNumber, err := s.blockchain.GetBlockNumber()
	s.Nil(err)

	block, err := s.blockchain.GetBlockByNumber(blockNumber)
	s.Nil(err)

	block.Number()
	block.Timestamp()

	transactions := block.GetTransactions()
	s.NotZero(len(transactions))

	hash := transactions[0].GetHash()

	tx, err := s.blockchain.GetTransaction(hash)
	s.Nil(err)
	s.Equal(hash, tx.GetHash())

	txReceipt, err := s.blockchain.GetTransactionReceipt(hash)
	s.Nil(err)
	s.Equal(blockNumber, txReceipt.GetBlockNumber())

	tx2, txReceipt2, err := s.blockchain.GetTransactionAndReceipt(hash)
	txReceipt2.GetResult()
	s.Nil(err)
	s.Equal(hash, tx2.GetHash())
	s.Equal(blockNumber, txReceipt2.GetBlockNumber())
}

func (s *ethereumTestSuite) TestAddTailingZero() {
	s.Equal("3000000000", addTailingZero("3", 10))
}

func (s *ethereumTestSuite) TestAddLeadingZero() {
	s.Equal("0000000003", addLeadingZero("3", 10))
}

func (s *ethereumTestSuite) TestGetTokenBalance() {
	balance := s.blockchain.GetTokenBalance("0x4C4Fa7E8EA4cFCfC93DEAE2c0Cff142a1DD3a218", "0x126aa4ef50a6e546aa5ecd1eb83c060fb780891a")
	s.Equal("100000000000000000000000", balance.String())
}

func (s *ethereumTestSuite) TestGetTokenAllowance() {
	allowance := s.blockchain.GetTokenAllowance("0x4C4Fa7E8EA4cFCfC93DEAE2c0Cff142a1DD3a218", "0x04f67E8b7C39A25e100847Cb167460D715215FEb", "0x126aa4ef50a6e546aa5ecd1eb83c060fb780891a")
	s.True(allowance.GreaterThanOrEqual(utils.StringToDecimal("0x0f00000000000000000000000000000000000000000000000000000000000000")))
}

func (s *ethereumTestSuite) TestSignAndVerify() {
	address := "0x126aa4ef50a6e546aa5ecd1eb83c060fb780891a"
	privakeKey := "a6553a3cbade744d6c6f63e557345402abd93e25cd1f1dba8bb0d374de2fcf4f"
	message := "🌞🌛👌😄💗"

	sigBytes, err := crypto.PersonalSign([]byte(message), privakeKey)
	s.Nil(err)

	match, err := IsValidSignature(address, message, "0x"+hex.EncodeToString(sigBytes))
	s.Nil(err)

	s.True(match)
}

func TestSignAndVerify1(t *testing.T) {
	privakeKey := "b7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c"
	message := "0xc44ba4a2d189d0eacd82c83f63dd5dd681b21800f317a3e94fe306db4fb91cf0"

	sigBytes, _ := crypto.PersonalSign(utils.Hex2Bytes(message[2:]), privakeKey)

	utils.Infof("0x" + hex.EncodeToString(sigBytes))
}

func TestEthereumSuite(t *testing.T) {
	suite.Run(t, new(ethereumTestSuite))
}
