# Hydro SDK Backend

[![CircleCI](https://circleci.com/gh/HydroProtocol/hydro-sdk-backend.svg?style=svg)](https://circleci.com/gh/HydroProtocol/hydro-sdk-backend)
[![Go Report Card](https://goreportcard.com/badge/github.com/hydroprotocol/hydro-sdk-backend)](https://goreportcard.com/report/github.com/hydroprotocol/hydro-sdk-backend)
[![microbadger](https://images.microbadger.com/badges/image/hydroprotocolio/hydro-sdk-backend.svg)](https://microbadger.com/images/hydroprotocolio/hydro-sdk-backend)
[![Docker Pulls](https://img.shields.io/docker/pulls/hydroprotocolio/hydro-sdk-backend.svg)](https://hub.docker.com/r/hydroprotocolio/hydro-sdk-backend)
[![Docker Cloud Automated build](https://img.shields.io/docker/cloud/automated/hydroprotocolio/hydro-sdk-backend.svg)](https://hub.docker.com/r/hydroprotocolio/hydro-sdk-backend)
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/hydroprotocolio/hydro-sdk-backend.svg)](https://hub.docker.com/r/hydroprotocolio/hydro-sdk-backend)

```
go mod download
```


## Break down to each part

### Watcher

The Blockchain Watcher is responsible for monitoring blockchain block. 
Each transaction on blockchain will go through this method.
You should register a handler via `RegisterHandler` function.
In this handler, you can emit events and routing them to the proper component.

### Websocket

The Websocket package allows you to start a websocket server easily. 
The server is channel based.
A user can join multiple channels, and can leave at any time.

The websocket server should have a message source. 
Every message read from the source will be broadcast to the channel.
All users in the channel will receive the message.

If you want to make some special logic other than just broadcast the message.
It can be achieved by creating your own channel. 

Any struct implemented the IChannel interface can be registered into the websocket server.

There are already a customized channel called `MarketChannel` in this package. 
It keep maintaining the newest order book in memory.  
If a new user joins this channel, 
it will send a snapshot of current market order book to the user.
And after receive a new event from source, 
it will update the order book in memory, 
then push the change event to all subscribers.

```golang
import (
    github.com/hydroprotocol/hydor-sdk-backend/common
    github.com/hydroprotocol/hydor-sdk-backend/websocket
)

// new a source queue
queue, _ := common.InitQueue(&common.RedisQueueConfig{
    Name:   common.HYDRO_WEBSOCKET_MESSAGES_QUEUE_KEY,
    Ctx:    ctx,
    Client: redisClient,
})

// new a websockert server
wsServer := websocket.NewWSServer("localhost:3002", queue)

websocket.RegisterChannelCreator(
    common.MarketChannelPrefix,
    websocket.NewMarketChannelCreator(&websocket.DefaultHttpSnapshotFetcher{
        ApiUrl: os.Getenv("HSK_API_URL"),
    }),
)

// Start the server
// It will block the current process to listen on the `addr` your provided. 
wsServer.Start()
```