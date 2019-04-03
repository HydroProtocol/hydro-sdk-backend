package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/config"
	"io/ioutil"
	"net/http"
)

type SnapshotFetcher interface {
	GetV2(marketID string) *common.SnapshotV2
}

type DefaultSnapshotFetcher struct{}

func (*DefaultSnapshotFetcher) GetV2(marketID string) *common.SnapshotV2 {
	res, err := http.Get(fmt.Sprintf("%s/markets/%s/orderbook", config.Getenv("HSK_API_URL"), marketID))

	if err != nil {
		panic(err)
	}

	bts, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	var resStruct struct {
		Status int
		Data   *common.SnapshotV2
	}

	err = json.Unmarshal(bts, &resStruct)

	if err != nil {
		panic(err)
	}

	return resStruct.Data
}

var defaultSnapshotFetcher SnapshotFetcher = &DefaultSnapshotFetcher{}

func GetMarketOrderbookSnapshotV2(fetcher SnapshotFetcher, marketID string) *common.SnapshotV2 {
	if fetcher == nil {
		fetcher = defaultSnapshotFetcher
	}

	return fetcher.GetV2(marketID)
}
