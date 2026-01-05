package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wfunc/go/bot"
	"github.com/wfunc/util/xmap"
)

// buildTelegramHTML composes an HTML message for Telegram from
// a generic webhook payload, e.g. {service, type, timestamp, data:{message,timestamp}}
func buildTelegramHTML(payload xmap.M, clientIP string) string {
	if payload == nil {
		return ""
	}
    // Extract fields
    service := payload.Str("service")
    typ := payload.Str("type")
    ts := payload.Str("timestamp")

    var message string
    if v, ok := payload["message"].(string); ok && v != "" {
        message = v
    }

    // Common nested fields under data
    var (
        dataPlatform   string
        dataAccount    string
        dataAccountID  string
        dataStatus     string
        dataErrorCode  string
        dataReason     string
    )
    if data, ok := payload["data"].(map[string]any); ok {
        if message == "" {
            if v, ok2 := data["message"].(string); ok2 {
                message = v
            }
        }
        if ts == "" {
            if v, ok2 := data["timestamp"].(string); ok2 {
                ts = v
            }
        }
        if v, ok2 := data["platform"].(string); ok2 {
            dataPlatform = v
        }
        if v, ok2 := data["accountName"].(string); ok2 {
            dataAccount = v
        }
        if v, ok2 := data["accountId"].(string); ok2 {
            dataAccountID = v
        }
        if v, ok2 := data["status"].(string); ok2 {
            dataStatus = v
        }
        if v, ok2 := data["errorCode"].(string); ok2 {
            dataErrorCode = v
        }
        if v, ok2 := data["reason"].(string); ok2 {
            dataReason = v
        }
    }

    // Decide icon by type/status
    icon := "🔔"
    switch strings.ToLower(typ) {
    case "test", "ping":
        icon = "🧪"
    case "alert", "error", "incident", "anomaly", "accountanomaly", "failure", "failed":
        icon = "🚨"
    case "warn", "warning":
        icon = "⚠️"
    case "ok", "info", "notice":
        icon = "ℹ️"
    case "event":
        icon = "📣"
    }
    // If status explicitly indicates error/warn/ok, refine icon
    switch strings.ToLower(strings.TrimSpace(dataStatus)) {
    case "error", "failed", "fail":
        icon = "🚨"
    case "warn", "warning":
        icon = "⚠️"
    case "ok", "success", "passed":
        // keep informational when type is not severe
        if icon == "🔔" || icon == "ℹ️" {
            icon = "✅"
        }
    }

    esc := func(s string) string { return html.EscapeString(strings.TrimSpace(s)) }

    // Platform icon selector
    platformIcon := func(p string) string {
        p = strings.ToLower(strings.TrimSpace(p))
        switch {
        case p == "claude-oauth" || strings.Contains(p, "claude") || strings.Contains(p, "anthropic"):
            return "🤖"
        case p == "openai" || strings.Contains(p, "gpt"):
            return "🤖"
        case strings.Contains(p, "azure-openai"):
            return "☁️"
        case strings.Contains(p, "cloudflare"):
            return "🛡️"
        case strings.Contains(p, "github"):
            return "🐙"
        case strings.Contains(p, "gitlab"):
            return "🦊"
        case strings.Contains(p, "stripe"):
            return "💳"
        case strings.Contains(p, "slack"):
            return "💬"
        case strings.Contains(p, "discord"):
            return "🟣"
        case strings.Contains(p, "feishu") || strings.Contains(p, "lark"):
            return "🪶"
        case strings.Contains(p, "dingtalk") || strings.Contains(p, "dingding"):
            return "🛎️"
        case strings.Contains(p, "wechat") || strings.Contains(p, "wecom"):
            return "💬"
        case strings.Contains(p, "aws"):
            return "☁️"
        case strings.Contains(p, "gcp") || strings.Contains(p, "google cloud"):
            return "☁️"
        case strings.Contains(p, "azure"):
            return "☁️"
        default:
            return "💻"
        }
    }

    // (deprecated) pickPlatformKey removed; platformIcon serves display purpose.

    // Decorate platform value with its icon for display
    if dp := strings.TrimSpace(dataPlatform); dp != "" {
        dataPlatform = strings.TrimSpace(platformIcon(dp) + " " + dp)
    }

    b := &strings.Builder{}
    title := "Webhook 通知"
    if strings.ToLower(typ) == "test" {
        title = "Webhook 测试"
    }
    fmt.Fprintf(b, "%s <b>%s</b>\n", icon, title)
    if service = strings.TrimSpace(service); service != "" {
        fmt.Fprintf(b, "🏷️ 服务: <code>%s</code>\n", esc(service))
    }
    if typ = strings.TrimSpace(typ); typ != "" {
        fmt.Fprintf(b, "📌 类型: <code>%s</code>\n", esc(typ))
    }
    if ts = strings.TrimSpace(ts); ts != "" {
        fmt.Fprintf(b, "⏰ 时间: <code>%s</code>\n", esc(ts))
    }
    if ip := strings.TrimSpace(clientIP); ip != "" {
        fmt.Fprintf(b, "🌐 IP: <code>%s</code>\n", esc(ip))
    }
    // Structured details (if present)
    if dataPlatform = strings.TrimSpace(dataPlatform); dataPlatform != "" {
        fmt.Fprintf(b, "💻 平台: <code>%s</code>\n", esc(dataPlatform))
    }
    if dataAccount = strings.TrimSpace(dataAccount); dataAccount != "" {
        fmt.Fprintf(b, "👤 账号: <code>%s</code>\n", esc(dataAccount))
    }
    if dataAccountID = strings.TrimSpace(dataAccountID); dataAccountID != "" {
        fmt.Fprintf(b, "🆔 账户ID: <code>%s</code>\n", esc(dataAccountID))
    }
    if dataStatus = strings.TrimSpace(dataStatus); dataStatus != "" {
        fmt.Fprintf(b, "⚙️ 状态: <code>%s</code>\n", esc(dataStatus))
    }
    if dataErrorCode = strings.TrimSpace(dataErrorCode); dataErrorCode != "" {
        fmt.Fprintf(b, "📛 错误码: <code>%s</code>\n", esc(dataErrorCode))
    }
    // Platform-specific extra fields
    // Best-effort keys commonly found across platforms; shown when present
    if dataMap, ok := payload["data"].(map[string]any); ok {
        // Candidate keys (union across platforms)
        candidates := []string{
            "resetAt", "retryAfter", "rateLimit",
            "model", "organization", "deployment", "requestId",
            "zone", "rayId",
            "repo", "project", "ref", "workflow", "runId", "pipelineId",
            "customerId", "invoiceId", "subscriptionId", "amount", "currency",
        }
        labels := map[string]string{
            "resetAt":        "重置时间",
            "retryAfter":     "重试等待",
            "rateLimit":      "限流",
            "model":          "模型",
            "organization":   "组织",
            "deployment":     "部署",
            "requestId":      "请求ID",
            "zone":           "区域/Zone",
            "rayId":          "Ray ID",
            "repo":           "仓库",
            "project":        "项目",
            "ref":            "引用/分支",
            "workflow":       "工作流",
            "runId":          "运行ID",
            "pipelineId":     "流水线ID",
            "customerId":     "客户ID",
            "invoiceId":      "发票ID",
            "subscriptionId": "订阅ID",
            "amount":         "金额",
            "currency":       "货币",
        }
        toStr := func(v any) string {
            switch t := v.(type) {
            case string:
                return t
            case fmt.Stringer:
                return t.String()
            case json.Number:
                return t.String()
            default:
                return fmt.Sprint(t)
            }
        }
        for _, k := range candidates {
            if v, ok := dataMap[k]; ok {
                if s := strings.TrimSpace(toStr(v)); s != "" {
                    label := labels[k]
                    if label == "" { label = k }
                    fmt.Fprintf(b, "📎 %s: <code>%s</code>\\n", esc(label), esc(s))
                }
            }
        }
    }
    // Prefer 'reason' as content if present, otherwise 'message'
    content := strings.TrimSpace(dataReason)
    if content == "" {
        content = strings.TrimSpace(message)
    }
    if content != "" {
        fmt.Fprintf(b, "📝 内容:\n<blockquote>%s</blockquote>", esc(content))
    }
    return b.String()
}

func main() {
    if err := bot.Bootstrap(os.Getenv("BOT_TOKEN"), os.Getenv("CHAT_ID")); err != nil {
        // 初始化失败时打印日志，但不中断服务；发送失败会进入重试/回放逻辑
        fmt.Printf("init bot failed: %v\n", err)
    }
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Backlog configuration
	enableBacklog := strings.TrimSpace(os.Getenv("BOT_ENABLE_BACKLOG"))
	if enableBacklog == "" {
		enableBacklog = "1"
	}
	backlogPath := strings.TrimSpace(os.Getenv("BACKLOG_PATH"))
	if backlogPath == "" {
		backlogPath = "data/backlog.jsonl"
	}
	replayIntervalSec := 60
	if v := strings.TrimSpace(os.Getenv("BACKLOG_REPLAY_INTERVAL")); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			replayIntervalSec = i
		}
	}
	adminToken := strings.TrimSpace(os.Getenv("QUEUE_ADMIN_TOKEN"))

	var backlog *Backlog
	if enableBacklog == "1" {
		if bl, err := NewBacklog(backlogPath); err == nil {
			backlog = bl
		} else {
			fmt.Printf("init backlog failed: %v\n", err)
		}
	}

	// Health endpoint for container healthcheck
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Simple send endpoint via query param
	r.GET("/send", func(c *gin.Context) {
		msg := c.Query("msg")
		if msg == "" {
			c.JSON(400, gin.H{"error": "msg is required"})
			return
		}
		if err := sendText(func() (string, int64) { return "", 0 }, msg); err != nil && backlog != nil {
			_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, CreatedAt: time.Now().Unix()})
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	// Webhook endpoint (GET): be lenient so external "测试连接" can pass
	r.GET("/webhook", func(c *gin.Context) {
		msg := c.Query("msg")
		if msg != "" {
			// Include requester IP for visibility when sending plain text
			ip := c.ClientIP()
			if ip != "" {
				msg = fmt.Sprintf("%s (IP: %s)", msg, ip)
			}
			if err := sendText(func() (string, int64) { return "", 0 }, msg); err != nil && backlog != nil {
				_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, CreatedAt: time.Now().Unix()})
			}
		}
		// Always 200 to satisfy generic webhook testers
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/send", func(c *gin.Context) {
		payload := xmap.M{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "invalid payload"})
			return
		}
		msg, ok := payload["msg"].(string)
		if !ok || msg == "" {
			c.JSON(400, gin.H{"error": "msg is required"})
			return
		}

		token := payload.Str("token")
		chatID := payload.Int64("chat_id")
		if len(token) > 0 && chatID > 0 {
			if err := sendText(func() (string, int64) { return token, chatID }, msg); err != nil && backlog != nil {
				_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, Token: token, ChatID: chatID, CreatedAt: time.Now().Unix()})
			}
		} else {
			if err := sendText(func() (string, int64) { return "", 0 }, msg); err != nil && backlog != nil {
				_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, CreatedAt: time.Now().Unix()})
			}
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	// Webhook endpoint (POST): accept diverse payloads from various platforms
	r.POST("/webhook", func(c *gin.Context) {
		payload := xmap.M{}
		// Optional signature verification (HMAC-SHA256 over raw body)
		if secret := strings.TrimSpace(os.Getenv("BOT_WEBHOOK_SECRET")); secret != "" {
			sig := c.GetHeader("X-Signature")
			if !verifySignature(secret, sig, c.Request) {
				c.JSON(401, gin.H{"error": "invalid signature"})
				return
			}
		}
		// Read raw body once so we can attempt JSON and also allow form parsing
		raw, _ := io.ReadAll(c.Request.Body)
		// Restore body for any further parsing by Gin
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		if len(raw) > 0 {
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err == nil {
				payload = xmap.M(m)
			}
		}

		// 打印 webhook payload
		fmt.Printf("Webhook Payload: %+v\n", payload)
		// Best-effort message extraction
		msg := ""
		if payload != nil {
			if v, ok := payload["msg"].(string); ok && v != "" {
				msg = v
			}
			if msg == "" {
				if v, ok := payload["message"].(string); ok && v != "" {
					msg = v
				}
			}
			if msg == "" {
				if v, ok := payload["text"].(string); ok && v != "" {
					msg = v
				}
			}
			if msg == "" {
				if v, ok := payload["content"].(string); ok && v != "" {
					msg = v
				}
			}
			// dingtalk style: {"text": {"content": "..."}}
			if msg == "" {
				if t, ok := payload["text"].(map[string]any); ok {
					if v, ok2 := t["content"].(string); ok2 && v != "" {
						msg = v
					}
				}
			}
			if msg == "" {
				title := payload.Str("title")
				body := payload.Str("body")
				if title != "" || body != "" {
					msg = strings.TrimSpace(title + " " + body)
				}
			}
		}

		// Form or query fallbacks
		if msg == "" {
			if v := c.PostForm("msg"); v != "" {
				msg = v
			} else if v := c.PostForm("message"); v != "" {
				msg = v
			} else if v := c.PostForm("text"); v != "" {
				msg = v
			} else if v := c.PostForm("content"); v != "" {
				msg = v
			} else if v := c.Query("msg"); v != "" {
				msg = v
			}
		}

		// Prefer structured HTML if available; otherwise fallback to plain msg
		formatted := buildTelegramHTML(payload, c.ClientIP())
		token := payload.Str("token")
		chatID := payload.Int64("chat_id")
		if formatted != "" {
			if len(token) > 0 && chatID > 0 {
				if err := sendHTML(func() (string, int64) { return token, chatID }, formatted); err != nil && backlog != nil {
					_ = backlog.Append(BacklogMessage{Type: "html", Content: formatted, Token: token, ChatID: chatID, CreatedAt: time.Now().Unix()})
				}
			} else {
				if err := sendHTML(func() (string, int64) { return "", 0 }, formatted); err != nil && backlog != nil {
					_ = backlog.Append(BacklogMessage{Type: "html", Content: formatted, CreatedAt: time.Now().Unix()})
				}
			}
		} else if msg != "" {
			if ip := c.ClientIP(); ip != "" {
				msg = fmt.Sprintf("%s (IP: %s)", msg, ip)
			}
			if len(token) > 0 && chatID > 0 {
				if err := sendText(func() (string, int64) { return token, chatID }, msg); err != nil && backlog != nil {
					_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, Token: token, ChatID: chatID, CreatedAt: time.Now().Unix()})
				}
			} else {
				if err := sendText(func() (string, int64) { return "", 0 }, msg); err != nil && backlog != nil {
					_ = backlog.Append(BacklogMessage{Type: "text", Content: msg, CreatedAt: time.Now().Unix()})
				}
			}
		}

		// Always return 200 to be compatible with various webhook testers
		c.JSON(200, gin.H{"status": "ok"})
	})
	// Admin endpoints for backlog
	if backlog != nil {
		r.GET("/queue/stats", func(c *gin.Context) {
			if adminToken != "" && c.GetHeader("X-Admin-Token") != adminToken {
				c.JSON(403, gin.H{"error": "forbidden"})
				return
			}
			cnt, err := backlog.Count()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"pending": cnt})
		})
		r.POST("/queue/replay", func(c *gin.Context) {
			if adminToken != "" && c.GetHeader("X-Admin-Token") != adminToken {
				c.JSON(403, gin.H{"error": "forbidden"})
				return
			}
			ok, fail, err := backlog.Replay(
				func(token string, chatID int64, msg string) error {
					if token != "" && chatID > 0 {
						_, e := bot.SendMessageWithBot(token, chatID, msg)
						return e
					}
					_, e := bot.SendMessage(msg)
					return e
				},
				func(token string, chatID int64, msg string) error {
					if token != "" && chatID > 0 {
						_, e := bot.SendHTMLMessageWithBot(token, chatID, msg)
						return e
					}
					_, e := bot.SendHTMLMessage(msg)
					return e
				},
			)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": ok, "fail": fail})
		})
		// periodic replay
		go func() {
			tk := time.NewTicker(time.Duration(replayIntervalSec) * time.Second)
			defer tk.Stop()
			for range tk.C {
				_, _, _ = backlog.Replay(
					func(token string, chatID int64, msg string) error {
						if token != "" && chatID > 0 {
							_, e := bot.SendMessageWithBot(token, chatID, msg)
							return e
						}
						_, e := bot.SendMessage(msg)
						return e
					},
					func(token string, chatID int64, msg string) error {
						if token != "" && chatID > 0 {
							_, e := bot.SendHTMLMessageWithBot(token, chatID, msg)
							return e
						}
						_, e := bot.SendHTMLMessage(msg)
						return e
					},
				)
			}
		}()
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// TLS support via env TLS_CERT_FILE/TLS_KEY_FILE
	cert := strings.TrimSpace(os.Getenv("TLS_CERT_FILE"))
	key := strings.TrimSpace(os.Getenv("TLS_KEY_FILE"))
	if cert != "" && key != "" {
		_ = r.RunTLS(addr, cert, key)
	} else {
		_ = r.Run(addr)
	}
}

// verifySignature checks HMAC-SHA256 in hex of the request body using the given secret.
func verifySignature(secret, headerSig string, r *http.Request) bool {
	// headerSig is expected as lowercase hex of sha256 HMAC
	if headerSig == "" {
		return false
	}
	// read raw body without consuming (handled by caller)
	// The caller already captured raw body to reassign; we verify with that
	// For safety, we re-read now (noop as caller reset it)
	raw, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(raw))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(headerSig)))
}

// sendText sends a text message with retry (3 attempts)
// If token/chatID provider returns empty token/chatID, uses default bot.
func sendText(tokenChat func() (string, int64), msg string) error {
	token, chat := tokenChat()
	return retry(3, 500*time.Millisecond, func() error {
		if token != "" && chat > 0 {
			_, err := bot.SendMessageWithBot(token, chat, msg)
			return err
		}
		_, err := bot.SendMessage(msg)
		return err
	})
}

// sendHTML sends an HTML message with retry (3 attempts)
func sendHTML(tokenChat func() (string, int64), htmlMsg string) error {
	token, chat := tokenChat()
	return retry(3, 500*time.Millisecond, func() error {
		if token != "" && chat > 0 {
			_, err := bot.SendHTMLMessageWithBot(token, chat, htmlMsg)
			return err
		}
		_, err := bot.SendHTMLMessage(htmlMsg)
		return err
	})
}

func retry(times int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < times; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	fmt.Printf("send retry failed after %d attempts: %v\n", times, err)
	return err
}
