package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/wfunc/go/bot"
	"github.com/wfunc/util/xmap"
)

func main() {
	bot.Bootstrap(os.Getenv("BOT_TOKEN"), os.Getenv("CHAT_ID"))
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/send", func(c *gin.Context) {
		msg := c.Query("msg")
		if msg == "" {
			c.JSON(400, gin.H{"error": "msg is required"})
			return
		}
		bot.SendMessage(msg)
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
		bot.SendMessage(msg)
		c.JSON(200, gin.H{"status": "ok"})
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	r.Run(addr)
}
