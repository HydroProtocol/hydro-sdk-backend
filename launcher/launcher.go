package launcher

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
)

type Launcher struct {
	Ctx             context.Context `json:"ctx"`
	GasPriceDecider GasPriceDecider
	SignService     ISignService
	BlockChain      sdk.BlockChain
}

func NewLauncher(ctx context.Context, sign ISignService, hydro sdk.Hydro, gasPriceDecider GasPriceDecider) *Launcher {
	return &Launcher{
		Ctx:             ctx,
		SignService:     sign,
		BlockChain:      hydro,
		GasPriceDecider: gasPriceDecider,
	}
}

func (l *Launcher) add(launchLog *LaunchLog) {
	launchLog.GasPrice = decimal.NullDecimal{
		Decimal: l.GasPriceDecider.GasPriceInWei(),
		Valid:   true,
	}

	signedRawTransaction := l.SignService.Sign(launchLog)
	transactionHash, err := l.BlockChain.SendRawTransaction(signedRawTransaction)

	if err != nil {
		utils.Debugf("%+v", launchLog)
		utils.Infof("Send Tx failed, launchLog ID: %d, err: %+v", launchLog.ID, err)
		panic(err)
	}

	utils.Infof("Send Tx, launchLog ID: %d, hash: %s", launchLog.ID, transactionHash)

	launchLog.Status = common.STATUS_PENDING

	if err != nil {
		utils.Infof("Update Launch Log Failed, ID: %d, err: %s", launchLog.ID, err)
		panic(err)
	}

	l.SignService.AfterSign()
}
