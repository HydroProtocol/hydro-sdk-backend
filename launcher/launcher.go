package launcher

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
)

type Launcher struct {
	Ctx         context.Context `json:"ctx"`
	GasPrice    func() decimal.Decimal
	SignService ISignService
	BlockChain  sdk.BlockChain
}

func NewLauncher(ctx context.Context, sign ISignService, hydro sdk.Hydro, gasPrice func() decimal.Decimal) *Launcher {
	return &Launcher{
		Ctx:         ctx,
		SignService: sign,
		BlockChain:  hydro,
		GasPrice:    gasPrice,
	}
}

func (l *Launcher) add(launchLog *LaunchLog) {
	launchLog.GasPrice = decimal.NullDecimal{
		Decimal: l.GasPrice(),
		Valid:   true,
	}

	signedRawTransaction := l.SignService.Sign(launchLog)
	transactionHash, err := l.BlockChain.SendRawTransaction(signedRawTransaction)

	if err != nil {
		utils.Debug("%+v", launchLog)
		utils.Info("Send Tx failed, launchLog ID: %d, err: %+v", launchLog.ID, err)
		panic(err)
	}

	utils.Info("Send Tx, launchLog ID: %d, hash: %s", launchLog.ID, transactionHash)

	launchLog.Status = common.STATUS_PENDING

	if err != nil {
		utils.Info("Update Launch Log Failed, ID: %d, err: %s", launchLog.ID, err)
		panic(err)
	}

	l.SignService.AfterSign()
}
