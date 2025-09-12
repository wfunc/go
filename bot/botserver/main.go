package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

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
    }

    // Decide icon by type
    icon := "🔔"
    switch strings.ToLower(typ) {
    case "test", "ping":
        icon = "🧪"
    case "alert", "error", "incident":
        icon = "🚨"
    case "warn", "warning":
        icon = "⚠️"
    case "ok", "info", "notice":
        icon = "ℹ️"
    case "event":
        icon = "📣"
    }

    esc := func(s string) string { return html.EscapeString(strings.TrimSpace(s)) }

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
    if message = strings.TrimSpace(message); message != "" {
        fmt.Fprintf(b, "📝 内容:\n<blockquote>%s</blockquote>", esc(message))
    }
    return b.String()
}

func main() {
	bot.Bootstrap(os.Getenv("BOT_TOKEN"), os.Getenv("CHAT_ID"))
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// Simple send endpoint via query param
	r.GET("/send", func(c *gin.Context) {
		msg := c.Query("msg")
		if msg == "" {
			c.JSON(400, gin.H{"error": "msg is required"})
			return
		}
		bot.SendMessage(msg)
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
            bot.SendMessage(msg)
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
			bot.SendMessageWithBot(token, chatID, msg)
		} else {
			bot.SendMessage(msg)
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	// Webhook endpoint (POST): accept diverse payloads from various platforms
	r.POST("/webhook", func(c *gin.Context) {
		payload := xmap.M{}
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
                bot.SendHTMLMessageWithBot(token, chatID, formatted)
            } else {
                bot.SendHTMLMessage(formatted)
            }
        } else if msg != "" {
            if ip := c.ClientIP(); ip != "" {
                msg = fmt.Sprintf("%s (IP: %s)", msg, ip)
            }
            if len(token) > 0 && chatID > 0 {
                bot.SendMessageWithBot(token, chatID, msg)
            } else {
                bot.SendMessage(msg)
            }
        }

		// Always return 200 to be compatible with various webhook testers
		c.JSON(200, gin.H{"status": "ok"})
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	r.Run(addr)
}
