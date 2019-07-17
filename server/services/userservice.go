package services

import (
	"chattingroom/common/infos"
	"chattingroom/common/messages"
	"chattingroom/common/models"
	"chattingroom/server/repositories"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type UserService struct {
	conn              net.Conn
	userrepo          repositories.UserRepository
	User              *models.User
	messageSenderChan chan *messages.Message
}

func NewUserService(conn net.Conn, mschan chan *messages.Message) (userservice *UserService) {
	userservice = &UserService{
		conn:              conn,
		userrepo:          *repositories.NewUserRepository(),
		messageSenderChan: mschan,
	}
	return
}

// Login 登录消息的业务逻辑
func (this *UserService) Login(message *messages.Message) (user *models.User, err error) {
	resmes := &messages.Message{
		Type: messages.LoginResponseMessageType,
		UUID: message.UUID,
	}
	loginresmes := messages.LoginResponseMessage{}

	// 取data，反序列化
	var loginmes messages.LoginMessage
	err = json.Unmarshal([]byte(message.Data), &loginmes)
	if err != nil {
		loginresmes.Code = 500
		loginresmes.Error = "消息错误，反序列失败"
		fmt.Println("server user login failed")
	} else {
		user, err = this.userrepo.GetUserByID(loginmes.UserID)
		if err != nil || user == nil {
			loginresmes.Code = 500
			loginresmes.Error = "用户不存在"
			fmt.Println("server user login failed")
		} else if user != nil && user.UserPwd != loginmes.UserPwd {
			loginresmes.Code = 500
			loginresmes.Error = "密码错误"
			fmt.Println("server user login failed")
		} else {
			loginresmes.Code = 200
			this.User = user
			fmt.Println("server user login succeed")
		}
	}

	data, err := json.Marshal(loginresmes)
	if err != nil {
		return
	}
	resmes.Data = string(data)

	if loginresmes.Code == 200 {

		this.messageSenderChan <- resmes

		// 添加用户到全局在线用户中
		Usermanager.Add(this)

		// 用户登录推送通知
		lpmes := &messages.Message{
			Type: messages.UserStateChangeMessageType,
		}
		usmes := messages.UserStateChangeMessage{
			UserID:    user.UserID,
			UserState: messages.UserOnline,
		}
		data, err = json.Marshal(usmes)
		if err != nil {
			fmt.Println(err)
			return
		}
		lpmes.Data = string(data)
		Usermanager.PushServerMessage(lpmes)
	}

	return
}

// Register 注册消息的业务逻辑
func (this *UserService) Register(message *messages.Message) (err error) {
	mes := &messages.Message{
		Type: messages.RegistryReponseMessageType,
		UUID: message.UUID,
	}
	sendmes := messages.LoginResponseMessage{}

	var rm messages.RegistryMessage
	err = json.Unmarshal([]byte(message.Data), &rm)
	if err != nil {
		sendmes.Code = 500
		sendmes.Error = "消息错误，反序列失败"
		fmt.Println("server register user faild")
	} else {
		err = this.userrepo.Add(&rm.User)
		if err == infos.ERR_USER_EXISTS {
			sendmes.Code = 500
			sendmes.Error = "注册失败，用户已存在"
			fmt.Println("server register user faild")
		} else if err != nil {
			sendmes.Code = 500
			sendmes.Error = err.Error()
			fmt.Println("server register user faild")
		} else {
			sendmes.Code = 200
			fmt.Println("server register user succeed")
		}
	}

	data, err := json.Marshal(sendmes)
	if err != nil {
		return
	} else {
		mes.Data = string(data)
	}

	this.messageSenderChan <- mes

	return
}

func (this *UserService) GetOnlineUsers(message *messages.Message) (err error) {

	mes := &messages.Message{
		Type: messages.GetOnlineUsersResponseMessageType,
		UUID: message.UUID,
	}
	requestmes := messages.GetOnlineUsersMessage{}
	resultmes := messages.GetOnlineUsersResponseMessage{}

	err = json.Unmarshal([]byte(message.Data), &requestmes)
	if err != nil {
		fmt.Println(err)
		resultmes.Code = 500
		resultmes.Error = err.Error()
		return
	}

	users := Usermanager.GetAll()
	m := make(map[int]int)
	for k, v := range users {
		m[k] = v
	}
	resultmes.UserID = requestmes.UserID
	resultmes.OnlineUsers = m
	resultmes.Code = 200

	data, err := json.Marshal(resultmes)
	if err != nil {
		fmt.Println(err)
		resultmes.Code = 500
		resultmes.Error = err.Error()
		return
	}

	mes.Data = string(data)

	this.messageSenderChan <- mes
	return
}

func (this *UserService) Logout(message *messages.Message) {
	var logoutmes messages.LogoutMessage
	err := json.Unmarshal([]byte(message.Data), &logoutmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 从在线用户中移除退出的用户
	Usermanager.Remove(logoutmes.UserID)

	mes := &messages.Message{
		Type: messages.UserStateChangeMessageType,
	}
	spmes := messages.UserStateChangeMessage{
		UserID:    logoutmes.UserID,
		UserState: messages.UserOffline,
	}
	data, err := json.Marshal(spmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)

	Usermanager.PushServerMessage(mes)
}

func (this *UserService) PushServerMessage(message *messages.Message) {
	this.messageSenderChan <- message

}

func (this *UserService) HeartBeat(message *messages.Message) {
	this.conn.SetDeadline(time.Now().Add(120 * time.Second))
}
