package websocket

import (
	"context"
	"encoding/json"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
)

// StartConsumer initializes a queue instance and ready events from it
func StartConsumer(ctx context.Context, queueConfig interface{}) {
	queue, err := common.InitQueue(queueConfig)

	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			utils.Info("Websocket Consumer Exit")
			return
		default:

			// This method should not block this go thread all the time to make it has chance to exit gracefully
			msg, err := queue.Pop()

			if err != nil {
				utils.Error("read message error %v", err)
				continue
			}

			var wsMsg common.WebSocketMessage

			_ = json.Unmarshal(msg, &wsMsg)

			channel := FindChannel(wsMsg.ChannelID)

			if channel == nil {
				channel = CreateChannelByID(wsMsg.ChannelID)
			}

			channel.AddMessage(&wsMsg)
		}
	}
}
