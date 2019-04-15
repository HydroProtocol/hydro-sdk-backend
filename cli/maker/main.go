package main

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
	"sync"
	"time"
)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
)

func randomNumber(min, max, decimals float64) float64 {
	r := rand.Float64()*(max-min) + min
	pow := math.Pow(10, float64(decimals))
	return math.Floor(r*pow) / pow
}

var apiURL = os.Getenv("HSK_API_URL")

// ethereum-test-node
// maker pk and address
// https://github.com/HydroProtocol/ethereum-test-node
const pk = "0xa6553a3cbade744d6c6f63e557345402abd93e25cd1f1dba8bb0d374de2fcf4f"
const address = "0x126aa4ef50a6e546aa5ecd1eb83c060fb780891a"

// First augur market has two options
const longMarket = "1-long"   // This is the id of long market
const shortMarket = "1-short" // This is the id of short market
const longMarketPrice = 0.6   // Let's assume the current market price is 0.6

type MarketMaking struct {
	LongMarketID    string
	ShortMarketID   string
	LongMarketPrice float64
}

var markets = []*MarketMaking{
	{
		"1-long",
		"1-short",
		0.6,
	},
	{
		"2-long",
		"2-short",
		0.87,
	},
}

func getHydroAuthenticationHeader() string {
	message := "HYDRO-AUTHENTICATION"
	signature, _ := ethereum.PersonalSign([]byte(message), pk)
	return fmt.Sprintf("%s#%s#%s", address, message, utils.Bytes2HexP(signature))
}

func setReqHeader(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Hydro-Authentication", getHydroAuthenticationHeader())
}

func cancelOrder(orderID string) {
	cancelOrderPayload, _ := json.Marshal(map[string]interface{}{
		"id": orderID,
	})

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/orders/%s", apiURL, orderID), bytes.NewReader(cancelOrderPayload))
	setReqHeader(req)
	_, err := http.DefaultClient.Do(req)

	if err != nil {
		utils.Error("cancel order req error: %v", err)
	}

	utils.Info("cancel order success %s", orderID)
}

func placeOrder(price, amount float64, side string, marketID string) string {
	body, _ := json.Marshal(map[string]interface{}{
		"amount":      fmt.Sprintf("%f", amount),
		"price":       fmt.Sprintf("%f", price),
		"side":        side,
		"orderType":   "limit",
		"marketID":    marketID,
		"isMakerOnly": false,
	})

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/orders/build", apiURL), bytes.NewReader(body))
	setReqHeader(req)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		utils.Error("build order req error: %v", err)
	}
	utils.Info("build order success %s", body)

	resBytes, _ := ioutil.ReadAll(res.Body)

	var buildOrderRes struct {
		Status int `json:"status"`
		Data   struct {
			Order struct {
				ID string `json:"id"`
			} `json:"order"`
		} `json:"data"`
	}

	_ = json.Unmarshal(resBytes, &buildOrderRes)

	orderID := buildOrderRes.Data.Order.ID

	signature, _ := ethereum.PersonalSign(
		utils.Hex2Bytes(orderID),
		pk,
	)

	placeOrderRequestBody, _ := json.Marshal(map[string]interface{}{
		"orderID":   orderID,
		"signature": utils.Bytes2HexP(toOrderSignature(signature)),
		"method":    0,
	})

	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/orders", apiURL), bytes.NewReader(placeOrderRequestBody))
	setReqHeader(req)
	res, err = http.DefaultClient.Do(req)

	if err != nil {
		utils.Error("place order req error: %v", err)
	}

	utils.Info("place order success %s", orderID)

	return orderID
}

func toOrderSignature(sign []byte) []byte {
	var res [96]byte
	copy(res[:], []byte{sign[64] + 27})
	copy(res[32:], sign[:64])
	return res[:]
}

func popID(ids []string, length int) ([]string, string) {
	id := ids[0]
	copy(ids, ids[1:length+1])
	ids = ids[:length]
	return ids, id
}

func run(market *MarketMaking) {
	var price, amount float64
	var id string
	const maxOrdersCount = 20

	buyIDs := make([]string, 0, maxOrdersCount+10)
	sellIDs := make([]string, 0, maxOrdersCount+10)

	for {
		// place buy order
		price = randomNumber(0, market.LongMarketPrice-0.01, 2)
		amount = randomNumber(1, 5, 4)
		id = placeOrder(price, amount, "buy", market.LongMarketID)
		buyIDs = append(buyIDs, id)

		// place sell order
		price = randomNumber(market.LongMarketPrice+0.01, 1, 2)
		amount = randomNumber(1, 5, 4)
		id = placeOrder(price, amount, "sell", market.LongMarketID)
		sellIDs = append(sellIDs, id)

		for len(buyIDs) > maxOrdersCount {
			buyIDs, id = popID(buyIDs, maxOrdersCount)
			cancelOrder(id)
		}

		for len(sellIDs) > maxOrdersCount {
			sellIDs, id = popID(sellIDs, maxOrdersCount)
			cancelOrder(id)
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	wg := sync.WaitGroup{}

	// run two makers and mirror for augur1 and augur2 long markets
	for i := range markets {
		wg.Add(1)

		go func(market *MarketMaking) {
			defer wg.Done()
			run(market)
		}(markets[i])
	}

	wg.Wait()
}
