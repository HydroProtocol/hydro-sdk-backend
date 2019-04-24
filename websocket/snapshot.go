package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"io/ioutil"
	"net/http"
)

type SnapshotFetcher interface {
	GetV2(marketID string) *common.SnapshotV2
}

type DefaultHttpSnapshotFetcher struct {
	ApiUrl string
}

func (f *DefaultHttpSnapshotFetcher) GetV2(marketID string) *common.SnapshotV2 {
	res, err := http.Get(fmt.Sprintf("%s/markets/%s/orderbook", f.ApiUrl, marketID))

	if err != nil {
		panic(err)
	}

	bts, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	var resStruct struct {
		Status int
		Data   struct {
			Orderbook *common.SnapshotV2 `json:"orderBook"`
		}
	}

	err = json.Unmarshal(bts, &resStruct)

	if err != nil {
		panic(err)
	}

	return resStruct.Data.Orderbook
}
