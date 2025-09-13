# QuickStart 示例

一个最小可运行的示例，展示如何快速接入会话（Session）、短信（SMS）、邮件（Email）和数据库（basedb）。

## 运行

```bash
# 可选：覆盖默认环境变量
export REDIS_URI="redis.loc:6379?db=1"
export PG_URL="postgresql://dev:123@psql.loc:5432/base"

# 一键启动依赖（Postgres/Redis/可选面板/botserver）
cd examples/quickstart && docker compose up -d && cd -

# 运行基础示例（标准 http.Server）
go run ./examples/quickstart --listen :8080

# 运行标准 http.Server 版本
go run ./examples/quickstart-httpserver --listen :8080

# 运行 Gin 集成版本
go run ./examples/quickstart-gin --listen :8080
```

启动后默认监听 :8080（可通过 LISTEN_ADDR 或 --listen 覆盖），可在浏览器访问：

- /set：设置会话（写 Cookie）
- /get：读取会话
- /pub/sendLoginSms?phone=1234567890：发送登录短信（示例中只打印日志）
- /pub/loadPhoneCode?key=login&phone=1234567890：调试读取短信验证码（仅示例）
- /pub/sendLoginEmail?email=demo@example.com：发送登录邮件（示例中只打印日志）
- /pub/loadEmailCode?key=login&email=demo@example.com：调试读取邮件验证码（仅示例）
- /conf/set：向数据库存储配置（需要 PG_URL 可用）
- /conf/get：从数据库读取配置

### 管理面板

- pgAdmin（PostgreSQL）：http://localhost:5050
  - 登录：邮箱 admin@admin.com，密码 admin
  - 添加服务器：
    - Name：随意（如 base）
    - Host：postgres（容器内服务名）或 127.0.0.1
    - Port：5432
    - Username：dev
    - Password：123
    - Database：base
- RedisInsight：http://localhost:5540
  - 添加 Redis 连接：
    - Host：redis（容器内服务名）或 127.0.0.1
    - Port：6379

### BotServer（Telegram）

- 配置环境变量（任选其一）：
  - 在仓库根目录 `.env` 中添加：
    - BOT_TOKEN=你的机器人 Token
    - CHAT_ID=你的聊天 ID
  - 或在 `examples/quickstart/.env` 中添加同名变量
- 启动服务：
  - 在 examples/quickstart 目录下执行：`make bot-up` 或 `docker compose up -d botserver`
  - 监听地址：http://localhost:8082
- 快速测试：
  - 文本测试：`curl "http://localhost:8082/webhook?msg=hello"`
  - JSON 测试：
    ```bash
    curl -X POST "http://localhost:8082/webhook" \
      -H 'Content-Type: application/json' \
      -d '{
        "service":"demo",
        "type":"test",
        "message":"Hello from webhook",
        "timestamp":"2025-01-01T00:00:00Z"
      }'
    ```
  - 使用特定 token/chat：
    ```bash
    curl -X POST "http://localhost:8082/webhook" \
      -H 'Content-Type: application/json' \
      -d '{
        "msg":"override token",
        "token":"YOUR_TOKEN",
        "chat_id":123456789
      }'
    ```

#### 安全与 TLS

- 可选签名校验：设置环境变量 `BOT_WEBHOOK_SECRET`，并在请求头携带 `X-Signature`（HMAC-SHA256 原始 Body 的十六进制小写值）。
- 可选 TLS：设置 `TLS_CERT_FILE` 与 `TLS_KEY_FILE` 后，服务将以 HTTPS 方式启动。

### Makefile（简化命令）

在 examples/quickstart 目录下：

- `make up` / `make down` / `make ps` / `make logs`
- `make run`（标准 http.Server 示例，默认 :8080）
- `make run-http`（备用 httpServer 示例，默认 :8081）
- `make run-gin`（Gin 集成示例，默认 :8082）
- `make bot-up`（仅启动 botserver）/ `make bot-logs`（查看日志）
- `make seed`（写入示例 Seed 数据到数据库）
- `make clean`（移除容器与数据卷，慎用）

### BotServer 队列与管理

- 失败缓存与重放：
  - 开关：`BOT_ENABLE_BACKLOG=1`（默认开启）
  - 路径：`BACKLOG_PATH=/home/app/data/backlog.jsonl`
  - 定时重放间隔（秒）：`BACKLOG_REPLAY_INTERVAL=60`
  - 管理接口（可选鉴权）：设置 `QUEUE_ADMIN_TOKEN` 后，调用队列接口需在请求头携带 `X-Admin-Token`；
    - GET `/queue/stats`：查看待重放数量
    - POST `/queue/replay`：立即重放

## 说明

- 示例默认使用标准 http.Server；亦提供 quickstart-gin 做 Gin 集成参考。
- 提供 docker-compose.yml，支持本地一键启动 Postgres 与 Redis。
- Session 默认在 HTTP 环境下使用 SameSite=Lax 且不设置 Secure，避免 Cookie 丢失；在 HTTPS 环境下为 SameSite=None 且 Secure=true。你也可以通过：
  - 全局：`session.SetDefaultCookiePolicy(...)`
  - 构造器选项：`session.NewDbSessionBuilder(session.WithCookie...)`
  进行覆盖。
- 短信与邮件的发送实现使用 `UseSender(...)` 注入，你可以接入真实服务商；验证码为 `templateParam["code"]`。
