package service

import (
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/middleware/redis"
	"testing"
)

func TestIsFavourite(t *testing.T) {
	// 初始化数据库
	dao.InitDB()

	// 初始化redis-DB0的连接，follow选择的DB0.
	redis.InitRedis()

}
