package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func Int2Hex(number uint64) string {
	return fmt.Sprintf("%x", number)
}

// just return uint64 type
func Hex2Int(hex string) uint64 {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	intNumber, err := strconv.ParseUint(hex, 16, 64)

	if err != nil {
		return 0
	}

	return uint64(intNumber)
}

func Bytes2Hex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func Hex2Bytes(str string) []byte {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}

	if len(str)%2 == 1 {
		str = "0" + str
	}

	h, _ := hex.DecodeString(str)
	return h
}

// with prefix '0x'
func Bytes2HexP(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

func Hex2BigInt(str string) *big.Int {
	bytes := Hex2Bytes(str)
	b := big.NewInt(0)
	b.SetBytes(bytes)
	return b
}

func Bytes2BigInt(bytes []byte) *big.Int {
	b := big.NewInt(0)
	b.SetBytes(bytes)
	return b
}

// RightPadBytes zero-pads slice to the right up to length l.
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func Int2Bytes(i uint64) []byte {
	return Hex2Bytes(Int2Hex(i))
}
