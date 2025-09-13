package basedb

import (
    "context"
    "math/rand"
    "strings"

    "github.com/wfunc/crud/pgx"
    "github.com/wfunc/go/baseupgrade"
    "github.com/wfunc/util/xtime"
)

func init() {
	rand.Seed(xtime.Now())
}

var SYS = "_sys"

// Pool will return database connection pool
var Pool = func() *pgx.PgQueryer {
	panic("db is not initial")
}

// CheckDb will check database if is initial
func CheckDb() (created bool, err error) {
	_, _, err = Pool().Exec(context.Background(), `select key from `+SYS+`_config limit 1`)
	if err != nil {
		_, _, err = Pool().Exec(context.Background(), strings.ReplaceAll(baseupgrade.LATEST, "_sys_", SYS+"_"))
		created = true
	}
	return
}

type BaseTableName string

func (b BaseTableName) GetTableName(args ...any) string {
	if len(b) > 0 && len(args) < 2 {
		return SYS + "_" + string(b)
	}
	return SYS + "_" + args[1].(string)
}

// Bootstrap 使用连接字符串初始化 pgx 连接池，设置 basedb.Pool，并检查/初始化数据库结构。
// 便于在应用启动时一行完成数据库准备。
func Bootstrap(connURL string) error {
    if _, err := pgx.Bootstrap(connURL); err != nil {
        return err
    }
    Pool = pgx.Pool
    _, err := CheckDb()
    return err
}
