package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var RdbLikeUserId *redis.Client

var RdbLikeVideoId *redis.Client

var RdbVCid *redis.Client
var RdbCVid *redis.Client

func InitRedis() {
	RdbLikeUserId = redis.NewClient(&redis.Options{
		Addr: "192.168.31.226:6379",
		DB:   0,
	})

	RdbLikeVideoId = redis.NewClient(&redis.Options{
		Addr: "192.168.31.226:6379",
		DB:   1,
	})

	RdbVCid = redis.NewClient(&redis.Options{
		Addr: "192.168.31.226:6379",
		DB:   2,
	})

	RdbCVid = redis.NewClient(&redis.Options{
		Addr: "192.168.31.226:6379",
		DB:   3,
	})

}
