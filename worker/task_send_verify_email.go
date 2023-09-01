package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaksSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistibutor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create a worker: %s", err)
	}
	task := asynq.NewTask(TaksSendVerifyEmail, jsonPayload, opts...)
	taskinfo, err := distributor.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("failed to create a worker: %s", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("queue", taskinfo.Queue).
		Int("max_retry", taskinfo.MaxRetry).Msg("enqueue task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendverifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshall payload: %w ", asynq.SkipRetry)
	}
	user, err := processor.store.GetUsers(ctx, payload.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("failed to get user: %w ", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w ", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
