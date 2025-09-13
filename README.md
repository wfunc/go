# WFunc Go Library

A comprehensive Go library providing reusable components for building web applications with session management, authentication, messaging, and database utilities.

## Overview

WFunc Go is a modular library designed to accelerate web application development by providing battle-tested components for common functionality including:

- Session management with Redis caching
- SMS and Email verification systems
- Telegram bot integration
- Database migrations and ORM utilities
- HTTP transport and WebSocket support
- Captcha generation and verification
- Configuration management
- File upload handling

## Features

### Core Components

#### 🔐 **Session Management** (`session/`)
- Redis-backed session storage
- Cookie-based session handling
- Secure session lifecycle management
- Integration with Gin web framework

#### 📱 **SMS Service** (`sms/`)
- Phone number verification
- SMS code generation and validation
- Redis-based code storage
- Support for login and verification codes
- Captcha integration for security

#### 📧 **Email Service** (`email/`)
- Email verification system
- Login code generation
- SMTP email sender configuration
- Case-insensitive code validation
- Redis-based code caching

#### 🤖 **Telegram Bot** (`bot/`)
- Telegram Bot API integration
- Message sending utilities
- HTML and Markdown message formatting
- Deposit/Withdrawal notification templates
- Bot server implementation

#### 🗄️ **Database Utilities** (`basedb/`)
- PostgreSQL integration with pgx
- Auto-generated models and functions
- Configuration management
- Object storage patterns
- Announcement system

#### 🔄 **Database Migrations** (`baseupgrade/`)
- SQL migration management
- Version tracking
- Automated upgrade scripts
- Database schema initialization

#### 🌐 **API Base** (`baseapi/`)
- RESTful API handlers
- Configuration API endpoints
- Version object management
- File upload handling
- Announcement system APIs

#### 🔗 **Transport Layer** (`transport/`)
- HTTP request forwarding
- WebSocket handler implementation
- Proxy functionality

#### 🛡️ **Captcha** (`captcha/`)
- Image captcha generation
- Captcha verification endpoints
- Default configuration

#### 📊 **Logging** (`xlog/`)
- Structured logging with zap
- Production and development configurations
- Log level management

#### 🛠️ **Utilities** (`util/`)
- Scheduled task runners
- Time-based execution utilities
- Web helper functions

## Installation

```bash
go get github.com/wfunc/go
```

## Dependencies

The library uses the following key dependencies:

- **Web Framework**: [gin-gonic/gin](https://github.com/gin-gonic/gin) v1.10.0
- **Redis**: [gomodule/redigo](https://github.com/gomodule/redigo) v1.9.2
- **PostgreSQL**: [jackc/pgx](https://github.com/jackc/pgx) v4
- **Telegram Bot**: [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) v5.5.1
- **Logging**: [uber-go/zap](https://go.uber.org/zap) v1.27.0
- **Captcha**: [dchest/captcha](https://github.com/dchest/captcha) v1.1.0

## Usage Examples

### Session Management

Option A — set global cookie policy defaults once and use simple constructor

```go
package main

import (
    "net/http"
    "os"

    "github.com/Centny/rediscache"
    "github.com/wfunc/go/session"
)

func init() {
    // Configure global cookie policy: HTTP uses Lax + not Secure; HTTPS uses None + Secure
    session.SetDefaultCookiePolicy(session.CookiePolicy{
        SecureOnHTTP:    false,
        SameSiteOnHTTP:  http.SameSiteLaxMode,
        SecureOnHTTPS:   true,
        SameSiteOnHTTPS: http.SameSiteNoneMode,
    })
}

func setup() *session.DbSessionBuilder {
    // Create builder with defaults
    sb := session.NewDbSessionBuilder()
    // Provide Redis connection factory (using rediscache for example)
    redisURI := os.Getenv("REDIS_URI")
    if redisURI == "" { redisURI = "redis.loc:6379?db=1" }
    rediscache.InitRedisPool(redisURI)
    sb.Redis = rediscache.C
    return sb
}
```

Option B — configure per-instance via constructor options

```go
package main

import (
    "net/http"
    "os"

    "github.com/Centny/rediscache"
    "github.com/wfunc/go/session"
)

func setup() *session.DbSessionBuilder {
    // Customize cookie behavior for this builder only
    sb := session.NewDbSessionBuilder(
        session.WithCookieSecureOnHTTP(false),
        session.WithCookieSameSiteOnHTTP(http.SameSiteLaxMode),
        session.WithCookieSecureOnHTTPS(true),
        session.WithCookieSameSiteOnHTTPS(http.SameSiteNoneMode),
    )
    redisURI := os.Getenv("REDIS_URI")
    if redisURI == "" { redisURI = "redis.loc:6379?db=1" }
    rediscache.InitRedisPool(redisURI)
    sb.Redis = rediscache.C
    return sb
}
```

Attach the builder to your web router/mux (example uses `github.com/wfunc/web`):

```go
import (
    "github.com/wfunc/web"
    "github.com/wfunc/web/httptest"
)

sb := setup()
srv := httptest.NewMuxServer()
srv.Mux.Builder = sb

srv.Mux.HandleFunc("^/set$", func(hs *web.Session) web.Result {
    hs.SetValue("k", "v")
    _ = hs.Flush()
    return hs.Printf("ok")
})
```

### SMS Verification

Register HTTP handlers and provide Redis + sender implementation

```go
package main

import (
    "os"
    "github.com/Centny/rediscache"
    "github.com/wfunc/go/sms"
    "github.com/wfunc/util/xmap"
    "github.com/wfunc/web/httptest"
)

func setupSMS() *httptest.Server {
    // Init Redis pool and wire into sms package
    redisURI := os.Getenv("REDIS_URI")
    if redisURI == "" { redisURI = "redis.loc:6379?db=1" }
    rediscache.InitRedisPool(redisURI)
    sms.UseRedis(rediscache.C)

    // Provide your SMS sender (templateParam["code"]) to call a real provider
    sms.UseSender(func(v *sms.VerifyPhone, phone string, templateParam xmap.M) error {
        // TODO: integrate with provider
        return nil
    })

    srv := httptest.NewMuxServer()
    sms.Hand("", srv.Mux)       // public + user endpoints
    sms.HandDebug("", srv.Mux)   // optional: debug endpoint to read codes
    return srv
}

// Usage
srv := setupSMS()
srv.GetMap("/pub/sendLoginSms?phone=1234567890")
// Debug only (non-production):
srv.GetMap("/pub/loadPhoneCode?key=login&phone=1234567890")
```

Optional provider templates:

- Aliyun SMS template: `examples/providers/aliyun_sms_template.go` (params + signature flow skeleton)
- SMTP email template: `examples/providers/email_smtp_template.go` (configure from env)

### Email Service

Register HTTP handlers and provide Redis + sender implementation

```go
package main

import (
    "os"
    "github.com/Centny/rediscache"
    "github.com/wfunc/go/email"
    "github.com/wfunc/web/httptest"
    "github.com/wfunc/util/xprop"
)

func setupEmail(cfg *xprop.Config) (*email.EmailSender, *httptest.Server, error) {
    // Create SMTP sender from config
    sender, err := email.NewEmailSenderFromConfig(cfg)
    if err != nil { return nil, nil, err }

    // Wire package-level dependencies
    redisURI := os.Getenv("REDIS_URI")
    if redisURI == "" { redisURI = "redis.loc:6379?db=1" }
    rediscache.InitRedisPool(redisURI)
    email.UseRedis(rediscache.C)
    email.UseEmailSender(sender)

    srv := httptest.NewMuxServer()
    email.Hand("", srv.Mux)
    email.HandDebug("", srv.Mux) // optional: debug endpoint to read codes
    return sender, srv, nil
}
```
Optional:
- Configure sender from env: `email.UseEmailSenderFromEnv()`
- See `examples/providers/email_smtp_template.go`.

### Telegram Bot

```go
import "github.com/wfunc/go/bot"

// Initialize bot
if err := bot.Bootstrap(token, chatID); err != nil {
    // handle init error (invalid token/chatID)
    panic(err)
}

// Send message
bot.SendMessage("Hello World!")

// Send HTML message
bot.SendHTMLMessage("<b>Important</b> notification")
```

### Database Operations

Bootstrap pgx, set the Pool, then use helpers

```go
package main

import (
    "context"
    "os"
    "github.com/wfunc/go/basedb"
    "github.com/wfunc/util/xmap"
)

func setupDB() error {
    pgURL := os.Getenv("PG_URL")
    if pgURL == "" { pgURL = "postgresql://dev:123@psql.loc:5432/base" }
    return basedb.Bootstrap(pgURL)
}

func demo(ctx context.Context) error {
    // Store and load config
    _ = basedb.StoreConf(ctx, "site.title", "Welcome")
    var title string
    _ = basedb.LoadConf(ctx, "site.title", &title)

    // Object storage
    _, _ = basedb.UpsertObject(ctx, "profile:uid:123", xmap.M{"name":"Alice"})
    obj, _ := basedb.LoadObject(ctx, "profile:uid:123")

    // Versioned objects
    _ = basedb.UpsertVersionObject(ctx, &basedb.VersionObject{Key:"app", Pub:"web", Value:xmap.M{"v":"1.0.0"}})
    latest, _ := basedb.LoadLatestVersionObject(ctx, "app", "web")
    _, _ = obj, latest
    return nil
}
```

## Project Structure

```
.
├── baseapi/        # Base API handlers and endpoints
├── basedb/         # Database models and utilities
├── baseupgrade/    # Database migration tools
├── bot/            # Telegram bot implementation
│   └── botserver/  # Bot server application
├── captcha/        # Captcha generation and verification
├── define/         # Common definitions and constants
├── email/          # Email service implementation
├── item2md/        # Markdown converter utility
├── session/        # Session management
├── sms/            # SMS service implementation
├── testc/          # Test utilities
├── transport/      # HTTP/WebSocket transport layer
├── upgrade/        # Upgrade utilities
├── util/           # Common utilities
└── xlog/           # Logging configuration
```

## Building and Testing

### Run Tests

```bash
./build.sh
```

The build script will:
1. Build all packages
2. Run tests with coverage
3. Generate coverage reports (JSON, XML, HTML)

Integration prerequisites

- Some packages/tests require external services:
  - PostgreSQL at `psql.loc:5432` (used by basedb/baseapi tests)
  - Redis at `redis.loc:6379` (used by session/sms/email tests)
- Ensure these hostnames resolve to your services (e.g., via `/etc/hosts`) or run compatible containers.
 - Configure via env vars when running examples/tests:
   - `PG_URL` (e.g., `postgresql://dev:123@psql.loc:5432/base`)
   - `REDIS_URI` (e.g., `redis.loc:6379?db=1`)

### Sync Dependencies

```bash
./sync.sh
```

## QuickStart

A minimal runnable example is provided at `examples/quickstart`, covering Session, SMS, Email and simple DB config storage.

Run:

```bash
export REDIS_URI="redis.loc:6379?db=1"
export PG_URL="postgresql://dev:123@psql.loc:5432/base"
cd examples/quickstart && docker compose up -d && cd -
go run ./examples/quickstart               # standard http.Server
go run ./examples/quickstart-httpserver    # standard http.Server (alt)
go run ./examples/quickstart-gin           # Gin integration
```

See `examples/quickstart/README-zh.md` for Chinese instructions.

Examples support .env and flags:

- `.env` keys: `PG_URL`, `REDIS_URI`, `LISTEN_ADDR` (used when env is absent)
- flags: `--listen`, `--pg`, `--redis` (override env)
- helpers provided in `util/env.go`, `util/envload.go`, and `util/config.go`.

### Optional admin panels

- pgAdmin: http://localhost:5050 (default admin admin@admin.com/admin)
- RedisInsight: http://localhost:5540

## Configuration

Most modules support configuration through environment variables or configuration files. Common configurations include:

- Redis connection settings
- Database connection strings
- SMTP server settings
- Telegram bot tokens
- SMS provider credentials

### Environment Helpers

Convenience helpers in `util/env.go`:

- `util.EnvOrDefault(key, def string) string`
- `util.EnvBoolDefault(key string, def bool) bool`
- `util.EnvIntDefault(key string, def int) int`
- `util.EnvPGURL() string` (default `postgresql://dev:123@psql.loc:5432/base`)
- `util.EnvRedisURI() string` (default `redis.loc:6379?db=1`)

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass
2. Code follows Go best practices
3. New features include tests
4. Documentation is updated

## License

Please check with the repository owner for license information.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/wfunc/go).
