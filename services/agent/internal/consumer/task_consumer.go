package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Ljx201120/ai-agent-platform/agent/internal/llm"
)

const (
	streamName   = "task.created"
	groupName    = "agent-group"
	consumerName = "agent-1"
)

type TaskEvent struct {
	TaskID string `json:"task_id"`
	Prompt string `json:"prompt"`
}

type TaskConsumer struct {
	rdb    *redis.Client
	db     *gorm.DB
	llmCli *llm.DeepSeekClient
}

func NewTaskConsumer(rdb *redis.Client, db *gorm.DB, llmCli *llm.DeepSeekClient) *TaskConsumer {
	return &TaskConsumer{rdb: rdb, db: db, llmCli: llmCli}
}

func (c *TaskConsumer) Start(ctx context.Context) {
	// 创建消费者组（已存在则忽略）
	c.rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0")

	log.Println("Agent Consumer 开始监听任务...")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.poll(ctx)
		}
	}
}

func (c *TaskConsumer) poll(ctx context.Context) {
	result, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, ">"},
		Count:    1,
		Block:    2 * time.Second,
	}).Result()

	if err != nil {
		return
	}

	for _, stream := range result {
		for _, msg := range stream.Messages {
			c.process(ctx, msg)
		}
	}
}

func (c *TaskConsumer) process(ctx context.Context, msg redis.XMessage) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		return
	}

	var event TaskEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		log.Printf("解析任务事件失败: %v", err)
		return
	}

	log.Printf("处理任务: %s", event.TaskID)

	// 更新状态为 running
	c.db.Exec("UPDATE tasks SET status='running' WHERE id=?", event.TaskID)

	// 调用 LLM
	result, err := c.llmCli.Chat(ctx, event.Prompt)
	if err != nil {
		log.Printf("LLM 调用失败: %v", err)
		c.db.Exec("UPDATE tasks SET status='failed' WHERE id=?", event.TaskID)
	} else {
		c.db.Exec("UPDATE tasks SET status='completed', result=? WHERE id=?", result, event.TaskID)
	}

	// 确认消息
	c.rdb.XAck(ctx, streamName, groupName, msg.ID)
	log.Printf("任务完成: %s", event.TaskID)
}
