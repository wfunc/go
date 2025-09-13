package util

// AppConfig 为通用应用配置
type AppConfig struct {
    Listen   string
    PGURL    string
    RedisURI string
}

// LoadAppConfig 加载应用配置，顺序：.env -> 环境变量 -> 命令行 flags（flags 优先级最高）
// 支持 flags：--listen/--pg/--redis
func LoadAppConfig() AppConfig {
    // 1) 读取 .env（不覆盖已有环境变量）
    _ = LoadDotEnv(".env")

    // 2) flags + env + 默认值
    flags := ParseFlagsWithEnv(
        FlagSpec{Name: "listen", EnvKey: "LISTEN_ADDR", Default: ":8080", Usage: "监听地址"},
        FlagSpec{Name: "pg", EnvKey: "PG_URL", Default: EnvPGURL(), Usage: "PostgreSQL 连接串"},
        FlagSpec{Name: "redis", EnvKey: "REDIS_URI", Default: EnvRedisURI(), Usage: "Redis 连接 URI"},
    )

    return AppConfig{
        Listen:   flags["listen"],
        PGURL:    flags["pg"],
        RedisURI: flags["redis"],
    }
}

