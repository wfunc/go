package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/Centny/rediscache"
    "github.com/wfunc/go/basedb"
    "github.com/wfunc/go/email"
    "github.com/wfunc/go/session"
    "github.com/wfunc/go/sms"
    "github.com/wfunc/go/util"
    "github.com/wfunc/util/xmap"
    "github.com/wfunc/web"
)

func main() {
	// 1) 全局 Cookie 策略（HTTP 环境下保留 Cookie）
	session.SetDefaultCookiePolicy(session.CookiePolicy{
		SecureOnHTTP:    false,
		SameSiteOnHTTP:  http.SameSiteLaxMode,
		SecureOnHTTPS:   true,
		SameSiteOnHTTPS: http.SameSiteNoneMode,
	})

	// 2) 初始化 Redis（用于 session/sms/email）
	redisURI := util.EnvRedisURI()
	rediscache.InitRedisPool(redisURI)

	// 3) 初始化数据库（可选），依赖 PG_URL
	//    如不想启动数据库，可注释下面两行。
	if err := basedb.Bootstrap(util.EnvPGURL()); err != nil {
		fmt.Printf("[quickstart] 警告：初始化数据库失败（可忽略以仅体验会话/短信/邮件），err=%v\n", err)
	}

	// 4) 创建 session 构造器
	sb := session.NewDbSessionBuilder()
	sb.Redis = rediscache.C

    // 5) 创建原生 SessionMux，并使用标准 http.Server 暴露接口
    mux := web.NewSessionMux("")
    mux.Builder = sb

	// 6) 注册演示路由：/set、/get
    mux.HandleFunc("^/set$", func(hs *web.Session) web.Result {
        hs.SetValue("k", "v")
        _ = hs.Flush()
        return hs.Printf("ok")
    })
    mux.HandleFunc("^/get$", func(hs *web.Session) web.Result {
        return hs.Printf("%v", hs.Value("k"))
    })

	// 7) 配置短信/邮件依赖与路由
	sms.UseRedis(rediscache.C)
	// 演示用发送实现：仅打印日志
	sms.UseSender(func(v *sms.VerifyPhone, phone string, param xmap.M) error {
		fmt.Printf("[quickstart] 发送短信到 %s, code=%v\n", phone, param["code"])
		return nil
	})
	// 验证码类型为 captcha 时的校验逻辑（演示环境直接通过）
	sms.UseCaptchaVerifier(func(v *sms.VerifyPhone, id, code string) error { return nil })
    sms.Hand("", mux)
    sms.HandDebug("", mux)

	// 邮件
	email.UseRedis(rediscache.C)
	email.UseSender(func(v *email.VerifyEmail, addr string, param xmap.M) error {
		fmt.Printf("[quickstart] 发送邮件到 %s, code=%v\n", addr, param["code"])
		return nil
	})
	email.UseCaptchaVerifier(func(v *email.VerifyEmail, id, code string) error { return nil })
    email.Hand("", mux)
    email.HandDebug("", mux)

	// 8) 可选：演示基于 basedb 的简单存取
    mux.HandleFunc("^/conf/set$", func(hs *web.Session) web.Result {
        _ = basedb.StoreConf(context.Background(), "site.title", "欢迎使用 QuickStart")
        return hs.Printf("ok")
    })
    mux.HandleFunc("^/conf/get$", func(hs *web.Session) web.Result {
        var title string
        _ = basedb.LoadConf(context.Background(), "site.title", &title)
        return hs.Printf("%s", title)
    })
    // 9) 启动 HTTP 服务器（统一配置加载）
    cfg := util.LoadAppConfig()
    fmt.Printf("QuickStart 监听：%s\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/set\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/get\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/pub/sendLoginSms?phone=1234567890\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/pub/loadPhoneCode?key=login&phone=1234567890\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/pub/sendLoginEmail?email=demo@example.com\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/pub/loadEmailCode?key=login&email=demo@example.com\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/conf/set\n", cfg.Listen)
    fmt.Printf("- http://127.0.0.1%s/conf/get\n", cfg.Listen)
    if err := http.ListenAndServe(cfg.Listen, mux); err != nil { panic(err) }
}
