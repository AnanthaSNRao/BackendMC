package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistibutor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistibutor struct {
	client *asynq.Client
}

func NewRedisTaskDistibutor(redisOpt asynq.RedisClientOpt) TaskDistibutor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistibutor{
		client: client,
	}
}
