package infos

import (
	"errors"
)

var (
	ERR_USER_NOTEXISTS    = errors.New("用户不存在")
	ERR_USER_EXISTS       = errors.New("用户已经存在")
	ERR_USER_INCORRECTPWD = errors.New("用户密码错误")
)
