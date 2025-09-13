package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/Centny/rediscache"
    "github.com/gin-gonic/gin"
    "github.com/wfunc/go/basedb"
    "github.com/wfunc/go/email"
    "github.com/wfunc/go/session"
    "github.com/wfunc/go/sms"
    "github.com/wfunc/go/util"
    "github.com/wfunc/util/xmap"
    "github.com/wfunc/web"
)

func main() {
	// Session Cookie 策略（HTTP 环境默认保留 Cookie）
	session.SetDefaultCookiePolicy(session.CookiePolicy{
		SecureOnHTTP:    false,
		SameSiteOnHTTP:  http.SameSiteLaxMode,
		SecureOnHTTPS:   true,
		SameSiteOnHTTPS: http.SameSiteNoneMode,
	})

	// Redis
	rediscache.InitRedisPool(util.EnvRedisURI())

	// DB（可选）
	if err := basedb.Bootstrap(util.EnvPGURL()); err != nil {
		fmt.Printf("[quickstart-gin] 警告：初始化数据库失败（可忽略），err=%v\n", err)
	}

    // 创建原生 SessionMux
    mux := web.NewSessionMux("")

	// 构造器
	sb := session.NewDbSessionBuilder()
	sb.Redis = rediscache.C
	mux.Builder = sb

	// 路由（注册到 mux）
	mux.HandleFunc("^/set$", func(hs *web.Session) web.Result {
		hs.SetValue("k", "v")
		_ = hs.Flush()
		return hs.Printf("ok")
	})
	mux.HandleFunc("^/get$", func(hs *web.Session) web.Result {
		return hs.Printf("%v", hs.Value("k"))
	})

	// 短信/邮件
	sms.UseRedis(rediscache.C)
	sms.UseSender(func(v *sms.VerifyPhone, phone string, param xmap.M) error {
		fmt.Printf("[quickstart-gin] 发送短信到 %s, code=%v\n", phone, param["code"])
		return nil
	})
	sms.UseCaptchaVerifier(func(v *sms.VerifyPhone, id, code string) error { return nil })
	sms.Hand("", mux)
	sms.HandDebug("", mux)

	email.UseRedis(rediscache.C)
	email.UseSender(func(v *email.VerifyEmail, addr string, param xmap.M) error {
		fmt.Printf("[quickstart-gin] 发送邮件到 %s, code=%v\n", addr, param["code"])
		return nil
	})
	email.UseCaptchaVerifier(func(v *email.VerifyEmail, id, code string) error { return nil })
	email.Hand("", mux)
	email.HandDebug("", mux)

	// basedb 演示
	mux.HandleFunc("^/conf/set$", func(hs *web.Session) web.Result {
		_ = basedb.StoreConf(context.Background(), "site.title", "欢迎使用 QuickStart Gin")
		return hs.Printf("ok")
	})
	mux.HandleFunc("^/conf/get$", func(hs *web.Session) web.Result {
		var title string
		_ = basedb.LoadConf(context.Background(), "site.title", &title)
		return hs.Printf("%s", title)
	})

	// Gin 集成：将 SessionMux 作为兜底处理（或挂到某个分组）
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 也可以在这里注册 Gin 自身的路由（仅示例）
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	// 将 web.SessionMux 作为统一路由处理（匹配所有路径）
	r.NoRoute(gin.WrapH(mux))

    cfg := util.LoadAppConfig()
    fmt.Printf("Listening on %s (gin)\n", cfg.Listen)
    if err := r.Run(cfg.Listen); err != nil {
        panic(err)
    }
}
