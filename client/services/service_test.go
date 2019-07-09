package services_test

import (
	"chattingroom/client/services"
	"fmt"
	"net"
	"testing"
	"time"
)

// 模拟多用户同时登录和发送消息，当用户达到一定数量时就会出现并发问题
// concurrent map iteration and map write
func TestConcurrency(t *testing.T) {
	fmt.Println("TestConcurrency start")
	chanServices := make(chan *services.UserService, 100)

	for i := 1000; i < 1100; i++ {
		go clientRequest(i, "111111", chanServices)
	}

	time.Sleep(3 * time.Second)

	// 不用发送消息，都会出现map并发读写问题
	// for i := 1000; i < 1100; i++ {
	// 	go clientSendMessage(chanServices)
	// }

	time.Sleep(10 * time.Second)

	for {
		us, ok := <-chanServices
		if !ok {
			return
		}
		us.Conn.Close()
	}
	fmt.Println("TestConcurrency done")
}

func clientRequest(userid int, userpwd string, chanServices chan *services.UserService) {
	conn, err := net.Dial("tcp", "localhost:10001")
	// if err != nil {
	// 	fmt.Println("net dial err", err)
	// 	return
	// }
	// defer func() {
	// 	conn.Close()
	// 	if err := recover(); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }()
	userservice := &services.UserService{
		Conn: conn,
	}
	err = userservice.Login(userid, userpwd)
	if err != nil {
		fmt.Println(err)
	}

	chanServices <- userservice
}

func clientSendMessage(chanServices chan *services.UserService) {
	userService := <-chanServices
	userService.SendShortMessage(1, "", fmt.Sprintln("hello i am", userService.UserID))
}
