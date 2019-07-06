package repositories

import (
	"github.com/gomodule/redigo/redis"
)

var redispool *redis.Pool

// 这里就直接写死初始化参数，不做传参配置
func init() {
	redispool = &redis.Pool{
		MaxIdle:     10,  // 最大空闲连接数
		MaxActive:   0,   //最大连接数
		IdleTimeout: 120, //最大空闲时间
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "192.168.5.139:6379") // 初始化连接
		},
	}
}
