package types

import (
	"math/big"
)

type Transaction struct {
	Nonce     uint64  `json:"nonce"`
	Value     big.Int `json:"value"`
	To        string  `json:"to"`
	Data      []byte  `json:"data"`
	GasPrice  big.Int `json:"gasPrice"`
	GasLimit  uint64  `json:"gasLimit"`
	Signature []byte  `json:"signature"`
}

func NewTransaction(nonce uint64, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return &Transaction{
		Nonce:    nonce,
		Value:    *amount,
		To:       to,
		GasPrice: *gasPrice,
		GasLimit: gasLimit,
		Data:     data,
	}
}
