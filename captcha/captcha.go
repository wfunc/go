package captcha

import (
	"github.com/dchest/captcha"
	"github.com/wfunc/go/define"
	"github.com/wfunc/go/util"
	"github.com/wfunc/util/xmap"
	"github.com/wfunc/web"
)

var DefaultLen = 6

func CaptchaVerify(id, code string) (err error) {
	if !captcha.VerifyString(id, code) {
		err = define.ErrCodeInvalid
	}
	return
}

//NewCaptchaH is http handler
/**
 *
 * @api {GET} /pub/captcha/new New Captcha
 * @apiName NewCaptcha
 * @apiGroup Captcha
 *
 * @apiSuccess (200) {Number} code the respnose code, see the common define,
 * @apiSuccess (200) {String} captcha_id the captcha image id, to load image by /pub/captcha/${captcha_id}.png
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
func NewCaptchaH(s *web.Session) web.Result {
	return s.SendJSON(xmap.M{
		"code":       define.Success,
		"captcha_id": captcha.NewLen(DefaultLen),
	})
}

//VerifyCaptchaH is http handler
/**
 *
 * @api {GET} /pub/captcha/verify Verify Captcha
 * @apiName VerifyCaptcha
 * @apiGroup Captcha
 *
 * @apiParam  {String} captcha_id the captcha id
 * @apiParam  {String} captcha_code the captcha code
 *
 * @apiSuccess (200) {Number} code the respnose code, see the common define,
 *
 *
 * @apiParamExample  {Query} Request-Example:
 * captcha_id=xx&captcha_code=1234
 *
 *
 * @apiSuccessExample {JSON} Success-Response:
 * {
 *     "code": 0
 * }
 * @apiSuccessExample {JSON} Frequently-Response:
 * {
 *     "code": 1600,
 *     "after": 10000,
 * }
 *
 */
func VerifyCaptchaH(s *web.Session) web.Result {
	var captchaID, captchaCode string
	err := s.ValidFormat(`
		captcha_id,R|S,L:0;
		captcha_code,R|S,L:0;
	`, &captchaID, &captchaCode)
	if err != nil {
		return util.ReturnCodeLocalErr(s, define.ArgsInvalid, "arg-err", err)
	}
	if !captcha.VerifyString(captchaID, captchaCode) {
		return util.ReturnCodeLocalErr(s, define.CodeInvalid, "arg-err", define.ErrCodeInvalid)
	}
	return s.SendJSON(xmap.M{
		"code": define.Success,
	})
}

func Hand(pre string, mux *web.SessionMux) {
	mux.HandleFunc("^"+pre+"/pub/captcha/new(\\?.*)?$", NewCaptchaH)
	mux.HandleFunc("^"+pre+"/pub/captcha/verify(\\?.*)?$", VerifyCaptchaH)
	mux.HandleNormal("^"+pre+"/pub/captcha/.*$", captcha.Server(captcha.StdWidth, captcha.StdHeight))
}
