# AI Agent Platform

基于 Go 微服务架构的 AI 任务调度平台，支持异步任务队列、LLM 调用和容器化部署。

## 项目架构
用户请求 → Gateway Service → Task Service → Redis Stream → Agent Service → DeepSeek API
↓                              ↓
PostgreSQL                     PostgreSQL
## 技术栈

- **语言**：Go 1.24
- **Web 框架**：Gin
- **数据库**：PostgreSQL + GORM
- **消息队列**：Redis Stream
- **LLM**：DeepSeek API
- **容器化**：Docker + Docker Compose
- **CI/CD**：GitHub Actions

## 服务说明

| 服务 | 端口 | 职责 |
|------|------|------|
| task-service | 8081 | 任务 CRUD、发布事件到 Redis Stream |
| agent-service | - | 消费事件、调用 LLM、写回结果 |

## 快速开始

### 环境要求

- Docker
- Docker Compose

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 填入 DEEPSEEK_API_KEY
```

### 启动服务

```bash
docker-compose up --build
```

### API 使用

**创建任务**

```bash
curl -X POST http://localhost:8081/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "prompt": "用一句话介绍北京"
  }'
```

**查询任务**

```bash
curl http://localhost:8081/api/tasks/{task_id}
```

**查询用户所有任务**

```bash
curl "http://localhost:8081/api/tasks?user_id=550e8400-e29b-41d4-a716-446655440000"
```

### 任务状态

| 状态 | 说明 |
|------|------|
| pending | 任务已创建，等待处理 |
| running | Agent 正在处理 |
| completed | 处理完成 |
| failed | 处理失败 |

## 项目结构
ai-agent-platform/
├── services/
│   ├── task/               # 任务管理服务
│   │   ├── internal/
│   │   │   ├── model/      # 数据模型
│   │   │   ├── repository/ # 数据库操作
│   │   │   ├── service/    # 业务逻辑
│   │   │   ├── handler/    # HTTP handler
│   │   │   └── event/      # 事件发布
│   │   └── Dockerfile
│   └── agent/              # Agent 执行服务
│       ├── internal/
│       │   ├── llm/        # DeepSeek 客户端
│       │   └── consumer/   # Redis Stream 消费者
│       └── Dockerfile
├── proto/                  # gRPC 定义（规划中）
├── docker-compose.yml
└── .github/workflows/      # CI/CD

## 开发

本地开发需要先启动依赖服务：

```bash
# 启动 PostgreSQL 和 Redis
docker-compose up postgres redis

# 启动 task-service
cd services/task
DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres DB_NAME=agent_platform REDIS_ADDR=localhost:6379 go run main.go

# 启动 agent-service
cd services/agent
DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres DB_NAME=agent_platform REDIS_ADDR=localhost:6379 DEEPSEEK_API_KEY=你的key go run main.go
```

## CI/CD

推送到 `main` 分支或提交 PR 时自动触发：

- Go 代码构建
- 单元测试
- Docker 镜像构建