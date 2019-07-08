package repositories

import (
	"chattingroom/common/models"
	"encoding/json"
	"fmt"
	"strconv"
)

// Seed 初始化1000个用户
func Seed() (err error) {
	conn, err := redispool.Dial()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 1000; i <= 1999; i++ {
		user := models.User{
			UserID:   i,
			UserPwd:  "111111",
			UserName: strconv.Itoa(i),
		}
		data, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
			return err
		}

		reply, err := conn.Do("hset", "users", user.UserID, string(data))
		if err != nil {
			return err
		}
		fmt.Println(reply)
	}
	return
}
