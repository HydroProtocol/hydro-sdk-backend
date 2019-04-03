package websocket

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/stretchr/testify/mock"
)

type MockSnapshotFetcher struct {
	mock.Mock
}

func (m *MockSnapshotFetcher) GetV2(marketID string) *common.SnapshotV2 {
	args := m.Called(marketID)
	return args.Get(0).(*common.SnapshotV2)
}

func NewMockSnapshotFetcher(expectedSnapshot *common.SnapshotV2) *MockSnapshotFetcher {
	fetcher := new(MockSnapshotFetcher)
	fetcher.On("GetV2", mock.Anything).Return(expectedSnapshot)
	return fetcher
}
