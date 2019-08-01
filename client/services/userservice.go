package services

import (
	"bufio"
	"chattingroom/common/messages"
	"chattingroom/common/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// UserService 常用用户服务，包括注册、登录、退出等
type UserService struct {
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
	ConnRWer.ResponseOfRequestContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ConnRWer.ResponseOfRequestContainer.Remove(mes.UUID)
	}()

	ConnRWer.MessageSenderChan <- mes
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
			if !this.showMenus() {
				return
			}
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
	ConnRWer.ResponseOfRequestContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ConnRWer.ResponseOfRequestContainer.Remove(mes.UUID)
	}()

	ConnRWer.MessageSenderChan <- mes
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
func (this *UserService) showMenus() bool {
	fmt.Println()
	fmt.Println("--------恭喜登录成功-----------")
	fmt.Println("--------1. 显示在线用户--------")
	fmt.Println("--------2. 发送聊天消息--------")
	fmt.Println("--------3. 退出登录------------")
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
				return true
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
						return true
					}
					content = string(s)
				}

			case 3:
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
						return true
					}
					content = string(s)
				}
			default:
				fmt.Println("您选择的操作代码有误,请重新输入")
				return true

			}
			err := this.SendShortMessage(scope, dstuserstrs, content)
			if err != nil {
				fmt.Println(err)
				return true
			}
		}
	case 3:
		{
			fmt.Println("正在退出")
			this.Logout()
			return false
		}
	default:
		fmt.Println("选择操作代码有误，请重新选择")
		return true
	}
	return true
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
	ConnRWer.ResponseOfRequestContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ConnRWer.ResponseOfRequestContainer.Remove(mes.UUID)
	}()

	ConnRWer.MessageSenderChan <- mes
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

	reschan := make(chan *messages.Message, 1)
	ConnRWer.ResponseOfRequestContainer.Set(mes.UUID, reschan)
	defer func() {
		close(reschan)
		ConnRWer.ResponseOfRequestContainer.Remove(mes.UUID)
	}()

	ConnRWer.MessageSenderChan <- mes
	_ = <-reschan

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

	ConnRWer.MessageSenderChan <- mes
	// res := <-reschan

	return nil
}
