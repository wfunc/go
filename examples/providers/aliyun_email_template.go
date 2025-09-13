package providers

import (
    "fmt"
    "os"
    "strings"

    "github.com/wfunc/go/email"
    "github.com/wfunc/util/xmap"
)

// SetupAliyunEmailFromEnv 示例：使用阿里云邮件（直邮/邮件推送）发送验证码（示例性质）。
// 实际参数与签名流程请参照阿里云最新文档，以下仅展示如何接入 email.UseSender。
// 所需环境变量（示例）：ALIYUN_MAIL_ACCESS_KEY、ALIYUN_MAIL_SECRET、ALIYUN_MAIL_ACCOUNT_NAME、ALIYUN_MAIL_FROM_ALIAS
func SetupAliyunEmailFromEnv() error {
    key := strings.TrimSpace(os.Getenv("ALIYUN_MAIL_ACCESS_KEY"))
    secret := strings.TrimSpace(os.Getenv("ALIYUN_MAIL_SECRET"))
    account := strings.TrimSpace(os.Getenv("ALIYUN_MAIL_ACCOUNT_NAME"))
    fromAlias := strings.TrimSpace(os.Getenv("ALIYUN_MAIL_FROM_ALIAS"))
    if key == "" || secret == "" || account == "" {
        return fmt.Errorf("missing env for aliyun email")
    }
    email.UseSender(func(v *email.VerifyEmail, addr string, param xmap.M) error {
        // TODO: 调用阿里云邮件发送 API，按其签名流程构造请求。
        // 发送内容可使用 param["code"] 作为验证码。
        // 这里只做示例打印，并返回 nil。
        fmt.Printf("[aliyun-email-template] send to %s with code=%v as AccountName=%s FromAlias=%s\n", addr, param["code"], account, fromAlias)
        return nil
    })
    return nil
}

