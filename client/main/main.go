package main

import (
	"chattingroom/client/services"
	"fmt"
	"net"
	"os"
)

var (
	userservice *services.UserService
)

func main() {

	// 程序开始时即链接到服务器
	conn, err := net.Dial("tcp", "localhost:10001")
	if err != nil {
		fmt.Println("net dial err", err)
		return
	}
	defer func() {
		conn.Close()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	userservice = &services.UserService{
		Conn: conn,
	}

	for {
		var (
			key           int
			userid        int
			username      string
			passwd        string
			passwdconfirm string
		)
		fmt.Println()
		fmt.Println("--------欢迎登陆在线聊天系统-------")
		fmt.Println("--------1. 注册-------------------")
		fmt.Println("--------2. 登录-------------------")
		fmt.Println("--------3. 退出-------------------")
		fmt.Println("--------请选择(1-3)---------------")
		fmt.Scanf("%d\n", &key)

		if key == 1 {
			fmt.Println("请输入用户ID")
			fmt.Scanf("%d\n", &userid)
			fmt.Println("请输入用户名")
			fmt.Scanf("%s\n", &username)
			fmt.Println("请输入用户密码")
			fmt.Scanf("%s\n", &passwd)
			fmt.Println("请再次输入用户密码")
			fmt.Scanf("%s\n", &passwdconfirm)
			if passwd != passwdconfirm {
				fmt.Println("两次输入的用户密码不一致")
				continue
			}
			fmt.Println("注册用户中，请稍等")
			err := userservice.Register(userid, username, passwd)
			if err != nil {
				fmt.Println("注册用户失败", err)
				continue
			}
			fmt.Println("注册用户成功,请登录")
			continue

		} else if key == 2 {
			fmt.Println("请输入用户ID")
			fmt.Scanf("%d\n", &userid)
			fmt.Println("请输入用户密码")
			fmt.Scanf("%s\n", &passwd)
			err := userservice.Login(userid, passwd)
			if err != nil {
				fmt.Println("用户登录失败", err)
				continue
			}

		} else if key == 3 {
			fmt.Println("正在退出")
			os.Exit(0)

		} else {
			fmt.Println("您输入的操作代码有误，请重新输入")

		}
	}
}
