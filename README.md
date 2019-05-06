# Hydro SDK Backend

[![CircleCI](https://circleci.com/gh/HydroProtocol/hydro-sdk-backend.svg?style=svg)](https://circleci.com/gh/HydroProtocol/hydro-sdk-backend)
[![Go Report Card](https://goreportcard.com/badge/github.com/hydroprotocol/hydro-sdk-backend)](https://goreportcard.com/report/github.com/hydroprotocol/hydro-sdk-backend)

The Hydro SDK is a collection of golang language packages.
You can use it to build a Dapp application backend based on the Hydro contract quickly. 
It can help to communicate with Ethereum node, match orders, monitor Ethereum results and so on. 
Some general data structures are also provided.

This project cannot be used alone.
You need to add your own application logic. 
The following projects are built on top of this SDK.

- [hydro-scaffold-dex](https://github.com/hydroprotocol/hydro-scaffold-dex) 
- [hydro-augur-scaffold](https://github.com/hydroprotocol/hydro-augur-scaffold) (working in progress)

## Break down to each package

### sdk

The main function of this package is to define the interface to communicate with a blockchain.
We have implemented Ethereum communication codes based on this interface spec.
So as long as the interface is implemented for a blockchain, 
hydro SDK backend can be used on top it.This makes it possible to support multi-chain environments easily.

### common

We put some common data structures and interface definitions into this package for sharing with other projects.

### engine

The engine maintains a series of market orderbooks. 
It is responsible for handling all placing orders and cancel requests. 
Requests in each market are processed serially, 
and multiple markets are concurrent.

The engine in this package only maintains the orderbook based on the received message 
and returns the result of the operation. 
It is not responsible for persisting these changes, 
nor for pushing messages to users. 
Persistent data and push messages are business logic and should be done by the upper application.


### watcher

Blockchain Watcher is responsible for monitoring blockchain changes. 
Whenever a new block is generated, 
it gets all the transactions in that block. 
And pass each transaction to a specific method to deal with. 
This method requires you to register with the `RegisterHandler` function. 
You can process the transactions you are interested in as needed and skip unrelated transactions.

### websocket

The Websocket package allows you to easily launch a websocket server. 
The server is channel based.
Users can join multiple channels and can leave at any time.

The websocket server should have a message source. 
Every message read from the source will be broadcast to that channel.
All users in the channel will receive this message.

If you want to make some special logic and not just broadcast the message.
This can be done by creating your own channel. 

Any structure that implements the IChannel interface can be registered to the websocket server.

There are already a customized channel called `MarketChannel` in this package. 
It keep maintaining the newest order book in memory.  
If a new user joins this channel, 
it sends a snapshot of current market order book to the user.
After receive a new event from source, 
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

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details
