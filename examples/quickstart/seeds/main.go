package main

import (
	"context"
	"fmt"

	"github.com/wfunc/go/basedb"
	"github.com/wfunc/go/util"
	"github.com/wfunc/util/xmap"
	"github.com/wfunc/util/xsql"
)

func main() {
	// 初始化数据库
	if err := basedb.Bootstrap(util.EnvPGURL()); err != nil {
		panic(err)
	}
	ctx := context.Background()

	// 写入配置示例
	_ = basedb.StoreConf(ctx, "site.title", "WFunc QuickStart")
	_ = basedb.StoreConf(ctx, "site.desc", "一套简单好用的 Go 工具库")

	// 写入对象存储示例
	_, _ = basedb.UpsertObject(ctx, "profile:uid:1001", xmap.M{"name": "Alice", "role": "admin"})
	_, _ = basedb.UpsertObject(ctx, "profile:uid:1002", xmap.M{"name": "Bob", "role": "user"})

	// 写入版本化对象示例
	_ = basedb.UpsertVersionObject(ctx, &basedb.VersionObject{Key: "app", Pub: "web", Value: xsql.M{"version": "1.0.0"}})
	_ = basedb.UpsertVersionObject(ctx, &basedb.VersionObject{Key: "app", Pub: "ios", Value: xsql.M{"version": "1.0.1"}})

	fmt.Println("Seed 数据写入完成！")
}
