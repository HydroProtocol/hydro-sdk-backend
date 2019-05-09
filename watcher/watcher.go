package watcher

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
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

	TransactionHandler *TransactionHandler
}

type TransactionHandler interface {
	Update(sdk.Transaction, uint64)
}

func (w *Watcher) RegisterHandler(handler TransactionHandler) {
	w.TransactionHandler = &handler
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
				utils.Errorf("Watcher GetBlockNumber Failed, %v", err)
				w.Sleep()
				continue
			}

			utils.Debugf("CurrentNumber: %d, lastSyncedNumber: %d", currentBlockNumber, w.lastSyncedBlockNumber)

			if currentBlockNumber <= w.lastSyncedBlockNumber {
				utils.Infof("Watcher is Synchronized, sleep %s Seconds", SleepSeconds*time.Second)
				w.Sleep()
				continue
			}

			err = w.syncNextBlock()

			if err != nil {
				utils.Errorf("Watcher Sync Blokc Error %v", err)
				w.Sleep()
				continue
			}

			w.lastSyncedBlockNumber = w.lastSyncedBlockNumber + 1
			err = w.KVClient.Set(common.HYDRO_WATCHER_BLOCK_NUMBER_CACHE_KEY, strconv.FormatUint(w.lastSyncedBlockNumber, 10), 0)

			if err != nil {
				utils.Errorf("Watcher Save LastSyncedBlockNumber Error %v", err)
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
		utils.Debugf("Cache block number is nil, use current block number: %d", blockNumber)
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
	utils.Debugf("Sync Block %d", w.lastSyncedBlockNumber+1)

	block, err := w.Hydro.GetBlockByNumber(w.lastSyncedBlockNumber + 1)

	if err != nil {
		utils.Errorf("Sync Block %d Error, %+v", w.lastSyncedBlockNumber+1, err)
		return
	}

	txs := block.GetTransactions()

	for i := range txs {
		w.syncTransaction(txs[i], block.Timestamp())
	}

	return
}

func (w *Watcher) syncTransaction(tx sdk.Transaction, timestamp uint64) {
	if w.TransactionHandler != nil {
		(*w.TransactionHandler).Update(tx, timestamp)
	}
}
