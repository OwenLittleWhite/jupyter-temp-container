package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"manager/setting"
)

// 实际生产环境下 context.Background() 按需替换

var (
	client *redis.Client
	Nil    = redis.Nil
)

// Init 初始化连接
func Init(cfg *setting.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func RedisClient() *redis.Client {
	return client
}

func RunWithLock(keyPrefix string, timeout time.Duration, callback func() (interface{}, error)) (rtn interface{}, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("lock:%s", keyPrefix)
	// 尝试获取锁
	locked, err := client.SetNX(context.Background(), key, "1", timeout).Result()
	if err != nil {
		return
	}

	if !locked {
		fmt.Printf("%s Lock already acquired\n", key)
		return
	}
	fmt.Printf("%s Lock acquired, running business logic...", key)
	defer func() {
		_, _ = client.Del(ctx, key).Result()
	}()
	return callback()
}

func Close() {
	_ = client.Close()
}
