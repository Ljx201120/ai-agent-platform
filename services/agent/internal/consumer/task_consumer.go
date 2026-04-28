package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Ljx201120/ai-agent-platform/agent/internal/llm"
)

const (
	streamName   = "task.created"
	retryStream  = "task.retry"
	groupName    = "agent-group"
	consumerName = "agent-1"
	maxRetries   = 3
	llmTimeout   = 30 * time.Second
	workerCount  = 5
)

type TaskEvent struct {
	TaskID   string `json:"task_id"`
	Prompt   string `json:"prompt"`
	RetryNum int    `json:"retry_num"`
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
	// 创建消费者组
	c.rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0")
	c.rdb.XGroupCreateMkStream(ctx, retryStream, groupName, "0")

	// 启动定期扫描 running 状态残留的任务
	go c.scanStuckTasks(ctx)

	log.Println("Agent Consumer 开始监听任务...")

	// 用 channel 控制并发
	msgCh := make(chan redis.XMessage, workerCount)

	// 启动多个 worker goroutine
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range msgCh {
				c.process(ctx, msg, streamName)
			}
		}()
	}

	// 主循环拉取消息
	for {
		select {
		case <-ctx.Done():
			close(msgCh)
			wg.Wait()
			return
		default:
			c.poll(ctx, msgCh)
		}
	}
}

func (c *TaskConsumer) poll(ctx context.Context, msgCh chan redis.XMessage) {
	result, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, retryStream, ">", ">"},
		Count:    int64(workerCount),
		Block:    2 * time.Second,
	}).Result()

	if err != nil {
		return
	}

	for _, stream := range result {
		for _, msg := range stream.Messages {
			msgCh <- msg
		}
	}
}

func (c *TaskConsumer) process(ctx context.Context, msg redis.XMessage, stream string) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		c.rdb.XAck(ctx, stream, groupName, msg.ID)
		return
	}

	var event TaskEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		log.Printf("解析任务事件失败: %v", err)
		c.rdb.XAck(ctx, stream, groupName, msg.ID)
		return
	}

	log.Printf("处理任务: %s (重试次数: %d)", event.TaskID, event.RetryNum)

	// 更新状态为 running，如果失败不继续处理
	if err := c.db.Exec(
		"UPDATE tasks SET status='running', updated_at=NOW() WHERE id=?",
		event.TaskID,
	).Error; err != nil {
		log.Printf("更新任务状态为 running 失败: %v", err)
		return // 不 XAck，让消息重新被消费
	}

	// 设置 LLM 超时
	llmCtx, cancel := context.WithTimeout(ctx, llmTimeout)
	defer cancel()

	result, err := c.llmCli.Chat(llmCtx, event.Prompt)
	if err != nil {
		log.Printf("LLM 调用失败: %v (任务: %s)", err, event.TaskID)
		c.handleFailure(ctx, msg, stream, event)
		return
	}

	// 更新为 completed
	if err := c.db.Exec(
		"UPDATE tasks SET status='completed', result=?, updated_at=NOW() WHERE id=?",
		result, event.TaskID,
	).Error; err != nil {
		log.Printf("更新任务状态为 completed 失败: %v", err)
		return // 不 XAck，让消息重新被消费
	}

	// 确认消息
	c.rdb.XAck(ctx, stream, groupName, msg.ID)
	log.Printf("任务完成: %s", event.TaskID)
}

func (c *TaskConsumer) handleFailure(ctx context.Context, msg redis.XMessage, stream string, event TaskEvent) {
	if event.RetryNum >= maxRetries {
		// 超过最大重试次数，标记为 failed
		if err := c.db.Exec(
			"UPDATE tasks SET status='failed', updated_at=NOW() WHERE id=?",
			event.TaskID,
		).Error; err != nil {
			log.Printf("更新任务状态为 failed 失败: %v", err)
			return
		}
		c.rdb.XAck(ctx, stream, groupName, msg.ID)
		log.Printf("任务超过最大重试次数，标记为 failed: %s", event.TaskID)
		return
	}

	// 放入重试队列
	event.RetryNum++
	data, _ := json.Marshal(event)
	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: retryStream,
		Values: map[string]any{"data": string(data)},
	}).Err(); err != nil {
		log.Printf("放入重试队列失败: %v", err)
		return
	}

	c.rdb.XAck(ctx, stream, groupName, msg.ID)
	log.Printf("任务放入重试队列 (第 %d 次): %s", event.RetryNum, event.TaskID)
}

// scanStuckTasks 定期扫描长时间处于 running 状态的任务
func (c *TaskConsumer) scanStuckTasks(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stuckTimeout := time.Now().Add(-5 * time.Minute)
			result := c.db.Exec(
				"UPDATE tasks SET status='failed', updated_at=NOW() WHERE status='running' AND updated_at < ?",
				stuckTimeout,
			)
			if result.Error != nil {
				log.Printf("扫描残留 running 任务失败: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("清理残留 running 任务: %d 条", result.RowsAffected)
			}
		}
	}
}

// handleFailure 中计算退避延迟（可选）
func retryDelay(retryNum int) time.Duration {
	return time.Duration(retryNum*retryNum) * time.Second
}

func init() {
	_ = fmt.Sprintf // 避免未使用导入报错
}
