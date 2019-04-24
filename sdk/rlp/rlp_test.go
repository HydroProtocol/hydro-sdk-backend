package rlp

import (
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

type rlpTestSuite struct {
	suite.Suite
}

func (s *rlpTestSuite) TestString() {
	s.Equal([]byte{0x83, 'd', 'o', 'g'}, Encode([]byte("dog")))

	longStringBytes := []byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit")
	s.Equal(append([]byte{0xb8, 0x38}, longStringBytes...), Encode(longStringBytes))

}

func (s *rlpTestSuite) TestStringList() {
	s.Equal([]byte{0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g'}, Encode([]interface{}{
		[]byte("cat"),
		[]byte("dog"),
	}))
}

func (s *rlpTestSuite) TestEmptyString() {
	s.Equal([]byte{0x80}, Encode([]byte("")))
}

func (s *rlpTestSuite) TestUint64Zero() {
	s.Equal([]byte{0x80}, Encode(EncodeUint64ToBytes(0)))
}

func (s *rlpTestSuite) TestEmptyList() {
	s.Equal([]byte{0xc0}, Encode([]interface{}{}))
}

func (s *rlpTestSuite) TestZero() {
	s.Equal([]byte{0x00}, Encode([]byte{0}))
}

func (s *rlpTestSuite) TestNumber() {
	s.Equal([]byte{0x0f}, Encode([]byte{15}))
	s.Equal([]byte{0x82, 0x04, 0x00}, Encode([]byte{0x04, 0x00}))
}

func (s *rlpTestSuite) TestNestedList() {
	s.Equal([]byte{0xc7, 0xc0, 0xc1, 0xc0, 0xc3, 0xc0, 0xc1, 0xc0}, Encode([]interface{}{
		[]interface{}{},
		[]interface{}{
			[]interface{}{},
		},
		[]interface{}{
			[]interface{}{},
			[]interface{}{
				[]interface{}{},
			},
		},
	}))
}

func (s *rlpTestSuite) TestOldEncodeCase() {
	data1 := []byte{0x1}
	data2 := []byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}

	s.Equal("0xd20190ffffffffffffffffffffffffffffffff", utils.Bytes2HexP(Encode([]interface{}{data1, data2})))
}

func TestRlpSuite(t *testing.T) {
	suite.Run(t, new(rlpTestSuite))
}
