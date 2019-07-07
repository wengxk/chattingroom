package repositories

import (
	"chattingroom/common/infos"
	"chattingroom/common/models"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type UserRepository struct {
	pool *redis.Pool
}

func NewUserRepository() (repo *UserRepository) {
	repo = &UserRepository{
		pool: redispool,
	}
	return
}

func (this *UserRepository) GetUserByID(id int) (user *models.User, err error) {
	conn, err := this.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	r, err := redis.String(conn.Do("hget", "users", id))
	if err == redis.ErrNil {
		err = infos.ERR_USER_NOTEXISTS
		fmt.Println(err)
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal([]byte(r), &user)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (this *UserRepository) Update(user *models.User) (err error) {
	conn, err := this.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	u, err := this.GetUserByID(user.UserID)
	if u == nil {
		return infos.ERR_USER_NOTEXISTS
	}
	if err != nil {
		return err
	}

	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	_, err = conn.Do("hset", "users", user.UserID, string(data))
	if err != nil {
		return
	}
	return
}

func (this *UserRepository) Delete(user *models.User) (err error) {
	// TO-DO
	return
}

func (this *UserRepository) Add(user *models.User) (err error) {
	conn, err := this.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	u, err := this.GetUserByID(user.UserID)
	if u != nil {
		return infos.ERR_USER_EXISTS
	}

	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	_, err = conn.Do("hset", "users", user.UserID, string(data))
	if err != nil {
		return
	}
	return
}
