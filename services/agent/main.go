package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Ljx201120/ai-agent-platform/agent/config"
	"github.com/Ljx201120/ai-agent-platform/agent/internal/consumer"
	"github.com/Ljx201120/ai-agent-platform/agent/internal/llm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌加载配置失败: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	// 初始化 LLM 客户端
	llmCli := llm.NewDeepSeekClient(cfg.DeepSeekKey, cfg.DeepSeekURL)

	// 启动消费者
	c := consumer.NewTaskConsumer(rdb, db, llmCli)
	c.Start(context.Background())
}
