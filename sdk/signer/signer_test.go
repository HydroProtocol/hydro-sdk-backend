package signer

import (
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSignTx(t *testing.T) {
	privateKey := "b7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c"

	nonce := uint64(1)
	to := "0x93388b4efe13b9b18ed480783c05462409851547"
	amount := big.NewInt(10)
	gasLimit := uint64(2)
	gasPrice := big.NewInt(20)
	data := []byte("hello")
	//pk1, _ := crypto.HexToECDSA(privateKey)
	//
	//transaction1 := types.NewTransaction(
	//	nonce,
	//	common.HexToAddress(to),
	//	amount,
	//	gasLimit,
	//	gasPrice,
	//	data,
	//)
	//signedTransaction1, _ := types.SignTx(transaction1, types.HomesteadSigner{}, pk1)
	//fmt.Println(signedTransaction1.Hash().String())

	except := utils.Hex2Bytes("0xc7dfbe726632d8d4edcaf1a9f59eb8976e06fd55989c82c79b9654a93c703f01")
	pk2, _ := crypto.NewPrivateKeyByHex(privateKey)
	transaction2 := types.NewTransaction(
		nonce,
		to,
		amount,
		gasLimit,
		gasPrice,
		data,
	)
	signedTransaction2, _ := SignTx(transaction2, pk2)
	assert.EqualValues(t, except, Hash(signedTransaction2))
}

func TestRlpEncode(t *testing.T) {
	privateKey := "b7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c"

	nonce := uint64(1)
	to := "0x93388b4efe13b9b18ed480783c05462409851547"
	amount := big.NewInt(10)
	gasLimit := uint64(2)
	gasPrice := big.NewInt(20)
	data := []byte("hello")

	//pk1, _ := crypto.HexToECDSA(privateKey)
	//transaction1 := types.NewTransaction(
	//	nonce,
	//	common.HexToAddress(to),
	//	amount,
	//	gasLimit,
	//	gasPrice,
	//	data,
	//)
	//signedTransaction1, _ := types.SignTx(transaction1, types.HomesteadSigner{}, pk1)
	//buf := new(bytes.Buffer)
	//signedTransaction1.EncodeRLP(buf)
	//fmt.Println(utils.Bytes2HexP(buf.Bytes()))
	except := utils.Hex2Bytes("0xf8620114029493388b4efe13b9b18ed480783c054624098515470a8568656c6c6f1ba0218482bc5f636f0c0ac6bce6b24c2d12fbd8a96653fb6dfdc35d1cfbf6d8218da07148e667e2cb66568dcd4f777c1067afaaf4a8e55a5f604aba5172df54dd7d30")

	pk2, _ := crypto.NewPrivateKeyByHex(privateKey)
	transaction2 := types.NewTransaction(
		nonce,
		to,
		amount,
		gasLimit,
		gasPrice,
		data,
	)
	signedTransaction2, _ := SignTx(transaction2, pk2)
	assert.EqualValues(t, except, EncodeRlp(signedTransaction2))
}
