package launcher

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

type GasPriceDecider interface {
	GasPriceInWei() decimal.Decimal
}

type StaticGasPriceDecider struct {
	PriceInWei decimal.Decimal
}

func (s StaticGasPriceDecider) GasPriceInWei() decimal.Decimal {
	return s.PriceInWei
}

type GasStationPriceDeciderWithFallback struct {
	FallbackGasPriceInWei decimal.Decimal
}

func (s GasStationPriceDeciderWithFallback) GasPriceInWei() decimal.Decimal {
	url := "https://ethgasstation.info/json/ethgasAPI.json"
	resp, err := http.Get(url)
	if err != nil {
		return s.FallbackGasPriceInWei
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s.FallbackGasPriceInWei
	}

	gasStationResp := GasStationRespBody{}
	err = json.Unmarshal(body, &gasStationResp)
	if err != nil || gasStationResp.Fast.IsZero() {
		return s.FallbackGasPriceInWei
	}

	// returned value from gasStation api is in 0.1Gwei
	gwei := decimal.New(1, 9)
	return gasStationResp.Fast.Div(decimal.NewFromFloat(10)).Mul(gwei)
}

type GasStationRespBody struct {
	Fast    decimal.Decimal `json:"fast"`
	Average decimal.Decimal `json:"average"`
}

func NewStaticGasPriceDecider(gasPrice decimal.Decimal) GasPriceDecider {
	return StaticGasPriceDecider{
		PriceInWei: gasPrice,
	}
}

func NewGasStationGasPriceDecider(fallbackGasPrice decimal.Decimal) GasPriceDecider {
	return GasStationPriceDeciderWithFallback{
		FallbackGasPriceInWei: fallbackGasPrice,
	}
}
