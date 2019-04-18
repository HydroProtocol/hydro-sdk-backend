package utils

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

func String2BigInt(str string) big.Int {
	n := new(big.Int)
	n.SetString(str, 0)
	return *n
}

// To Decimal
func StringToDecimal(str string) decimal.Decimal {
	if len(str) >= 2 && str[:2] == "0x" {
		b := new(big.Int)
		b.SetString(str[2:], 16)
		d := decimal.NewFromBigInt(b, 0)
		return d
	} else {
		v, err := decimal.NewFromString(str)
		if err != nil {
			panic(err)
		}
		return v
	}
}

func IntToDecimal(value interface{}) decimal.Decimal {
	ret, err := decimal.NewFromString(NumberToString(value))
	if err != nil {
		panic(fmt.Errorf("IntToDecimal error %+v", value))
	}

	return ret
}

func DecimalToBigInt(d decimal.Decimal) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(d.StringFixed(0), 10)
	if !ok {
		panic(fmt.Errorf("decimalToBigInt error %+v", d))
	}
	return n
}

func NumberToString(number interface{}) string {
	return fmt.Sprintf("%d", number)
}

func ParseInt(number string, defaultNumber int) int {
	ret, err := strconv.ParseInt(number, 10, 32)
	if err != nil {
		return int(defaultNumber)
	}

	return int(ret)
}

// IntToHex convert int to hexadecimal representation
func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}
