package services

import (
	"chattingroom/common/messages"
	"chattingroom/common/models"
	"chattingroom/common/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type UserService struct {
	Conn          net.Conn
	UserID        int
	daemonService *DaemonService
}

func (this *UserService) Login(userid int, passwd string) (err error) {
	// 先发消息长度，再发具体消息
	// 准备消息
	mes := messages.Message{
		Type: messages.LoginMessageType,
	}
	// 准备登录消息
	sendmes := messages.LoginMessage{
		UserID:  userid,
		UserPwd: passwd,
	}
	// 序列化登录消息
	data, err := json.Marshal(sendmes)
	if err != nil {
		fmt.Println("json marshal err", err)
		return err
	}
	mes.Data = string(data)
	// 序列化要传输的消息
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json marshal err", err)
		return err
	}

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}
	err = mt.SendMessage(data)
	if err != nil {
		fmt.Println("conn write message err", err)
	}
	mes, err = mt.ReceiveMessage()
	if err != nil {
		fmt.Println("conn read message err", err)
	}

	var resultmes messages.LoginResponseMessage
	err = json.Unmarshal([]byte(mes.Data), &resultmes)
	if err != nil {
		fmt.Println(err)
	}
	if resultmes.Code == 200 {
		fmt.Println("登录成功")
		// 创建一个守护协程，用于接收服务器主动推送的消息
		this.daemonService = &DaemonService{}
		daemonConn, err := net.Dial("tcp", "localhost:10001")
		if err != nil {
			fmt.Println("create deamon conn failed", err)
		} else {
			this.UserID = userid
			this.daemonService.Conn = daemonConn
			this.daemonService.UserID = userid
			go this.daemonService.ProcessServerMessage()
		}
		// 在运行service_test.go时需要注释
		for {
			this.showMenus()
		}
	} else {
		err = errors.New(resultmes.Error)
	}
	return
}

func (this *UserService) Register(userid int, username string, userpwd string) (err error) {
	user := &models.User{
		UserID:   userid,
		UserName: username,
		UserPwd:  userpwd,
	}
	err = this.validateRegisterUser(user)
	if err != nil {
		return
	}
	//准备消息
	mes := messages.Message{
		Type: messages.RegistryMessageType,
	}
	sendmes := messages.RegistryMessage{
		User: *user,
	}
	data, err := json.Marshal(sendmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println(err)
		return
	}

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}
	err = mt.SendMessage(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes, err = mt.ReceiveMessage()
	if err != nil {
		fmt.Println(err)
	}

	var resultmes messages.RegistyResponseMessage
	err = json.Unmarshal([]byte(mes.Data), &resultmes)
	if err != nil {
		fmt.Println(err)
	}
	if resultmes.Code == 200 {
		fmt.Println("注册成功")
	} else {
		err = errors.New(resultmes.Error)
	}

	return
}

func (this *UserService) validateRegisterUser(user *models.User) (err error) {
	if user.UserID == 0 {
		return errors.New("用户ID不能为0")
	}
	if user.UserPwd == "" {
		return errors.New("用户密码不能为空")
	}
	if user.UserName == "" {
		user.UserName = strconv.Itoa(user.UserID)
	}
	return
}

// 显示用户登录后的交互菜单
func (this *UserService) showMenus() {
	fmt.Println()
	fmt.Println("--------恭喜登录成功-----------")
	fmt.Println("--------1. 显示在线用户--------")
	fmt.Println("--------2. 发送聊天消息--------")
	fmt.Println("--------3. 退出系统------------")
	fmt.Println("请选择(1-3)")

	var key int
	fmt.Scanf("%d\n", &key)
	switch key {
	case 1:
		{
			fmt.Println("显示在线用户")
			users, err := this.GetOnlineUsers()
			if err != nil {
				fmt.Println(err)
				return
			}
			for k, v := range users {
				fmt.Println(k, v)
			}

		}
	case 2:
		{
			var scope int
			var dstuserstrs string
			var content string
			fmt.Println("您希望发送给:")
			fmt.Println("1. 全部在线用户")
			fmt.Println("2. 某些用户")
			fmt.Println("请选择(1-2)")
			fmt.Scanf("%d\n", &key)
			switch key {
			case 1:
				{
					scope = messages.ToAll
					fmt.Println("请输入您要发送的内容")
					fmt.Scanf("%s\n", &content)

				}
			case 2:
				{
					scope = messages.ToUsers
					fmt.Println("请输入接收用户的ID,多个用户ID请以英文逗号隔开")
					fmt.Scanf("%s\n", &dstuserstrs)
					fmt.Println("请输入您要发送的内容")
					fmt.Scanf("%s\n", &content)

				}
			default:
				fmt.Println("您选择的操作代码有误,请重新输入")
				return

			}
			err := this.SendShortMessage(scope, dstuserstrs, content)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	case 3:
		{
			fmt.Println("正在退出")
			this.Logout()
			os.Exit(0)

		}
	default:
		fmt.Println("选择操作代码有误，请重新选择")

	}
}

func (this *UserService) GetOnlineUsers() (users map[int]int, err error) {
	mes := messages.Message{
		Type: messages.GetOnlineUsersMessageType,
	}
	requestmes := messages.GetOnlineUsersMessage{}
	responsemes := messages.GetOnlineUsersResponseMessage{}

	requestmes.UserID = this.UserID
	data, err := json.Marshal(requestmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println(err)
		return
	}

	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}
	err = mt.SendMessage(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes, err = mt.ReceiveMessage()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal([]byte(mes.Data), &responsemes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if responsemes.Code == 500 {
		return nil, errors.New(responsemes.Error)
	}
	return responsemes.OnlineUsers, nil
}

func (this *UserService) Logout() {
	mes := messages.Message{
		Type: messages.LogoutMessageType,
	}
	requesetmes := messages.LogoutMessage{
		UserID: this.UserID,
	}
	data, err := json.Marshal(requesetmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println(err)
		return
	}

	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}
	mt.SendMessage([]byte(data))
	return
}

func (this *UserService) SendShortMessage(scope int, dstuserstrs string, content string) (err error) {
	sms := messages.ShortMessageSenderMessage{
		Scope: scope,
		SrcUser: models.User{
			UserID: this.UserID,
		},
		Content: content,
	}
	if scope == messages.ToUsers {
		dstusers := make([]models.User, 5)
		ints := strings.Split(dstuserstrs, ",")
		for i := 0; i < len(ints); i++ {
			dtsuserid, err := strconv.Atoi(ints[i])
			if err != nil {
				fmt.Println(err)
				return err
			}
			dstusr := models.User{
				UserID: dtsuserid,
			}
			dstusers = append(dstusers, dstusr)
		}
		sms.DstUsers = dstusers
	}

	data, err := json.Marshal(sms)
	if err != nil {
		fmt.Println(err)
		return err
	}

	mes := messages.Message{
		Type: messages.ShortMessageSenderMessageType,
		Data: string(data),
	}
	data, err = json.Marshal(mes)
	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}

	err = mt.SendMessage(data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (this *UserService) HeartBeat() (err error) {
	mes := messages.Message{
		Type: messages.HeartBeatingMessageType,
	}
	heart := messages.HeartBeatingMessage{
		UserID: this.UserID,
	}
	data, err := json.Marshal(heart)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println(err)
		return
	}

	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}

	ticker := time.NewTicker(60 * time.Second)
	for {
		_, ok := <-ticker.C
		if ok {
			err = mt.SendMessage(data)
			_, err = mt.ReceiveMessage()
			//有错误，说明连接有问题，可能需要重连，一般会自动尝试重连几次，若还是失败则会告知用户
			if err != nil {
				ticker.Stop()
				fmt.Println(err)
				return
			}
		}
	}

	return
}
