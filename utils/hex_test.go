package utils

import (
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"testing"
)

func TestInt2Bytes(t *testing.T) {
	assert.EqualValues(t, []byte{0x64}, Int2Bytes(100))
	assert.EqualValues(t, []byte{0x64, 0x0}, Int2Bytes(25600))

	assert.EqualValues(t, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Int2Bytes(math.MaxUint64))
	assert.EqualValues(t, []byte{0xff, 0xff, 0xff, 0xff}, Int2Bytes(math.MaxUint32))

	assert.EqualValues(t, []byte{1}, Int2Bytes(1))
	assert.EqualValues(t, []byte{254}, Int2Bytes(254))
	assert.EqualValues(t, []byte{0x1, 0}, Int2Bytes(256))
}

func TestInt2Hex(t *testing.T) {
	assert.EqualValues(t, "64", Int2Hex(100))
	assert.EqualValues(t, "6400", Int2Hex(25600))

	assert.EqualValues(t, "ffffffffffffffff", Int2Hex(math.MaxUint64))
	assert.EqualValues(t, "ffffffff", Int2Hex(math.MaxUint32))

}

func TestHex2Int(t *testing.T) {
	assert.EqualValues(t, 0, Hex2Int("0x0"))
	assert.EqualValues(t, 1, Hex2Int("0x01"))
	assert.EqualValues(t, 100, Hex2Int("64"))
	assert.EqualValues(t, 25600, Hex2Int("6400"))

	var maxUint64 uint64
	maxUint64 = math.MaxUint64
	assert.EqualValues(t, Hex2Int("ffffffffffffffff"), maxUint64)
	assert.EqualValues(t, Hex2Int("ffffffff"), math.MaxUint32)

	assert.EqualValues(t, 0, Hex2Int("-100"))
	assert.EqualValues(t, 0, Hex2Int("invalid number"))
}

func TestHex2Bytes(t *testing.T) {
	assert.EqualValues(t, []byte{0xff, 0xff}, Hex2Bytes("ffff"))
	assert.EqualValues(t, []byte{0x0f, 0xff}, Hex2Bytes("fff"))
	assert.EqualValues(t, []byte{0xff, 0xff}, Hex2Bytes("0xffff"))
}

func TestBytes2Hex(t *testing.T) {
	assert.EqualValues(t, "ffff", Bytes2Hex([]byte{0xff, 0xff}))
	assert.EqualValues(t, "ff0f", Bytes2Hex([]byte{0xff, 0xf}))
}

func TestBytes2HexP(t *testing.T) {
	assert.EqualValues(t, "0xff12", Bytes2HexP([]byte{0xff, 0x12}))
}

func TestBytes2BigInt(t *testing.T) {
	b := big.NewInt(0)
	b.SetString("ffffffffffffffff", 16)
	assert.EqualValues(t, b, Bytes2BigInt([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}))
}

func TestHex2BigInt(t *testing.T) {
	b := big.NewInt(0)
	b.SetString("ffffffffffffffff", 16)
	assert.EqualValues(t, b, Hex2BigInt("ffffffffffffffff"))
}
