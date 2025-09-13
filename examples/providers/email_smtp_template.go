package providers

import (
    "fmt"
    "os"

    "github.com/wfunc/go/email"
)

// SetupSMTPFromEnv 读取常见环境变量并配置 SMTP 邮件发送
// 需要：EMAIL_USERNAME, EMAIL_PASSWORD, EMAIL_SMTP_HOST, EMAIL_SMTP_PORT,
//       EMAIL_FROM, EMAIL_FROM_NAME, EMAIL_TITLE, EMAIL_BODY
func SetupSMTPFromEnv() error {
    if err := email.UseEmailSenderFromEnv(); err != nil {
        return fmt.Errorf("smtp env invalid: %w", err)
    }
    // 可选：忽略大小写
    if os.Getenv("EMAIL_IGNORE_CASE") == "1" {
        email.IgnoreCase = true
    }
    return nil
}

