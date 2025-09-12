## Bot Webhook Relay (Telegram)

A tiny HTTP service that accepts generic webhook requests and relays them to a Telegram bot/chat. It is tolerant of different payload shapes (for example: `msg`, `message`, `text`, `data.message`), and formats a clean HTML message with emojis, including the requester IP.

### Run

- Env vars:
  - `BOT_TOKEN` — Telegram bot token
  - `CHAT_ID` — Default chat ID to send to
  - `LISTEN_ADDR` — Bind address, default `:8080`

- From repo root:
  - `go run ./bot/botserver`
- Or inside this folder:
  - `go run .`

### Endpoints

- `GET /send?msg=...`
  - Sends plain text to the default `BOT_TOKEN`/`CHAT_ID`.
- `POST /send`
  - JSON: `{ "msg": "...", "token"?: "...", "chat_id"?: 123 }`
  - If `token` and `chat_id` are provided, the message is sent with that bot/chat.
- `GET /webhook?msg=...`
  - Returns 200 even without `msg` (for “test connection”).
  - If `msg` exists, sends plain text and appends requester IP.
- `POST /webhook`
  - Accepts JSON or form data. Tolerant fields for message extraction:
    - Top-level: `msg`, `message`, `text`, `content`, `title`+`body`
    - Nested: `text.content` (DingTalk style), `data.message`, `data.timestamp`
    - Top-level `service`, `type`, `timestamp` are also recognized
  - Builds and sends a Telegram HTML message:
    - Title with icon by `type` (test/ping 🧪, alert/error/incident 🚨, warn ⚠️, ok/info/notice ℹ️, default 🔔)
    - Lines for Service, Type, Time, IP
    - Content rendered as a quoted block
  - Always responds `200 {"status":"ok"}` for compatibility with webhook testers.

### Examples

- Basic send:
  - `curl 'http://localhost:8080/send?msg=hello'`
- Webhook JSON (generic):
  - `curl -X POST -H 'Content-Type: application/json' \\
     -d '{"service":"claude-relay-service","type":"test","timestamp":"2025-09-12T19:01:16+08:00","data":{"message":"Webhook 测试"}}' \\
     http://localhost:8080/webhook`
- Override bot/chat for one request:
  - `curl -X POST -H 'Content-Type: application/json' \\
     -d '{"msg":"hello","token":"<OTHER_BOT_TOKEN>","chat_id":123456789}' \\
     http://localhost:8080/send`

### IP Detection

- Uses Gin’s `ClientIP()` which honors `X-Forwarded-For` and `X-Real-IP` when running behind proxies. Ensure your reverse proxy forwards one of these headers so the real client IP is shown in Telegram.

### Notes

- The HTML formatter escapes content to avoid breaking Telegram HTML.
- For `/send` endpoints, `msg` is required and returns 400 if missing.
- For `/webhook`, 200 is returned even when no usable message is found to keep external “测试连接/测试发送” happy.
- Server logs will print the raw webhook payload for quick debugging.

### Roadmap

- Optional signature verification (DingTalk/飞书/企业微信) — not implemented yet.

