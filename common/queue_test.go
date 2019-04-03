package common

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type redisQueueTest struct {
	suite.Suite
}

func (s *redisQueueTest) SetupSuite() {
}

func (s *redisQueueTest) SetupTest() {
}

func (s *redisQueueTest) TearDownTest() {
}

func (s *redisQueueTest) TearDownSuite() {
}

//func (s *redisQueueTest) TestPushAndPot() {
//	redis := connection.NewRedisClient("redis://localhost:6379")
//
//	queue, err := InitQueue(&RedisQueueConfig{
//		Name:   "test",
//		Ctx:    context.Background(),
//		Client: redis,
//	})
//
//	if err != nil {
//		panic(err)
//	}
//
//	bts := make([]byte, 65535)
//	rand.Read(bts)
//
//	_ = queue.Push(bts)
//	bts2, _ := queue.Pop()
//
//	s.Equal(bts, bts2)
//}

func TestRedisQueue(t *testing.T) {
	suite.Run(t, new(redisQueueTest))
}
