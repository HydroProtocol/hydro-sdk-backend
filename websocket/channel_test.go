package websocket

import (
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
	"time"
)

type MockWebsocketConnection struct {
	mock.Mock
}

func (m *MockWebsocketConnection) ReadJSON(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockWebsocketConnection) WriteJSON(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockWebsocketConnection) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

type channelTestSuit struct {
	suite.Suite
}

func (s *channelTestSuit) InitClient() (*Client, *MockWebsocketConnection) {
	client := NewClient()
	conn := new(MockWebsocketConnection)
	conn.On("WriteJSON", mock.Anything).Return(nil)

	client.Conn = conn
	return client, conn
}

func (s *channelTestSuit) SetupSuite() {
}

func (s *channelTestSuit) TearDownSuite() {
}

func (s *channelTestSuit) TearDownTest() {
	defaultSnapshotFetcher = &DefaultSnapshotFetcher{}
}

func (s *channelTestSuit) TestFind() {
	channel := FindChannel("not-exist-channel")
	s.Nil(channel)
}

func (s *channelTestSuit) TestCreate() {
	id := "test-create-id"
	channel := CreateAddressChannel(id)
	s.Equal(id, channel.ID)
}

func (s *channelTestSuit) TestRunAddressChannel() {
	channel := CreateAddressChannel("test-channel")

	time.Sleep(time.Millisecond * 20)

	client, clientConn := s.InitClient()
	client2, client2Conn := s.InitClient()

	channel.AddSubscriber(client)
	channel.AddSubscriber(client2)
	time.Sleep(time.Millisecond * 20)

	message := &common.WebSocketMessage{
		ChannelID: "test-channel",
		Payload:   []byte(`{"success": true}`),
	}

	channel.AddMessage(message)
	time.Sleep(time.Millisecond * 20)

	clientConn.AssertCalled(s.T(), "WriteJSON", message.Payload)
	client2Conn.AssertCalled(s.T(), "WriteJSON", message.Payload)

	channel.RemoveSubscriber(client.ID)
	time.Sleep(time.Millisecond * 20)

	s.Equal(false, len(channel.Clients) <= 0)

	channel.RemoveSubscriber(client2.ID)
	time.Sleep(time.Millisecond * 20)
	s.Equal(true, len(channel.Clients) <= 0)
}

func (s *channelTestSuit) TestRunOrderbookChannel() {
	mockSnapshot := &common.SnapshotV2{
		Sequence: 12,
		Bids: [][2]string{
			{
				"1", "1",
			},
		},
		Asks: [][2]string{
			{
				"2", "1",
			},
		},
	}

	defaultSnapshotFetcher = NewMockSnapshotFetcher(mockSnapshot)

	channel := CreateMarketChannel("test-channel", "HOT-WETH")
	s.Equal(true, len(channel.Clients) <= 0)
	s.Equal(mockSnapshot.Bids, channel.Orderbook.SnapshotV2().Bids)
	s.Equal(mockSnapshot.Asks, channel.Orderbook.SnapshotV2().Asks)
	s.Equal(mockSnapshot.Sequence, channel.Orderbook.Sequence)

	c1, c1Connection := s.InitClient()
	channel.AddSubscriber(c1)
	time.Sleep(time.Millisecond * 20)
	s.Equal(false, len(channel.Clients) <= 0)
	c1Connection.AssertNumberOfCalls(s.T(), "WriteJSON", 1)

	// test receive overdue message
	// first overdue message
	channel.AddMessage(s.buildWesocketMessage(11, "buy", "1", "1"))
	time.Sleep(time.Millisecond * 20)
	c1Connection.AssertNumberOfCalls(s.T(), "WriteJSON", 1)
	s.Equal(uint64(12), channel.Orderbook.Sequence)

	// second overdue message
	channel.AddMessage(s.buildWesocketMessage(12, "buy", "1", "1"))
	time.Sleep(time.Millisecond * 20)
	c1Connection.AssertNumberOfCalls(s.T(), "WriteJSON", 1)
	s.Equal(uint64(12), channel.Orderbook.Sequence)

	// first valid message
	channel.AddMessage(s.buildWesocketMessage(13, "buy", "1", "1"))
	time.Sleep(time.Millisecond * 20)
	c1Connection.AssertNumberOfCalls(s.T(), "WriteJSON", 2)
	s.Equal(uint64(13), channel.Orderbook.Sequence)
	s.Equal([2]string{"1", "2"}, channel.Orderbook.SnapshotV2().Bids[0])
	s.Equal(mockSnapshot.Asks, channel.Orderbook.SnapshotV2().Asks)
}

func (s *channelTestSuit) buildWesocketMessage(sequence uint64, side, price, changedAmount string) *common.WebSocketMessage {

	payload := &common.WebsocketMarketOrderChangePayload{
		Side:     side,
		Price:    price,
		Amount:   changedAmount,
		Sequence: sequence,
	}

	return &common.WebSocketMessage{
		Payload: payload,
	}
}

func TestChannelSuit(t *testing.T) {
	suite.Run(t, new(channelTestSuit))
}
