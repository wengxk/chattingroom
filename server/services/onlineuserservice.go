package services

import (
	"chattingroom/common/messages"
	"encoding/json"
	"fmt"
	"sync"
)

var (
	// Usermanager 全局，在线用户管理
	Usermanager *UserOnlineManager
)

// UserOnlineManager 在线用户管理
type UserOnlineManager struct {
	// onlineUsers map[int]*DaemonService
	onlineUsers sync.Map
}

func init() {
	var smap sync.Map
	Usermanager = &UserOnlineManager{
		onlineUsers: smap,
	}
}

func (this *UserOnlineManager) Add(onlineUser *UserService) {
	this.onlineUsers.Store(onlineUser.User.UserID, onlineUser)
}

func (this *UserOnlineManager) Remove(userid int) {
	this.onlineUsers.Delete(userid)
	return
}

func (this *UserOnlineManager) Get(userid int) (onlineUser *UserService, err error) {
	u, ok := this.onlineUsers.Load(userid)
	if !ok {
		err = fmt.Errorf("当前用户不在线", userid)
	}
	onlineUser, ok = u.(*UserService)
	return
}

func (this *UserOnlineManager) GetAll() (allonlineUsers map[int]int) {
	allonlineUsers = make(map[int]int, 5)

	this.onlineUsers.Range(func(k, v interface{}) bool {
		id, _ := k.(int)

		allonlineUsers[id] = id
		return true
	})
	return
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

	this.onlineUsers.Range(func(k, v interface{}) bool {
		userid, _ := k.(int)
		if mes.UserID != userid {
			ds, ok := v.(*UserService)
			if ok {
				ds.PushServerMessage(message)
			}
		}
		return true
	})

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
		this.onlineUsers.Range(func(k, v interface{}) bool {
			userid, _ := k.(int)
			if sms.SrcUser.UserID != userid {
				ds, ok := v.(*UserService)
				if ok {
					ds.PushServerMessage(&mes)
				}
			}
			return true
		})
	} else if sms.Scope == messages.ToUsers {
		this.onlineUsers.Range(func(k, v interface{}) bool {
			userid, _ := k.(int)
			if sms.SrcUser.UserID != userid {
				s, ok := v.(*UserService)
				if ok {
					for i := 0; i < len(sms.DstUsers); i++ {
						if s.User.UserID == sms.DstUsers[i].UserID {
							s.PushServerMessage(&mes)
						}
					}
				}
			}
			return true
		})
	}

	return
}
