package services

import (
	"bufio"
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

	"github.com/google/uuid"
)

var (
	MessageSenderChan chan *messages.Message
	ResponseContainer utils.ConcurrentMap
)

type UserService struct {
	Conn   net.Conn
	UserID int
}

func (this *UserService) Login(userid int, passwd string) (err error) {
	// 准备消息
	mes := &messages.Message{
		Type: messages.LoginMessageType,
		UUID: uuid.New().String(),
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

	reschan := make(chan *messages.Message, 1)
	ResponseContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ResponseContainer.Remove(mes.UUID)
	}()

	MessageSenderChan <- mes
	res := <-reschan

	var resultmes messages.LoginResponseMessage
	err = json.Unmarshal([]byte(res.Data), &resultmes)
	if err != nil {
		fmt.Println(err)
	}
	if resultmes.Code == 200 {
		this.UserID = userid
		fmt.Println("登录成功")
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
	mes := &messages.Message{
		Type: messages.RegistryMessageType,
		UUID: uuid.New().String(),
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

	reschan := make(chan *messages.Message, 1)
	ResponseContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ResponseContainer.Remove(mes.UUID)
	}()

	MessageSenderChan <- mes
	res := <-reschan

	var resultmes messages.RegistyResponseMessage
	err = json.Unmarshal([]byte(res.Data), &resultmes)
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
					// fmt.Scanf("%s\n", &content) // 该写法无法处理中间带有空格的输入
					rd := bufio.NewReader(os.Stdin)
					s, err := rd.ReadString('\n')
					if err != nil {
						fmt.Println(err)
						return
					}
					content = string(s)
				}
			case 2:
				{
					scope = messages.ToUsers
					fmt.Println("请输入接收用户的ID,多个用户ID请以英文逗号隔开")
					fmt.Scanf("%s\n", &dstuserstrs)
					fmt.Println("请输入您要发送的内容")
					// fmt.Scanf("%s\n", &content)  // 该写法无法处理中间带有空格的输入
					rd := bufio.NewReader(os.Stdin)
					s, err := rd.ReadString('\n')
					if err != nil {
						fmt.Println(err)
						return
					}
					content = string(s)
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
	mes := &messages.Message{
		Type: messages.GetOnlineUsersMessageType,
		UUID: uuid.New().String(),
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

	reschan := make(chan *messages.Message, 1)
	ResponseContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ResponseContainer.Remove(mes.UUID)
	}()

	MessageSenderChan <- mes
	res := <-reschan

	err = json.Unmarshal([]byte(res.Data), &responsemes)
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
	mes := &messages.Message{
		Type: messages.LogoutMessageType,
		UUID: uuid.New().String(),
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

	// reschan := make(chan *messages.Message, 1)
	// ResponseContainer.Set(mes.UUID, reschan)

	MessageSenderChan <- mes
	// res := <-reschan

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

	mes := &messages.Message{
		Type: messages.ShortMessageSenderMessageType,
		Data: string(data),
		UUID: uuid.New().String(),
	}

	// reschan := make(chan *messages.Message, 1)
	// ResponseContainer.Set(mes.UUID, reschan)

	MessageSenderChan <- mes
	// res := <-reschan

	return nil
}

func (this *UserService) heartBeat() (err error) {
	mes := &messages.Message{
		Type: messages.HeartBeatingMessageType,
		UUID: uuid.New().String(),
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

	ticker := time.NewTicker(60 * time.Second)
	for {
		_, ok := <-ticker.C
		if ok {

			reschan := make(chan *messages.Message, 1)
			ResponseContainer.Set(mes.UUID, reschan)

			MessageSenderChan <- mes
			// res := <-reschan

			//有错误，说明连接有问题，可能需要重连，一般会自动尝试重连几次，若还是失败则会告知用户
			// if err != nil {
			// 	ticker.Stop()
			// 	fmt.Println(err)
			// 	return
			// }
		}
	}

	return
}

func (this *UserService) SequentialSendMessage() {
	MessageSenderChan = make(chan *messages.Message, 10)
	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}
	for {
		mes := <-MessageSenderChan
		err := mt.SendMessage(mes)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (this *UserService) SequentialReceiveMessage() {
	ResponseContainer = utils.NewConcurrentMap()
	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}
	for {
		mes, err := mt.ReceiveMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		reqchan, ok := ResponseContainer.Get(mes.UUID)
		if !ok {
			go this.handleMessageOfNoneRequest(mes)
			// fmt.Println("can not found original request from request container")
			continue
		}
		reqchan <- mes
	}
}

func (this *UserService) handleMessageOfNoneRequest(message *messages.Message) {
	switch message.Type {
	case messages.UserStateChangeMessageType:
		{
			this.handleUserStateChangeMessage(message)
		}
	case messages.ShortMessageReceiverMessageType:
		{
			this.handleShortReceiverMessage(message)
		}
	default:
		{
			fmt.Println("not supported message type", message.Type)
		}
	}
	return
}

func (this *UserService) handleUserStateChangeMessage(message *messages.Message) (err error) {
	var mes messages.UserStateChangeMessage
	err = json.Unmarshal([]byte(message.Data), &mes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if mes.UserState == messages.UserOnline {
		fmt.Println("用户", mes.UserID, "上线")
	} else if mes.UserState == messages.UserOffline {
		fmt.Println("用户", mes.UserID, "退出")
	} else {
	}
	return
}

func (this *UserService) handleShortReceiverMessage(message *messages.Message) (err error) {
	var mes messages.ShortMessageReceiverMessage
	err = json.Unmarshal([]byte(message.Data), &mes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("您收到来自用户", mes.SrcUser.UserID, "的消息:", mes.Content)
	return
}
