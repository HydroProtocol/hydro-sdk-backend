package websocket

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
)

var channelCreators = make(map[string]func(channelID string) IChannel)

type WSServer struct {
	addr        string        // addr the websocket is listened on
	sourceQueue common.IQueue // a queue to get
}

func NewWSServer(addr string, sourceQueue common.IQueue) *WSServer {
	if addr == "" {
		addr = ":3002"
	}

	s := &WSServer{
		addr:        addr,
		sourceQueue: sourceQueue,
	}

	return s
}

func RegisterChannelCreator(prefix string, fn func(channelID string) IChannel) {
	channelCreators[prefix] = fn
}

func (s *WSServer) Start(ctx context.Context) {
	go startConsumer(ctx, s.sourceQueue)
	startSocketServer(ctx, s.addr)
}
