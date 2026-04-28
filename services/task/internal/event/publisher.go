package event

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const StreamTaskCreated = "task.created"

type TaskCreatedEvent struct {
	TaskID string `json:"task_id"`
	Prompt string `json:"prompt"`
}

type Publisher struct {
	rdb *redis.Client
}

func NewPublisher(rdb *redis.Client) *Publisher {
	return &Publisher{rdb: rdb}
}

func (p *Publisher) PublishTaskCreated(ctx context.Context, taskID, prompt string) error {
	event := TaskCreatedEvent{
		TaskID: taskID,
		Prompt: prompt,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamTaskCreated,
		Values: map[string]any{"data": string(data)},
	}).Err()
}
