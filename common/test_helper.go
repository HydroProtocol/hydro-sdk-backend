package common

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockKVStore struct {
	mock.Mock
}

func (m *MockKVStore) Set(key string, value string, expire time.Duration) error {
	args := m.Called(key, value, expire)
	return args.Error(0)
}

func (m *MockKVStore) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

type MockQueue struct {
	mock.Mock
	Buffers [][]byte
}

func (m *MockQueue) Push(bts []byte) error {
	args := m.Called(bts)

	if m.Buffers == nil {
		m.ResetBuffer()
	}

	m.Buffers = append(m.Buffers, bts)
	return args.Error(0)
}

func (m *MockQueue) Pop() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockQueue) ResetBuffer() {
	m.Buffers = make([][]byte, 0, 0)
}
