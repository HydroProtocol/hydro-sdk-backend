package utils

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const DDexApiUrl = "https://api.ddex.io/v3"

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient(nil)
	assert.NotNil(t, client)
}

func TestRequest(t *testing.T) {
	client := NewHttpClient(nil)
	marketStatusUrl := fmt.Sprintf("%s/%s", DDexApiUrl, "markets/status")
	err, code, x := client.Request(http.MethodGet, marketStatusUrl, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, code, http.StatusOK)
}

func TestGet(t *testing.T) {
	client := NewHttpClient(nil)
	marketStatusUrl := fmt.Sprintf("%s/%s", DDexApiUrl, "markets/status")
	err, code, _ := client.Get(marketStatusUrl, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, code, http.StatusOK)
}

func TestPost(t *testing.T) {
	client := NewHttpClient(nil)
	marketStatusUrl := fmt.Sprintf("%s/%s", DDexApiUrl, "orders")
	err, code, resp := client.Post(marketStatusUrl, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, code, http.StatusOK)

	var orderResp map[string]interface{}
	json.Unmarshal(resp, &orderResp)
	assert.EqualValues(t, -11, orderResp["status"])
	assert.EqualValues(t, "Authentication check failed. Please connect your wallet.", orderResp["desc"])
}

func TestDelete(t *testing.T) {
	client := NewHttpClient(nil)
	marketStatusUrl := fmt.Sprintf("%s/%s/%s", DDexApiUrl, "orders", "orderid")
	err, code, resp := client.Delete(marketStatusUrl, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, code, http.StatusOK)

	var orderResp map[string]interface{}
	json.Unmarshal(resp, &orderResp)
	assert.EqualValues(t, -11, orderResp["status"])
	assert.EqualValues(t, "Authentication check failed. Please connect your wallet.", orderResp["desc"])
}

func TestPut(t *testing.T) {
	client := NewHttpClient(nil)
	marketStatusUrl := fmt.Sprintf("%s/%s/%s", DDexApiUrl, "orders", "orderid")
	err, code, _ := client.Put(marketStatusUrl, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, code, http.StatusNotFound)
}
