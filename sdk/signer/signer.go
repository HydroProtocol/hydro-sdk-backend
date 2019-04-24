package signer

import (
	"crypto/ecdsa"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/rlp"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
)

// HomesteadHash returns the hash of an unsigned transaction
func HomesteadHash(t *types.Transaction) []byte {
	rlpTx := rlp.Encode([]interface{}{
		rlp.EncodeUint64ToBytes(t.Nonce),
		t.GasPrice.Bytes(),
		rlp.EncodeUint64ToBytes(t.GasLimit),
		utils.Hex2Bytes(t.To[2:]),
		t.Value.Bytes(),
		t.Data,
	})
	hash := crypto.Keccak256(rlpTx)
	return hash
}

// EncodeRlp returns the rlp encoded content of a signed transaction
func EncodeRlp(t *types.Transaction) []byte {
	return rlp.Encode([]interface{}{
		rlp.EncodeUint64ToBytes(t.Nonce),
		t.GasPrice.Bytes(),
		rlp.EncodeUint64ToBytes(t.GasLimit),
		utils.Hex2Bytes(t.To[2:]),
		t.Value.Bytes(),
		t.Data,
		t.Signature[64:],
		t.Signature[0:32],
		t.Signature[32:64],
	})
}

// Hash returns the hash of a signed transaction
func Hash(t *types.Transaction) []byte {
	rlpTx := EncodeRlp(t)
	hash := crypto.Keccak256(rlpTx)
	return hash
}

func SignTx(transaction *types.Transaction, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	// We use homesteadHash to get best compatibility
	hash := HomesteadHash(transaction)

	sig, err := crypto.Sign(hash, key)

	// Since we are using HomesteadHash, the v is either 27 or 28
	// Mode details about EIP155 goes https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	if sig[64] < 27 {
		sig[64] = sig[64] + 27
	}

	transaction.Signature = sig
	return transaction, err
}
