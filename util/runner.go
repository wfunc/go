package util

import (
	"time"

	"github.com/codingeasygo/crud/pgx"
	"github.com/codingeasygo/util/xdebug"
	"github.com/wfunc/go/xlog"
)

// NamedRunner will run call by delay
func NamedRunner(name string, delay time.Duration, running *bool, call func() error) {
	xlog.Infof("%v is starting", name)
	var finishCount = 0
	runCall := func() error {
		defer func() {
			if perr := recover(); perr != nil {
				xlog.Errorf("%v is panic with %v, callstaick is \n%v", perr, xdebug.CallStack())
			}
		}()
		return call()
	}
	for *running {
		err := runCall()
		if err == nil {
			finishCount++
			continue
		}
		if err != pgx.ErrNoRows {
			xlog.Warnf("%v is fail with %v", name, err)
		} else if finishCount > 0 {
			xlog.Debugf("%v is having %v finished", name, finishCount)
		}
		finishCount = 0
		time.Sleep(delay)
	}
	xlog.Warnf("%v is stopped", name)
}

func NamedRunnerWithHMS(name string, hour, minute, second int64, running *bool, call func() error) {
	xlog.Infof("NamedRunnerWithHMS(%v) is starting", name)
	runCall := func() error {
		defer func() {
			if perr := recover(); perr != nil {
				xlog.Errorf("%v is panic with %v, callstaick is \n%v", perr, xdebug.CallStack())
			}
		}()
		return call()
	}
	var finishCount = 0
	first := NextDiff(hour, minute, second)
	xlog.Infof("NamedRunnerWithHMS(%v) first run on %v", name, first)
	time.Sleep(first)
	for *running {
		err := runCall()
		if err == nil || err == pgx.ErrNoRows {
			xlog.Infof("NamedRunnerWithHMS(%v) is having %v finished", name, finishCount)
			nextDiff := NextDiff(hour, minute, second)
			xlog.Infof("NamedRunnerWithHMS(%v) nextDiff %v", name, nextDiff)
			time.Sleep(nextDiff)
			finishCount = 0
		} else {
			xlog.Infof("NamedRunnerWithHMS(%v) is fail with %v", name, err)
			finishCount++
			continue
		}
	}
	xlog.Infof("NamedRunnerWithHMS(%v) is stopped", name)
}

// NamedRunner will run call by delay
func NamedRunnerWithSeconds(name string, seconds int, running *bool, call func() error) {
	xlog.Infof("NamedRunnerWithSeconds(%v) is starting \n", name)
	runCall := func() error {
		defer func() {
			if perr := recover(); perr != nil {
				xlog.Errorf("%v is panic with %v, callstaick is \n%v", perr, xdebug.CallStack())
			}
		}()
		return call()
	}
	var finishCount = 0
	firstDiff := time.Duration(RunPerfectTime(seconds)) * time.Second
	xlog.Infof("NamedRunnerWithSeconds(%v) first run on %v", name, firstDiff)
	time.Sleep(firstDiff)

	for *running {
		err := runCall()
		if err == nil || err == pgx.ErrNoRows {
			xlog.Infof("NamedRunnerWithSeconds(%v) is having %v finished", name, finishCount)
			nextDiff := time.Duration(RunPerfectTime(seconds)) * time.Second
			xlog.Infof("NamedRunnerWithSeconds(%v) nextDiff %v", name, nextDiff)
			time.Sleep(nextDiff)
			finishCount = 0
		} else {
			xlog.Infof("NamedRunnerWithHMS(%v) is fail with %v", name, err)
			finishCount++
			continue
		}
	}
	xlog.Infof("NamedRunnerWithSeconds(%v) is stopped", name)
}

// 指定时间 x时x分x秒
func NextDiff(hour, minute, second int64) (seconds time.Duration) {
	sec := hour*3600 + minute*60 + second
	now := time.Now()
	daySec := now.Hour()*3600 + now.Minute()*60 + now.Second()
	var secTemp int64
	if sec > int64(daySec) {
		secTemp = sec - int64(daySec)
	} else {
		secTemp = sec + 86400 - int64(daySec)
	}
	seconds = time.Duration(secTemp) * time.Second
	return
}

// 指定完整时间 比如 second = 3600 表示每隔整一个小时执行一次
func RunPerfectTime(second int) (restTime int) {
	now := time.Now()
	sec := now.Minute()*60 + now.Second()
	past := sec % second
	restTime = second - past
	return
}
