// rlp encoding standard
// https://github.com/ethereum/wiki/wiki/RLP
package rlp

import (
	"bytes"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
)

func Encode(items interface{}) []byte {
	switch v := items.(type) {
	case []byte:
		if len(v) == 1 && v[0] < 0x80 {
			return v
		} else {
			return append(encodeLength(len(v), 0x80), v...)
		}
	case []interface{}:
		res := make([]byte, 0, 128)

		for i := range v {
			res = append(res, Encode(v[i])...)
		}

		return append(encodeLength(len(res), 0xc0), res...)
	}

	return nil
}

func encodeLength(L int, offset int) []byte {
	var bts bytes.Buffer
	if L < 56 {
		bts.WriteByte(byte(L + offset))
		return bts.Bytes()
	} else if L < 1<<31-1 {
		// In the rlp wiki page, the upper bound of L is 256**8
		// There is no need to support such a big size in hydro, we use maxInt as a limit here.
		BL := utils.Int2Bytes(uint64(L))
		bts.WriteByte(byte(len(BL) + int(offset) + 55))
		bts.Write(BL)
		return bts.Bytes()
	} else {
		panic("input length out of range")
	}
}

func EncodeUint64ToBytes(n uint64) []byte {
	if n == 0 {
		return []byte{}
	}

	return utils.Int2Bytes(n)
}
