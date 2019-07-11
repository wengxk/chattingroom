package services_test

import (
	"chattingroom/client/services"
	"chattingroom/common/messages"
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

	for i := 1000; i < 1200; i++ {
		go clientRequest(i, "111111", chanServices)
	}

	time.Sleep(2 * time.Second)

	for {
		select {
		case userService, ok := <-chanServices:
			{
				if !ok {
					fmt.Println("select case break")
					break
				}
				userService.SendShortMessage(messages.ToAll, "", fmt.Sprintln("hello i am", userService.UserID))
			}
		}
	}

	time.Sleep(5 * time.Second)

	// for {
	// 	us, ok := <-chanServices
	// 	if !ok {
	// 		return
	// 	}
	// 	us.Conn.Close()
	// }
	fmt.Println("TestConcurrency done")
}

func clientRequest(userid int, userpwd string, chanServices chan *services.UserService) {
	conn, err := net.Dial("tcp", "localhost:10001")
	userservice := &services.UserService{
		Conn: conn,
	}
	err = userservice.Login(userid, userpwd)
	if err != nil {
		fmt.Println(err)
	}

	chanServices <- userservice
}
