package providers

import (
    "errors"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/wfunc/go/sms"
    "github.com/wfunc/util/xmap"
)

// SetupTencentSmsFromEnv 示例：腾讯云短信发送模板（TC3-HMAC-SHA256 签名流程较复杂，需参照官方文档完成）
// 需要：TENCENT_SECRET_ID、TENCENT_SECRET_KEY、TENCENT_SMS_SDK_APP_ID、TENCENT_SIGN_NAME、TENCENT_TEMPLATE_ID
// 注意：本模板仅为结构示例，实际签名与请求需严格按官方文档实现。
func SetupTencentSmsFromEnv() error {
    sid := os.Getenv("TENCENT_SECRET_ID")
    skey := os.Getenv("TENCENT_SECRET_KEY")
    appid := os.Getenv("TENCENT_SMS_SDK_APP_ID")
    sign := os.Getenv("TENCENT_SIGN_NAME")
    tpl := os.Getenv("TENCENT_TEMPLATE_ID")
    if sid == "" || skey == "" || appid == "" || sign == "" || tpl == "" {
        return errors.New("missing env for tencent sms")
    }
    endpoint := "https://sms.tencentcloudapi.com" // 参考地区与域名

    sms.UseSender(func(v *sms.VerifyPhone, phone string, param xmap.M) error {
        // TODO: 根据官方文档生成 TC3-HMAC-SHA256 Authorization 与必要头部
        // 文档参考：https://cloud.tencent.com/document/product/382/38778
        // 下面仅作占位示例，真实实现需使用 canonical request/signature 过程。
        req, _ := http.NewRequest("POST", endpoint, nil)
        req.Header.Set("Content-Type", "application/json; charset=utf-8")
        req.Header.Set("X-TC-Action", "SendSms")
        req.Header.Set("X-TC-Region", "ap-guangzhou")
        req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
        req.Header.Set("X-TC-Version", "2021-01-11")
        // req.Header.Set("Authorization", signTC3(...))

        // 实际生产中需设置请求体，包括 PhoneNumberSet、TemplateId、SignName、TemplateParamSet、SmsSdkAppId 等
        // 并发送请求、检查响应。
        // 此处仅返回未实现错误，避免误用。
        return errors.New("tencent sms not implemented: please implement TC3 signature and request body")
    })
    return nil
}

