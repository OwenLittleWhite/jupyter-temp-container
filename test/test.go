package test

import (
	"context"
	"fmt"
	"manager/dao/mysql"
	"manager/dao/redis"
	"manager/setting"
	"testing"
)

func Setup(t *testing.T) {
	// 加载配置
	if err := setting.Init("../conf/config_test.yaml"); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		return
	}
	if err := mysql.Init(setting.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
}

func Teardown(t *testing.T) {
	// 获取所有表名
	var tableNames []string
	mysql.Db.Raw("SHOW TABLES").Pluck("Tables_in_your_database", &tableNames)
	// 清空所有表
	for _, tableName := range tableNames {
		mysql.Db.Exec("TRUNCATE TABLE " + tableName)
	}
	mysql.Close()
	redis.RedisClient().FlushAll(context.Background())
	redis.Close()
}
