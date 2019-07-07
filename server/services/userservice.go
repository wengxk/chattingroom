package services

import (
	"chattingroom/common/infos"
	"chattingroom/common/messages"
	"chattingroom/common/models"
	"chattingroom/common/utils"
	"chattingroom/server/repositories"
	"encoding/json"
	"fmt"
	"net"
)

type UserService struct {
	conn     net.Conn
	userrepo repositories.UserRepository
	User     *models.User
}

func NewUserService(conn net.Conn) (userservice *UserService) {
	userservice = &UserService{
		conn:     conn,
		userrepo: *repositories.NewUserRepository(),
	}
	return
}

// Login 登录消息的业务逻辑
func (this *UserService) Login(message *messages.Message) (user *models.User, err error) {
	resmes := messages.Message{
		Type: messages.LoginResponseMessageType,
	}
	loginresmes := messages.LoginResponseMessage{}
	mf := utils.MessageTransfer{
		Conn: this.conn,
	}

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

	// 序列化要返回的消息
	data, err := json.Marshal(loginresmes)
	if err != nil {
		// loginresmes.Code = 500
		// loginresmes.Error = err.Error()
		// fmt.Println("server user login failed")
		return
	} else {
		resmes.Data = string(data)
		data, err = json.Marshal(resmes)
		if err != nil {
			// loginresmes.Code = 500
			// loginresmes.Error = err.Error()
			// fmt.Println("server user login failed")
			return
		} else {
			// loginresmes.Code = 200
			// fmt.Println("server user login succeed")
		}
	}

	err = mf.SendMessage(data)
	return
}

// Register 注册消息的业务逻辑
func (this *UserService) Register(message *messages.Message) (err error) {
	mes := messages.Message{
		Type: messages.RegistryReponseMessageType,
	}
	sendmes := messages.LoginResponseMessage{}
	mt := utils.MessageTransfer{
		Conn: this.conn,
	}

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
		// sendmes.Code = 500
		// sendmes.Error = err.Error()
		// fmt.Println("server register user faild")
	} else {
		mes.Data = string(data)
		data, err = json.Marshal(mes)
		if err != nil {
			// sendmes.Code = 500
			// sendmes.Error = err.Error()
			// fmt.Println("server register user faild")
		} else {
			// sendmes.Code = 200
		}
	}

	err = mt.SendMessage(data)
	return
}

func (this *UserService) GetOnlineUsers(message *messages.Message) (err error) {

	mes := messages.Message{
		Type: messages.GetOnlineUsersResponseMessageType,
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
		m[k] = v.UserID
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
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println(err)
		resultmes.Code = 500
		resultmes.Error = err.Error()
		return
	}
	mt := &utils.MessageTransfer{
		Conn: this.conn,
	}
	err = mt.SendMessage(data)

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

	mes := messages.Message{
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

	Usermanager.PushServerMessage(&mes)
}
