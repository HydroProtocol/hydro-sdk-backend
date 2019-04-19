package ethereum

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	// "encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	// "math/big"
	"testing"
)

type erc20TestSuite struct {
	suite.Suite
	erc20 IErc20
}

const rpcURL = "http://127.0.0.1:8545"

func TestErc20(t *testing.T) {
	os.Setenv("HSK_BLOCKCHAIN_RPC_URL", rpcURL)
	erc20 := NewErc20Service(nil)
	fmt.Println(erc20.Name("0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218"))
}

func (s *erc20TestSuite) SetupSuite() {
	os.Setenv("HSK_BLOCKCHAIN_RPC_URL", rpcURL)
	s.erc20 = NewErc20Service(nil)
	httpmock.Activate()
}

func (s *erc20TestSuite) TearDownSuite() {
	httpmock.Deactivate()
}

func (s *erc20TestSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *erc20TestSuite) methodEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "method").String()

	s.Require().Equal(expected, value)
}

func (s *erc20TestSuite) paramsEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "params").Raw
	if expected == "null" {
		s.Require().Equal(expected, value)
	} else {
		s.JSONEq(expected, value)
	}
}

func (s *erc20TestSuite) getBody(request *http.Request) []byte {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	s.Require().Nil(err)

	return body
}

func (s *erc20TestSuite) TestTotalSupply() {
	httpmock.RegisterResponder("POST", rpcURL, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "eth_call")
		s.paramsEqual(body, `[{"data": "0x18160ddd", "from":"", "to":"0x9af839687f6c94542ac5ece2e317daae355493a1"},"latest"]`)

		// 1560000000000000000000000000 total
		response := `{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000050a66d97430c80d18000000"}`
		return httpmock.NewStringResponse(200, response), nil
	})

	_, totalSupply := s.erc20.TotalSupply("0x9af839687f6c94542ac5ece2e317daae355493a1")
	s.Require().Equal("1560000000000000000000000000", totalSupply.String())
}

func (s *erc20TestSuite) TestGetSymbol() {
	httpmock.RegisterResponder("POST", rpcURL, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "eth_call")
		s.paramsEqual(body, `[{"data": "0x95d89b41", "from":"", "to":"0x9af839687f6c94542ac5ece2e317daae355493a1"},"latest"]`)
		response := `{"jsonrpc":"2.0","id":1,"result":"0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000003484f540000000000000000000000000000000000000000000000000000000000"}`
		return httpmock.NewStringResponse(200, response), nil
	})

	_, symbol := s.erc20.Symbol("0x9af839687f6c94542ac5ece2e317daae355493a1")
	s.Require().Equal("HOT", symbol)
}

func (s *erc20TestSuite) TestGetName() {
	httpmock.RegisterResponder("POST", rpcURL, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "eth_call")
		s.paramsEqual(body, `[{"data": "0x06fdde03", "from":"", "to":"0x9af839687f6c94542ac5ece2e317daae355493a1"},"latest"]`)
		response := `{"jsonrpc":"2.0","id":1,"result":"0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000014487964726f2050726f746f636f6c20546f6b656e000000000000000000000000"}`
		return httpmock.NewStringResponse(200, response), nil
	})

	_, name := s.erc20.Name("0x9af839687f6c94542ac5ece2e317daae355493a1")
	s.Require().Equal("Hydro Protocol Token", name)
}

func (s *erc20TestSuite) TestGetDecimals() {
	httpmock.RegisterResponder("POST", rpcURL, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "eth_call")
		s.paramsEqual(body, `[{"data": "0x313ce567", "from":"", "to":"0x9af839687f6c94542ac5ece2e317daae355493a1"},"latest"]`)
		response := `{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000012"}`
		return httpmock.NewStringResponse(200, response), nil
	})

	_, decimals := s.erc20.Decimals("0x9af839687f6c94542ac5ece2e317daae355493a1")
	s.Require().Equal(18, decimals)
}

func TestTuncate(t *testing.T) {
	if "醉翁之" != tuncate("醉翁之意不在酒", 3) {
		t.Error("wrong")
	}
	if "醉翁之意不在酒" != tuncate("醉翁之意不在酒", 300) {
		t.Error("wrong")
	}
}

func TestNewErc20Service(t *testing.T) {
	suite.Run(t, new(erc20TestSuite))
}
