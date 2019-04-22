package launcher

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/config"
	"github.com/HydroProtocol/hydro-sdk-backend/models"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
	"time"
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

const pollingIntervalSeconds = 5

func (l *Launcher) Run() {
	utils.Info("launcher start!")
	defer utils.Info("launcher stop!")

	for {
		launchLogs := models.LaunchLogDao.FindAllCreated()

		if len(launchLogs) == 0 {
			select {
			case <-l.Ctx.Done():
				utils.Info("main loop Exit")
				return
			default:
				utils.Info("no logs need to be sent. sleep %ds", pollingIntervalSeconds)
				time.Sleep(pollingIntervalSeconds * time.Second)
				continue
			}
		}

		for _, launchLog := range launchLogs {
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

			models.UpdateLaunchLogToPending(launchLog)

			if err != nil {
				utils.Info("Update Launch Log Failed, ID: %d, err: %s", launchLog.ID, err)
				panic(err)
			}

			l.SignService.AfterSign()
		}
	}
}

func Run(ctx context.Context) {
	// db
	models.ConnectDatabase("postgres", config.Getenv("HSK_DATABASE_URL"))

	// blockchain
	hydro := ethereum.NewEthereumHydro(config.Getenv("HSK_BLOCKCHAIN_RPC_URL"))
	signService := NewDefaultSignService(config.Getenv("HSK_RELAYER_PK"), hydro.GetTransactionCount)
	gasService := func() decimal.Decimal { return utils.StringToDecimal("3000000000") } // default 10 Gwei

	launcher := NewLauncher(ctx, signService, hydro, gasService)
	launcher.Run()
}
