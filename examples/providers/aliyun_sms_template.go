package providers

import (
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/wfunc/go/sms"
    "github.com/wfunc/util/xmap"
)

// SetupAliyunSmsFromEnv 设置一个阿里云短信的发送模板（示例性质）。
// 所需环境变量：ALIYUN_KEY_ID、ALIYUN_KEY_SECRET、ALIYUN_SIGN_NAME、ALIYUN_TEMPLATE_CODE
// 注意：该示例仅展示参数拼接与签名流程，具体参数与接口可能会更新，请以阿里云官方文档为准。
func SetupAliyunSmsFromEnv() error {
    keyID := strings.TrimSpace(os.Getenv("ALIYUN_KEY_ID"))
    keySecret := strings.TrimSpace(os.Getenv("ALIYUN_KEY_SECRET"))
    signName := strings.TrimSpace(os.Getenv("ALIYUN_SIGN_NAME"))
    templateCode := strings.TrimSpace(os.Getenv("ALIYUN_TEMPLATE_CODE"))
    if keyID == "" || keySecret == "" || signName == "" || templateCode == "" {
        return fmt.Errorf("missing env for aliyun sms")
    }
    endpoint := "https://dysmsapi.aliyuncs.com/" // 阿里云短信 API Endpoint

    sms.UseSender(func(v *sms.VerifyPhone, phone string, param xmap.M) error {
        // 1) 基础参数
        qs := url.Values{}
        qs.Set("Action", "SendSms")
        qs.Set("Version", "2017-05-25")
        qs.Set("RegionId", "cn-hangzhou")
        qs.Set("AccessKeyId", keyID)
        qs.Set("Format", "JSON")
        qs.Set("SignatureMethod", "HMAC-SHA1")
        qs.Set("SignatureNonce", fmt.Sprintf("%d", time.Now().UnixNano()))
        qs.Set("SignatureVersion", "1.0")
        qs.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
        qs.Set("PhoneNumbers", phone)
        qs.Set("SignName", signName)
        qs.Set("TemplateCode", templateCode)
        // 传验证码到模板参数（模板需要形如 {"code":"123456"}）
        code := fmt.Sprintf("%v", param["code"])
        qs.Set("TemplateParam", fmt.Sprintf("{\"code\":\"%s\"}", code))

        // 2) 计算签名
        // 参考官方签名步骤：对排序后的查询串进行编码，拼接 "GET&%2F&<encodedQuery>", 使用 HMAC-SHA1 + base64
        canonical := canonicalizedQuery(qs)
        stringToSign := "GET&%2F&" + url.QueryEscape(canonical)
        sig := signHMACSHA1(stringToSign, keySecret+"&")
        qs.Set("Signature", sig)

        // 3) 发送请求
        reqURL := endpoint + "?" + qs.Encode()
        // 实际生产建议使用 http.Client 并处理响应，这里仅演示
        resp, err := http.Get(reqURL)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        if resp.StatusCode/100 != 2 {
            return fmt.Errorf("aliyun sms http %d", resp.StatusCode)
        }
        return nil
    })
    return nil
}

func canonicalizedQuery(v url.Values) string {
    // 需要对 key/value 进行 RFC 3986 编码后再按 key 排序
    keys := make([]string, 0, len(v))
    for k := range v { keys = append(keys, k) }
    sort.Strings(keys)
    parts := make([]string, 0, len(keys))
    for _, k := range keys {
        parts = append(parts, percentEncode(k)+"="+percentEncode(v.Get(k)))
    }
    return strings.Join(parts, "&")
}

func percentEncode(s string) string {
    // 阿里云要求空格转 %20，* 转 %2A，~ 保留
    return strings.NewReplacer("+","%20","*","%2A","%7E","~").Replace(url.QueryEscape(s))
}

func signHMACSHA1(s, key string) string {
    mac := hmac.New(sha1.New, []byte(key))
    mac.Write([]byte(s))
    return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

