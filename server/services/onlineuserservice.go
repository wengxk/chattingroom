package services

import (
	"chattingroom/common/messages"
	"encoding/json"
	"fmt"
)

var (
	Usermanager *UserOnlineManager
)

// 在线用户管理
type UserOnlineManager struct {
	onlineUsers map[int]*DaemonService
}

func init() {
	Usermanager = &UserOnlineManager{
		onlineUsers: make(map[int]*DaemonService, 10),
	}
}

func (this *UserOnlineManager) Add(onlineUser *DaemonService) {
	this.onlineUsers[onlineUser.UserID] = onlineUser
}

func (this *UserOnlineManager) Remove(userid int) {
	delete(this.onlineUsers, userid)
	return
}

func (this *UserOnlineManager) Get(userid int) (onlineUser *DaemonService, err error) {
	onlineUser, ok := this.onlineUsers[onlineUser.UserID]
	if !ok {
		err = fmt.Errorf("当前用户不在线", userid)
	}
	return
}

func (this *UserOnlineManager) GetAll() (allonlineUsers map[int]*DaemonService) {
	allonlineUsers = this.onlineUsers
	return allonlineUsers
}

func (this *UserOnlineManager) PushServerMessage(message *messages.Message) (err error) {

	switch message.Type {
	case messages.UserStateChangeMessageType:
		{
			go this.pushUserStateChangeMessage(message)
		}
	case messages.ShortMessageSenderMessageType:
		{
			go this.pushShortMessage(message)
		}
	default:
		{
			fmt.Println("not supported message type", message.Type)
		}
	}

	return
}

func (this *UserOnlineManager) pushUserStateChangeMessage(message *messages.Message) (err error) {
	var mes messages.UserStateChangeMessage
	err = json.Unmarshal([]byte(message.Data), &mes)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range this.onlineUsers {
		if mes.UserID == v.UserID {
			continue
		}
		_ = v.PushServerMessage(message)
	}
	return
}

func (this *UserOnlineManager) pushShortMessage(message *messages.Message) (err error) {
	sms := messages.ShortMessageSenderMessage{}
	err = json.Unmarshal([]byte(message.Data), &sms)
	if err != nil {
		fmt.Println(err)
		return
	}

	sendmes := messages.ShortMessageReceiverMessage{
		Scope:    sms.Scope,
		SrcUser:  sms.SrcUser,
		DstUsers: sms.DstUsers,
		Content:  sms.Content,
	}
	data, err := json.Marshal(sendmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes := messages.Message{
		Type: messages.ShortMessageReceiverMessageType,
		Data: string(data),
	}

	if sms.Scope == messages.ToAll {
		for k, v := range this.onlineUsers {
			if k == sms.SrcUser.UserID {
				continue
			}
			v.PushServerMessage(&mes)
		}
	} else if sms.Scope == messages.ToUsers {
		for k, v := range this.onlineUsers {
			if k == sms.SrcUser.UserID {
				continue
			}
			for i := 0; i < len(sms.DstUsers); i++ {
				if v.UserID == sms.DstUsers[i].UserID {
					v.PushServerMessage(&mes)
				}
			}
		}
	} else {

	}

	return
}
