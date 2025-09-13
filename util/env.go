package util

import (
    "os"
    "strconv"
)

// EnvOrDefault 返回环境变量值，不存在则返回默认值
func EnvOrDefault(key, def string) string {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    return v
}

// EnvBoolDefault 读取布尔环境变量（"1","true","TRUE","True" 视为 true），不存在返回默认值
func EnvBoolDefault(key string, def bool) bool {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    switch v {
    case "1", "true", "TRUE", "True":
        return true
    case "0", "false", "FALSE", "False":
        return false
    default:
        // 尝试解析数字
        if i, err := strconv.Atoi(v); err == nil {
            return i != 0
        }
        return def
    }
}

// EnvIntDefault 读取整型环境变量，不存在或解析失败返回默认值
func EnvIntDefault(key string, def int) int {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    if i, err := strconv.Atoi(v); err == nil {
        return i
    }
    return def
}

// EnvPGURL 返回 Postgres 连接串（默认 postgresql://dev:123@psql.loc:5432/base）
func EnvPGURL() string {
    return EnvOrDefault("PG_URL", "postgresql://dev:123@psql.loc:5432/base")
}

// EnvRedisURI 返回 Redis 连接 URI（默认 redis.loc:6379?db=1）
func EnvRedisURI() string {
    return EnvOrDefault("REDIS_URI", "redis.loc:6379?db=1")
}

