# WFunc Go 库

一个全面的 Go 语言库，为构建 Web 应用程序提供可重用组件，包括会话管理、身份验证、消息服务和数据库工具。

## 概述

WFunc Go 是一个模块化的库，旨在通过提供经过实战检验的常用功能组件来加速 Web 应用程序开发，包括：

- 基于 Redis 缓存的会话管理
- 短信和邮件验证系统
- Telegram 机器人集成
- 数据库迁移和 ORM 工具
- HTTP 传输和 WebSocket 支持
- 验证码生成和验证
- 配置管理
- 文件上传处理

## 功能特性

### 核心组件

#### 🔐 **会话管理** (`session/`)
- Redis 支持的会话存储
- 基于 Cookie 的会话处理
- 安全的会话生命周期管理
- 与 Gin Web 框架集成

#### 📱 **短信服务** (`sms/`)
- 手机号码验证
- 短信验证码生成和验证
- 基于 Redis 的验证码存储
- 支持登录和验证码
- 集成验证码以提高安全性

#### 📧 **邮件服务** (`email/`)
- 邮件验证系统
- 登录验证码生成
- SMTP 邮件发送器配置
- 不区分大小写的验证码验证
- 基于 Redis 的验证码缓存

#### 🤖 **Telegram 机器人** (`bot/`)
- Telegram Bot API 集成
- 消息发送工具
- HTML 和 Markdown 消息格式化
- 充值/提现通知模板
- 机器人服务器实现

#### 🗄️ **数据库工具** (`basedb/`)
- PostgreSQL 集成（使用 pgx）
- 自动生成的模型和函数
- 配置管理
- 对象存储模式
- 公告系统

#### 🔄 **数据库迁移** (`baseupgrade/`)
- SQL 迁移管理
- 版本跟踪
- 自动升级脚本
- 数据库架构初始化

#### 🌐 **API 基础** (`baseapi/`)
- RESTful API 处理器
- 配置 API 端点
- 版本对象管理
- 文件上传处理
- 公告系统 API

#### 🔗 **传输层** (`transport/`)
- HTTP 请求转发
- WebSocket 处理器实现
- 代理功能

#### 🛡️ **验证码** (`captcha/`)
- 图片验证码生成
- 验证码验证端点
- 默认配置

#### 📊 **日志** (`xlog/`)
- 使用 zap 的结构化日志
- 生产和开发环境配置
- 日志级别管理

#### 🛠️ **工具集** (`util/`)
- 计划任务运行器
- 基于时间的执行工具
- Web 辅助函数

## 安装

```bash
go get github.com/wfunc/go
```

## 依赖项

该库使用以下主要依赖：

- **Web 框架**: [gin-gonic/gin](https://github.com/gin-gonic/gin) v1.10.0
- **Redis**: [gomodule/redigo](https://github.com/gomodule/redigo) v1.9.2
- **PostgreSQL**: [jackc/pgx](https://github.com/jackc/pgx) v4
- **Telegram 机器人**: [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) v5.5.1
- **日志**: [uber-go/zap](https://go.uber.org/zap) v1.27.0
- **验证码**: [dchest/captcha](https://github.com/dchest/captcha) v1.1.0

## 使用示例

### 会话管理

方式 A — 全局设置 Cookie 策略后使用简单构造器

```go
package main

import (
    "net/http"

    "github.com/Centny/rediscache"
    "github.com/wfunc/go/session"
)

func init() {
    // 全局 Cookie 策略：HTTP 使用 Lax + 非 Secure；HTTPS 使用 None + Secure
    session.SetDefaultCookiePolicy(session.CookiePolicy{
        SecureOnHTTP:    false,
        SameSiteOnHTTP:  http.SameSiteLaxMode,
        SecureOnHTTPS:   true,
        SameSiteOnHTTPS: http.SameSiteNoneMode,
    })
}

func setup() *session.DbSessionBuilder {
    // 默认构造器
    sb := session.NewDbSessionBuilder()
    // 设置 Redis 连接工厂（示例使用 rediscache）
    rediscache.InitRedisPool("redis.loc:6379?db=1")
    sb.Redis = rediscache.C
    return sb
}
```

方式 B — 通过构造器可选项进行实例级配置

```go
package main

import (
    "net/http"

    "github.com/Centny/rediscache"
    "github.com/wfunc/go/session"
)

func setup() *session.DbSessionBuilder {
    // 仅对当前构造器自定义 Cookie 行为
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

将构造器挂载到你的 Web 路由（示例使用 `github.com/wfunc/web`）：

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

### 短信验证

```go
import "github.com/wfunc/go/sms"

// 发送验证短信
sms.SendVerifySmsH(ctx)

// 验证手机验证码
result := sms.LoadPhoneCode(phone, code, codeType)
```

### 邮件服务

```go
import "github.com/wfunc/go/email"

// 创建邮件发送器
sender := email.NewEmailSenderFromConfig(config)

// 发送验证邮件
sender.SendEmail(to, subject, body)
```

### Telegram 机器人

```go
import "github.com/wfunc/go/bot"

// 初始化机器人
bot.Bootstrap(token, chatID)

// 发送消息
bot.SendMessage("Hello World!")

// 发送 HTML 消息
bot.SendHTMLMessage("<b>重要</b>通知")
```

### 数据库操作

```go
import "github.com/wfunc/go/basedb"

// 使用自动生成的模型
obj := basedb.FindObjectByID(id)

// 配置管理
config := basedb.LoadConfig(key)
```

## 项目结构

```
.
├── baseapi/        # 基础 API 处理器和端点
├── basedb/         # 数据库模型和工具
├── baseupgrade/    # 数据库迁移工具
├── bot/            # Telegram 机器人实现
│   └── botserver/  # 机器人服务器应用
├── captcha/        # 验证码生成和验证
├── define/         # 通用定义和常量
├── email/          # 邮件服务实现
├── item2md/        # Markdown 转换工具
├── session/        # 会话管理
├── sms/            # 短信服务实现
├── testc/          # 测试工具
├── transport/      # HTTP/WebSocket 传输层
├── upgrade/        # 升级工具
├── util/           # 通用工具
└── xlog/           # 日志配置
```

## 构建和测试

### 运行测试

```bash
./build.sh
```

构建脚本将：
1. 构建所有包
2. 运行带覆盖率的测试
3. 生成覆盖率报告（JSON、XML、HTML）

### 同步依赖

```bash
./sync.sh
```

## 配置

大多数模块支持通过环境变量或配置文件进行配置。常见配置包括：

- Redis 连接设置
- 数据库连接字符串
- SMTP 服务器设置
- Telegram 机器人令牌
- 短信服务提供商凭据

## 贡献

欢迎贡献！请确保：

1. 所有测试通过
2. 代码遵循 Go 最佳实践
3. 新功能包含测试
4. 更新文档

## 许可证

请向仓库所有者查询许可证信息。

## 支持

如有问题、疑问或想要贡献，请访问 [GitHub 仓库](https://github.com/wfunc/go)。
