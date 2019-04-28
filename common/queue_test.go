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

func TestRedisQueue(t *testing.T) {
	suite.Run(t, new(redisQueueTest))
}
