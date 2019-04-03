package watcher

import (
	"context"
	"encoding/json"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/models"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"strconv"
	"time"
)

type Watcher struct {
	lastSyncedBlockNumber uint64
	Ctx                   context.Context

	KVClient    common.IKVStore
	QueueClient common.IQueue
	Hydro       sdk.Hydro
}

const SleepSeconds = 3

func (w *Watcher) Run() {
	w.initBlockNumber()

	for {
		select {
		case <-w.Ctx.Done():
			return
		default:
			currentBlockNumber, err := w.Hydro.GetBlockNumber()

			if err != nil {
				utils.Error("Watcher GetBlockNumber Failed, %v", err)
				w.Sleep()
				continue
			}

			utils.Debug("CurrentNumber: %d, lastSyncedNumber: %d", currentBlockNumber, w.lastSyncedBlockNumber)

			if currentBlockNumber <= w.lastSyncedBlockNumber {
				utils.Info("Watcher is Synchronized, sleep %s Seconds", SleepSeconds*time.Second)
				w.Sleep()
				continue
			}

			err = w.syncNextBlock()

			if err != nil {
				utils.Error("Watcher Sync Blokc Error %v", err)
				w.Sleep()
				continue
			}

			w.lastSyncedBlockNumber = w.lastSyncedBlockNumber + 1
			err = w.KVClient.Set(common.HYDRO_WATCHER_BLOCK_NUMBER_CACHE_KEY, strconv.FormatUint(w.lastSyncedBlockNumber, 10), 0)

			if err != nil {
				utils.Error("Watcher Save LastSyncedBlockNumber Error %v", err)
			}
		}
	}
}

// Sleep allows watcher to exit even thought it is sleeping
func (w *Watcher) Sleep() {
	select {
	case <-w.Ctx.Done():
	case <-time.After(SleepSeconds * time.Second):
	}
}

func (w *Watcher) initBlockNumber() {
	var blockNumber uint64

	val, err := w.KVClient.Get(common.HYDRO_WATCHER_BLOCK_NUMBER_CACHE_KEY)

	if err == common.KVStoreEmpty {
		blockNumber, _ = w.Hydro.GetBlockNumber()
		utils.Debug("Cache block number is nil, use current block number: %d", blockNumber)
	} else if err != nil {
		panic(err)
	} else {
		blockNumber, err = strconv.ParseUint(val, 0, 64)

		if err != nil {
			panic(err)
		}
	}

	w.lastSyncedBlockNumber = blockNumber
	return
}

func (w *Watcher) syncNextBlock() (err error) {
	utils.Debug("Sync Block %d", w.lastSyncedBlockNumber+1)

	block, err := w.Hydro.GetBlockByNumber(w.lastSyncedBlockNumber + 1)

	if err != nil {
		utils.Error("Sync Block %d Error, %+v", w.lastSyncedBlockNumber+1, err)
		return
	}

	txs := block.GetTransactions()

	for i := range txs {
		w.syncTransaction(txs[i], block.Timestamp())
	}

	return
}

func (w *Watcher) syncTransaction(tx sdk.Transaction, timestamp uint64) {
	launchLog := models.LaunchLogDao.FindByHash(tx.GetHash())

	if launchLog == nil {
		utils.Debug("Skip useless transaction %s", tx.GetHash())
		return
	}

	if launchLog.Status != common.STATUS_PENDING {
		utils.Info("LaunchLog is not pending %s, skip", launchLog.Hash.String)
		return
	}

	if launchLog != nil {
		txReceipt, _ := w.Hydro.GetTransactionReceipt(tx.GetHash())
		result := txReceipt.GetResult()
		hash := tx.GetHash()
		transaction := models.TransactionDao.FindTransactionByID(launchLog.ItemID)
		utils.Info("Transaction %s result is %+v", tx.GetHash(), result)
		//w.handleTransaction(launchLog.ItemID, result)

		var status string

		if result {
			status = common.STATUS_SUCCESSFUL
		} else {
			status = common.STATUS_FAILED
		}

		event := &common.ConfirmTransactionEvent{
			common.Event{
				common.EventConfirmTransaction,
				transaction.MarketID,
			},
			hash,
			status,
			timestamp,
		}

		bts, _ := json.Marshal(event)

		err := w.QueueClient.Push(bts)

		if err != nil {
			utils.Error("Push event into Queue Error %v", err)
		}
	}
}
