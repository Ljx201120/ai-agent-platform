package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost      string
	DBPort      int // 注意：为了校验方便，改为 int 类型
	DBUser      string
	DBPassword  string
	DBName      string
	RedisAddr   string
	DeepSeekKey string
	DeepSeekURL string
}

// Load 加载 .env 并校验配置，返回 Config 或错误
func Load() (*Config, error) {
	// 加载 .env 文件（如果不存在则忽略）
	_ = godotenv.Load()

	cfg := &Config{}

	// 1. DB_HOST（必填）
	if v := os.Getenv("DB_HOST"); v == "" {
		return nil, fmt.Errorf("环境变量 DB_HOST 不能为空")
	} else {
		cfg.DBHost = v
	}

	// 2. DB_PORT（必填，且为有效数字）
	portStr := os.Getenv("DB_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("环境变量 DB_PORT 不能为空")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("环境变量 DB_PORT 必须是 1~65535 的数字，当前值为: %s", portStr)
	}
	cfg.DBPort = port

	// 3. DB_USER（必填）
	if v := os.Getenv("DB_USER"); v == "" {
		return nil, fmt.Errorf("环境变量 DB_USER 不能为空")
	} else {
		cfg.DBUser = v
	}

	// 4. DB_PASSWORD（可选，但如果存在不能为空字符串，这里允许空）
	cfg.DBPassword = os.Getenv("DB_PASSWORD")

	// 5. DB_NAME（必填）
	if v := os.Getenv("DB_NAME"); v == "" {
		return nil, fmt.Errorf("环境变量 DB_NAME 不能为空")
	} else {
		cfg.DBName = v
	}

	// 6. REDIS_ADDR（必填）
	if v := os.Getenv("REDIS_ADDR"); v == "" {
		return nil, fmt.Errorf("环境变量 REDIS_ADDR 不能为空")
	} else {
		// 简单格式检查：应该包含冒号
		if !strings.Contains(v, ":") {
			return nil, fmt.Errorf("REDIS_ADDR 格式应为 host:port，当前: %s", v)
		}
		cfg.RedisAddr = v
	}

	// 7. DEEPSEEK_API_KEY（必填，因为调用 API 需要）
	if v := os.Getenv("DEEPSEEK_API_KEY"); v == "" {
		return nil, fmt.Errorf("环境变量 DEEPSEEK_API_KEY 不能为空")
	} else {
		cfg.DeepSeekKey = v
	}

	// 8. DEEPSEEK_BASE_URL（可选但有默认值，不强制）
	if v := os.Getenv("DEEPSEEK_BASE_URL"); v != "" {
		cfg.DeepSeekURL = v
	} else {
		cfg.DeepSeekURL = "https://api.deepseek.com" // 仅这一个保留默认值
	}

	return cfg, nil
}
