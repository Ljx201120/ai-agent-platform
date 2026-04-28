package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     int // 改为 int 类型，方便校验
	DBUser     string
	DBPassword string
	DBName     string
	RedisAddr  string
}

// Load 加载 .env 文件并校验配置，返回 Config 或错误
func Load() (*Config, error) {
	// 加载 .env 文件（忽略文件不存在的错误）
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     0, // 先设置为默认值，后续会从环境变量中解析
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
	}

	// 1. 校验 DB_HOST
	if cfg.DBHost == "" {
		return nil, fmt.Errorf("环境变量 DB_HOST 不能为空")
	}

	// 2. 校验 DB_PORT（必须为有效数字）
	portStr := os.Getenv("DB_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("环境变量 DB_PORT 不能为空")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("环境变量 DB_PORT 必须是 1~65535 的数字，当前值为: %s", portStr)
	}
	cfg.DBPort = port

	// 3. 校验 DB_USER
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("环境变量 DB_USER 不能为空")
	}

	// 4. DB_PASSWORD 可为空（允许无密码），只校验不为空字符串时可选
	// 如果你的数据库强制需要密码，取消下面的注释
	// if cfg.DBPassword == "" {
	//     return nil, fmt.Errorf("环境变量 DB_PASSWORD 不能为空")
	// }

	// 5. 校验 DB_NAME
	if cfg.DBName == "" {
		return nil, fmt.Errorf("环境变量 DB_NAME 不能为空")
	}

	// 6. 校验 RedisAddr
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("环境变量 REDIS_ADDR 不能为空")
	}

	return cfg, nil
}
