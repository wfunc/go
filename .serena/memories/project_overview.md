# WFunc Go Project Overview

## Purpose
WFunc Go is a comprehensive Go library providing reusable components for building web applications with session management, authentication, messaging, and database utilities.

## Tech Stack
- **Language**: Go 1.24.1
- **Web Framework**: Gin (v1.10.0)
- **Database**: PostgreSQL with pgx driver
- **Cache**: Redis with redigo
- **Logging**: Zap (v1.27.0)
- **Bot Integration**: Telegram Bot API (v5.5.1)
- **Captcha**: dchest/captcha

## Main Components
1. **session/** - Redis-backed session management
2. **sms/** - SMS verification service
3. **email/** - Email verification service
4. **bot/** - Telegram bot integration
5. **basedb/** - Database models and utilities
6. **baseapi/** - RESTful API handlers
7. **baseupgrade/** - Database migration tools
8. **transport/** - HTTP/WebSocket transport layer
9. **captcha/** - Captcha generation/verification
10. **xlog/** - Structured logging with zap
11. **util/** - Common utilities and schedulers

## Entry Points
- `bot/botserver/main.go` - Bot server application
- `item2md/item2md.go` - Markdown converter utility
- `testc/testc.go` - Test utility

## Key Features
- Session management with Redis caching
- SMS and Email verification systems
- Telegram bot for notifications
- Database migrations and ORM
- WebSocket support
- Captcha security
- Structured logging