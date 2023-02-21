package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var RdbLikeUserId *redis.Client

var RdbLikeVideoId *redis.Client

func InitRedis() {
	RdbLikeUserId = redis.NewClient(&redis.Options{
		Addr: "192.168.200.130",
		DB:   0,
	})

	RdbLikeVideoId = redis.NewClient(&redis.Options{
		Addr: "192.168.200.130",
		DB:   0,
	})

}
