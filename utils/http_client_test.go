package utils

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const requestURL = "https://httpbin.org"

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient(nil)
	assert.NotNil(t, client)
}

func TestRequest(t *testing.T) {
	client := NewHttpClient(nil)
	err, code, _ := client.Request(http.MethodGet, requestURL, nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, code)
}

func TestGet(t *testing.T) {
	client := NewHttpClient(nil)
	err, code, _ := client.Get(requestURL+"/get", nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, code)
}

func TestPost(t *testing.T) {
	client := NewHttpClient(nil)
	err, code, _ := client.Post(requestURL+"/post", nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, code)
}

func TestDelete(t *testing.T) {
	client := NewHttpClient(nil)
	err, code, _ := client.Delete(requestURL+"/delete", nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, code)
}

func TestPut(t *testing.T) {
	client := NewHttpClient(nil)
	err, code, _ := client.Put(requestURL+"/put", nil, nil, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, code)
}
