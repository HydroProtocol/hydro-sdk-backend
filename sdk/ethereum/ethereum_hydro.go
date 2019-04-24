package ethereum

import (
	"fmt"
	"os"
)

type EthereumHydro struct {
	*Ethereum
	*EthereumHydroProtocol
}

func NewEthereumHydro(rpcURL, hybridExAddr string) *EthereumHydro {
	if rpcURL == "" {
		rpcURL = os.Getenv("HSK_BLOCKCHAIN_RPC_URL")
	}

	if rpcURL == "" {
		panic(fmt.Errorf("NewEthereumHydro need argument rpcURL"))
	}

	return &EthereumHydro{
		NewEthereum(rpcURL, hybridExAddr),
		&EthereumHydroProtocol{},
	}
}
