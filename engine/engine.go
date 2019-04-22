package engine

import (
	"context"
	"encoding/json"
	"github.com/HydroProtocol/hydro-sdk-backend/common"
	"github.com/HydroProtocol/hydro-sdk-backend/config"
	"github.com/HydroProtocol/hydro-sdk-backend/connection"
	"github.com/HydroProtocol/hydro-sdk-backend/models"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/go-redis/redis"
	"sync"
)

type Engine struct {
	// all redis queues handlers
	marketHandlerMap map[string]*MarketHandler
	queue            common.IQueue

	// Wait for all queue handler exit gracefully
	Wg sync.WaitGroup

	// global ctx, if this ctx is canceled, queue handlers should exit in a short time.
	ctx context.Context
}

func NewEngine(ctx context.Context, redis *redis.Client) *Engine {
	queue, _ := common.InitQueue(&common.RedisQueueConfig{
		Name:   common.HYDRO_ENGINE_EVENTS_QUEUE_KEY,
		Client: redis,
		Ctx:    ctx,
	})

	engine := &Engine{
		queue:            queue,
		ctx:              ctx,
		marketHandlerMap: make(map[string]*MarketHandler),
		Wg:               sync.WaitGroup{},
	}

	markets := models.MarketDao.FindAllMarkets()

	for _, market := range markets {
		kvStore, _ := common.InitKVStore(
			&common.RedisKVStoreConfig{
				Ctx:    ctx,
				Client: redis,
			},
		)
		marketHandler, err := NewMarket(ctx, kvStore, market)
		if err != nil {
			panic(err)
		}

		engine.marketHandlerMap[market.ID] = marketHandler
		utils.Info("market %s init done", marketHandler.market.ID)
	}

	return engine
}

func (e *Engine) start() {
	for i := range e.marketHandlerMap {
		marketHandler := e.marketHandlerMap[i]
		e.Wg.Add(1)

		go func() {
			defer e.Wg.Done()

			utils.Info("%s market handler is running", marketHandler.market.ID)
			defer utils.Info("%s market handler is stopped", marketHandler.market.ID)

			marketHandler.Run()
		}()
	}

	go func() {
		for {
			select {
			case <-e.ctx.Done():
				for _, handler := range e.marketHandlerMap {
					close(handler.queue)
				}
				return
			default:
				data, err := e.queue.Pop()
				if err != nil {
					panic(err)
				}
				var event common.Event
				err = json.Unmarshal(data, &event)
				if err != nil {
					utils.Error("wrong event format: %+v", err)
				}

				e.marketHandlerMap[event.MarketID].queue <- data
			}
		}
	}()
}

var hydroProtocol = &ethereum.EthereumHydroProtocol{}

func Run(ctx context.Context) {
	utils.Info("engine start...")

	// init redis
	redisClient := connection.NewRedisClient(config.Getenv("HSK_REDIS_URL"))

	// init message queue
	messageQueue, _ := common.InitQueue(
		&common.RedisQueueConfig{
			Name:   common.HYDRO_WEBSOCKET_MESSAGES_QUEUE_KEY,
			Ctx:    ctx,
			Client: redisClient,
		},
	)
	InitWsQueue(messageQueue)

	//init database
	models.ConnectDatabase("postgres", config.Getenv("HSK_DATABASE_URL"))

	//start engine
	engine := NewEngine(ctx, redisClient)
	engine.start()

	engine.Wg.Wait()
	utils.Info("engine stopped!")
}
