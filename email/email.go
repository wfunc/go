package email

import (
    "bytes"
    "fmt"
    "math"
    "math/rand"
    "net/smtp"
    "strings"
    "sync"
    "time"
    "os"

	"github.com/wfunc/go/define"
	"github.com/wfunc/go/util"
	"github.com/wfunc/go/xlog"
	"github.com/wfunc/util/xmap"
	"github.com/wfunc/util/xprop"
	"github.com/wfunc/util/xsql"
	"github.com/wfunc/web"

	"github.com/gomodule/redigo/redis"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	//EmailCodeTypeVerify is email code type to verify
	EmailCodeTypeVerify = "verify"
	//EmailCodeTypeLogin is email code type to login
	EmailCodeTypeLogin     = "login"
	VerifyEmailTypeUser    = "user"
	VerifyEmailTypePhone   = "email"
	VerifyEmailTypeCaptcha = "captcha"
)

// Redis will return redis connection
var Redis = func() redis.Conn {
	panic("redis is not initial")
}

// SendEmail will send message to email
var SendEmail = func(v *VerifyEmail, email string, templateParam xmap.M) (err error) {
	panic("send email is not initial")
}

var CaptchaVerify = func(v *VerifyEmail, id, code string) (err error) {
	panic("verify captcha is not initial")
}

// UseRedis 配置 Redis 连接工厂
func UseRedis(f func() redis.Conn) { Redis = f }

// UseSender 配置邮件发送实现
func UseSender(s func(v *VerifyEmail, email string, templateParam xmap.M) error) { SendEmail = s }

// UseEmailSender 直接使用 EmailSender 实例作为发送实现
func UseEmailSender(sender *EmailSender) { SendEmail = sender.SendEmail }

// UseEmailSenderFromEnv 读取环境变量并设置邮件发送实现
func UseEmailSenderFromEnv() error {
    sender, err := NewEmailSenderFromEnv()
    if err != nil { return err }
    UseEmailSender(sender)
    return nil
}

// UseCaptchaVerifier 配置验证码校验逻辑（当 Type = captcha 时）
func UseCaptchaVerifier(f func(v *VerifyEmail, id, code string) error) { CaptchaVerify = f }

// Default code length is 6
var CodeLen = 6

// Default doesn't ignore case
var IgnoreCase = false

// newCode will return code number
func newCode(n int) int {
	if n <= 0 {
		return newCode(6)
	}
	min := int(math.Pow(10, float64(n-1)))
	max := int(math.Pow(10, float64(n)))
	return rand.Intn(max-min) + min
}

type EmailSender struct {
	Username  string
	Passsword string
	SmtpHost  string
	SmtpPort  string
	From      string
	FromName  string
	Title     string
	Body      string
	BodyFile  string
	APIHost   string
}

func NewEmailSenderFromConfig(config *xprop.Config) (sender *EmailSender, err error) {
	sender = &EmailSender{}
	err = config.ValidFormat(`
		email/username,r|s,l:0;
		email/password,r|s,l:0;
		email/smtp_host,r|s,l:0;
		email/smtp_port,r|s,l:0;
		email/from,r|s,l:0;
		email/from_name,r|s,l:0;
		email/title,r|s,l:0;
		email/body,r|s,l:0;
		email/body_file,o|s,l:0;
		email/api_host,o|s,l:0;
	`, &sender.Username, &sender.Passsword, &sender.SmtpHost, &sender.SmtpPort, &sender.From, &sender.FromName, &sender.Title, &sender.Body, &sender.BodyFile, &sender.APIHost)
	return
}

// NewEmailSenderFromEnv 根据环境变量创建 EmailSender
// 需要的环境变量：
//   EMAIL_USERNAME, EMAIL_PASSWORD, EMAIL_SMTP_HOST, EMAIL_SMTP_PORT,
//   EMAIL_FROM, EMAIL_FROM_NAME, EMAIL_TITLE, EMAIL_BODY
func NewEmailSenderFromEnv() (*EmailSender, error) {
    get := func(k string) string { return strings.TrimSpace(os.Getenv(k)) }
    required := []string{"EMAIL_USERNAME","EMAIL_PASSWORD","EMAIL_SMTP_HOST","EMAIL_SMTP_PORT","EMAIL_FROM","EMAIL_FROM_NAME","EMAIL_TITLE","EMAIL_BODY"}
    for _, k := range required {
        if get(k) == "" {
            return nil, fmt.Errorf("missing env: %s", k)
        }
    }
    return &EmailSender{
        Username:  get("EMAIL_USERNAME"),
        Passsword: get("EMAIL_PASSWORD"),
        SmtpHost:  get("EMAIL_SMTP_HOST"),
        SmtpPort:  get("EMAIL_SMTP_PORT"),
        From:      get("EMAIL_FROM"),
        FromName:  get("EMAIL_FROM_NAME"),
        Title:     get("EMAIL_TITLE"),
        Body:      get("EMAIL_BODY"),
        BodyFile:  "",
        APIHost:   "",
    }, nil
}

func (e *EmailSender) SendEmail(v *VerifyEmail, email string, templateParam xmap.M) (err error) {
	// Receiver email address.
	to := []string{
		email,
	}
	// Authentication.
	auth := smtp.PlainAuth("", e.Username, e.Passsword, e.SmtpHost)
	message := bytes.NewBuffer(nil)
	fmt.Fprintf(message, "From: %v\r\n", e.From)
	fmt.Fprintf(message, "To: %v\r\n", e.From)
	fmt.Fprintf(message, "Subject:  %v\r\n\r\n", e.Title)
	fmt.Fprintf(message, "%v\r\n", fmt.Sprintf(e.Body, templateParam.StrDef("", "code")))

	// Sending email.
	err = smtp.SendMail(fmt.Sprintf("%v:%v", e.SmtpHost, e.SmtpPort), auth, e.From, to, message.Bytes())
	return
}

// LoadEmailCode will return send code
func LoadEmailCode(key, email string) (having string, err error) {
	conn := Redis()
	defer conn.Close()
	val, err := conn.Do("get", key+"_email_"+email)
	if err != nil {
		return
	}
	if val != nil {
		having, err = redis.String(val, nil)
	}
	return
}

// VerifyEmail is verify email impl
type VerifyEmail struct {
	Key           string
	Type          string
	UserKey       string
	Limit         int64
	CalledUser    map[string]int64
	CalledUserLck sync.RWMutex
}

// NewVerifyEmail will craete new VerifyEmail
func NewVerifyEmail(key, verifyType string, limit int64) (v *VerifyEmail) {
	v = &VerifyEmail{
		Key:           key,
		Type:          verifyType,
		Limit:         limit,
		CalledUser:    map[string]int64{},
		CalledUserLck: sync.RWMutex{},
	}
	return
}

// SrvHTTP is http handler
func (v *VerifyEmail) SrvHTTP(hs *web.Session) web.Result {
	var email, captchaID, captchaCode string
	err := hs.ValidFormat(`
		email,R|S,P:^.*@.*$;
		captcha_id,O|S,L:0;
		captcha_code,O|S,L:0;
	`, &email, &captchaID, &captchaCode)
	if err != nil {
		return util.ReturnCodeLocalErr(hs, define.ArgsInvalid, "arg-err", err)
	}
	if v.Type == "captcha" {
		err = CaptchaVerify(v, captchaID, captchaCode)
		if err != nil {
			return hs.SendJSON(map[string]any{
				"code":    define.CodeInvalid,
				"message": err.Error(),
			})
		}
	} else {
		unique := ""
		if v.Type == "user" {
			unique = hs.Str(v.UserKey)
		} else if v.Type == "email" {
			unique = email
		} else {
			unique = strings.Split(hs.R.RemoteAddr, ":")[0]
		}
		v.CalledUserLck.Lock()
		now := xsql.TimeNow().Timestamp()
		last := v.CalledUser[unique]
		if now-last < v.Limit {
			v.CalledUserLck.Unlock()
			// return util.ReturnCodeLocalErr(hs, define.Frequently, "srv-err", err)
			return hs.SendJSON(map[string]any{
				"code":    define.Frequently,
				"after":   v.Limit - (now - last),
				"message": "call too frequently",
			})
		}
		v.CalledUser[unique] = now
		v.CalledUserLck.Unlock()
	}
	number := newCode(CodeLen)
	err = SendEmail(v, email, xmap.M{
		"code": number,
	})
	if err != nil {
		xlog.Warnf("VerifyEmail send email by %v fail with %v", email, err)
		return util.ReturnCodeLocalErr(hs, define.ServerError, err.Error(), err)
	}
	conn := Redis()
	defer conn.Close()
	if IgnoreCase {
		email = strings.ToLower(strings.TrimSpace(email))
	}
	_, err = conn.Do("setex", v.Key+"_email_"+email, 1800, number)
	if err != nil {
		xlog.Warnf("VerifyEmail save sened sms by %v fail with %v", email, err)
		return util.ReturnCodeLocalErr(hs, define.ServerError, "srv-err", err)
	}
	return util.ReturnCodeData(hs, 0, "OK")
}

//SendVerifyEmailH is http handler
/**
 *
 * @api {GET} /usr/sendVerifyEmail Send Verify Email
 * @apiName SendVerifyEmail
 * @apiGroup User
 *
 *
 * @apiParam  {String} email the email number
 * @apiParam  {String} [captcha_id] the captcha id
 * @apiParam  {String} [captcha_code] the captcha code
 * @apiSuccess (200) {Number} code the respnose code, see the common define,
 * @apiSuccess (200) {Number} after the after time when call frequently
 *
 * @apiParamExample  {Query} Request-Example:
 * email=xx
 *
 *
 * @apiSuccessExample {JSON} Success-Response:
 * {
 *     "code": 0,
 *     "data": "OK"
 * }
 * @apiSuccessExample {JSON} Frequently-Response:
 * {
 *     "code": 1600,
 *     "after": 10000,
 * }
 *
 */
var SendVerifyEmailH = NewVerifyEmail(EmailCodeTypeVerify, VerifyEmailTypeCaptcha, 10000)

//SendLoginEmailH is http handler
/**
 *
 * @api {GET} /pub/sendLoginEmail Send Login Email
 * @apiName SendLoginEmail
 * @apiGroup User
 *
 *
 * @apiParam  {String} email the email number
 * @apiParam  {String} [captcha_id] the captcha id
 * @apiParam  {String} [captcha_code] the captcha code
 * @apiSuccess (200) {Number} code the respnose code, see the common define,
 * @apiSuccess (200) {Number} after the after time when call frequently
 *
 *
 * @apiParamExample  {Query} Request-Example:
 * email=xx
 *
 *
 * @apiSuccessExample {JSON} Success-Response:
 * {
 *     "code": 0,
 *     "data": "OK"
 * }
 * @apiSuccessExample {JSON} Frequently-Response:
 * {
 *     "code": 1600,
 *     "after": 10000,
 * }
 *
 */
var SendLoginEmailH = NewVerifyEmail(EmailCodeTypeLogin, VerifyEmailTypeCaptcha, 10000)

func LoadEmailCodeH(s *web.Session) web.Result {
	var key, email string
	var err = s.ValidFormat(`
		key,R|S,L:0;
		email,R|S,L:0;
	`, &key, &email)
	if err != nil {
		return util.ReturnCodeLocalErr(s, define.ArgsInvalid, "arg-err", err)
	}
	xlog.Warnf("debug api is enabled for load email code")
	having, err := LoadEmailCode(key, email)
	if err != nil {
		xlog.Warnf("DebugLoadEmailCodeH load %v sended email by %v fail with %v", key, email, err)
		return util.ReturnCodeLocalErr(s, define.ServerError, "srv-err", err)
	}
	xlog.Infof("DebugLoadEmailCodeH load %v sended email by %v is %v", key, email, having)
	return s.SendJSON(map[string]any{
		"code":      0,
		"emailCode": having,
	})
}

func Hand(pre string, mux *web.SessionMux) {
	mux.Handle("^"+pre+"/usr/sendVerifyEmail(\\?.*)?$", SendVerifyEmailH)
	mux.Handle("^"+pre+"/pub/sendLoginEmail(\\?.*)?$", SendLoginEmailH)
}

func HandDebug(pre string, mux *web.SessionMux) {
	xlog.Warnf("debug api is enabled for load email code")
	mux.HandleFunc("^"+pre+"/pub/loadEmailCode(\\?.*)?$", LoadEmailCodeH)
}
