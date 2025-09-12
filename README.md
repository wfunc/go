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
    rediscache.InitRedisPool("redis.loc:6379?db=1")
    sb.Redis = rediscache.C
    return sb
}
```

Option B — configure per-instance via constructor options

```go
package main

import (
    "net/http"

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
    rediscache.InitRedisPool("redis.loc:6379?db=1")
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

```go
import "github.com/wfunc/go/sms"

// Send verification SMS
sms.SendVerifySmsH(ctx)

// Verify phone code
result := sms.LoadPhoneCode(phone, code, codeType)
```

### Email Service

```go
import "github.com/wfunc/go/email"

// Create email sender
sender := email.NewEmailSenderFromConfig(config)

// Send verification email
sender.SendEmail(to, subject, body)
```

### Telegram Bot

```go
import "github.com/wfunc/go/bot"

// Initialize bot
bot.Bootstrap(token, chatID)

// Send message
bot.SendMessage("Hello World!")

// Send HTML message
bot.SendHTMLMessage("<b>Important</b> notification")
```

### Database Operations

```go
import "github.com/wfunc/go/basedb"

// Use auto-generated models
obj := basedb.FindObjectByID(id)

// Configuration management
config := basedb.LoadConfig(key)
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

### Sync Dependencies

```bash
./sync.sh
```

## Configuration

Most modules support configuration through environment variables or configuration files. Common configurations include:

- Redis connection settings
- Database connection strings
- SMTP server settings
- Telegram bot tokens
- SMS provider credentials

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
