package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Ljx201120/ai-agent-platform/task/config"
	"github.com/Ljx201120/ai-agent-platform/task/internal/event"
	"github.com/Ljx201120/ai-agent-platform/task/internal/handler"
	"github.com/Ljx201120/ai-agent-platform/task/internal/model"
	"github.com/Ljx201120/ai-agent-platform/task/internal/repository"
	"github.com/Ljx201120/ai-agent-platform/task/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌加载配置失败: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := db.AutoMigrate(&model.Task{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	publisher := event.NewPublisher(rdb)
	repo := repository.NewTaskRepository(db)
	svc := service.NewTaskService(repo, publisher)
	h := handler.NewTaskHandler(svc)

	r := gin.Default()
	h.RegisterRoutes(r)

	log.Println("Task Service 启动，监听 :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
