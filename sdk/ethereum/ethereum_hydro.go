package ethereum

type EthereumHydro struct {
	*Ethereum
	*EthereumHydroProtocol
}

func NewEthereumHydro(rpcURL string) *EthereumHydro {
	return &EthereumHydro{
		NewEthereum(rpcURL),
		&EthereumHydroProtocol{},
	}
}
