package util

import (
	"github.com/codingeasygo/util/xmap"
	"github.com/wfunc/web"
)

func jsonResult(code int, data any, message string, debug string) xmap.M {
	res := make(xmap.M)
	res["code"] = code
	if len(message) > 0 {
		res["message"] = message
	}
	if data != nil {
		res["data"] = data
	}
	if len(debug) > 0 {
		res["debug"] = debug
	}
	return res
}

func ReturnCodeLocalErr(s *web.Session, code int, key string, err error) web.Result {
	return s.SendJSON(jsonResult(code, nil, s.LocalValue(key), err.Error()))
}

func ReturnCodeErr(s *web.Session, code int, err string) web.Result {
	return s.SendJSON(jsonResult(code, nil, err, ""))
}

func ReturnCodeData(s *web.Session, code int, data any) web.Result {
	return s.SendJSON(jsonResult(code, data, "", ""))
}
