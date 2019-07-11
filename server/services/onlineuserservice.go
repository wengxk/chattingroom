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
	// ChanServerMessage 用于传递服务器需要主动推送的信息
	ChanServerMessage = make(chan messages.Message)
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

func (this *UserOnlineManager) Add(onlineUser *DaemonService) {
	// this.onlineUsers[onlineUser.UserID] = onlineUser
	this.onlineUsers.Store(onlineUser.UserID, onlineUser)
}

func (this *UserOnlineManager) Remove(userid int) {
	// delete(this.onlineUsers, userid)
	this.onlineUsers.Delete(userid)
	return
}

func (this *UserOnlineManager) Get(userid int) (onlineUser *DaemonService, err error) {
	// onlineUser, ok := this.onlineUsers[userid]
	// if !ok {
	// 	err = fmt.Errorf("当前用户不在线", userid)
	// }

	u, ok := this.onlineUsers.Load(userid)
	if !ok {
		err = fmt.Errorf("当前用户不在线", userid)
	}
	onlineUser, ok = u.(*DaemonService)
	return
}

func (this *UserOnlineManager) GetAll() (allonlineUsers map[int]int) {
	// allonlineUsers = this.onlineUsers
	// return allonlineUsers
	allonlineUsers = make(map[int]int, 5)

	this.onlineUsers.Range(func(k, v interface{}) bool {
		id, _ := k.(int)
		// u, _ := v.(*DaemonService)
		// allonlineUsers[id] = u
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
	// for _, v := range this.onlineUsers {
	// 	if mes.UserID == v.UserID {
	// 		continue
	// 	}
	// 	_ = v.PushServerMessage(message)
	// }
	this.onlineUsers.Range(func(k, v interface{}) bool {
		userid, _ := k.(int)
		if mes.UserID != userid {
			ds, ok := v.(*DaemonService)
			if ok {
				_ = ds.PushServerMessage(message)
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
		// for k, v := range this.onlineUsers {
		// 	if k == sms.SrcUser.UserID {
		// 		continue
		// 	}
		// 	v.PushServerMessage(&mes)
		// }
		this.onlineUsers.Range(func(k, v interface{}) bool {
			userid, _ := k.(int)
			if sms.SrcUser.UserID != userid {
				ds, ok := v.(*DaemonService)
				if ok {
					_ = ds.PushServerMessage(&mes)
				}
			}
			return true
		})
	} else if sms.Scope == messages.ToUsers {
		// for k, v := range this.onlineUsers {
		// 	if k == sms.SrcUser.UserID {
		// 		continue
		// 	}
		// 	for i := 0; i < len(sms.DstUsers); i++ {
		// 		if v.UserID == sms.DstUsers[i].UserID {
		// 			v.PushServerMessage(&mes)
		// 		}
		// 	}
		// }
		this.onlineUsers.Range(func(k, v interface{}) bool {
			userid, _ := k.(int)
			if sms.SrcUser.UserID != userid {
				s, ok := v.(*DaemonService)
				if ok {
					for i := 0; i < len(sms.DstUsers); i++ {
						if s.UserID == sms.DstUsers[i].UserID {
							_ = s.PushServerMessage(&mes)
						}
					}
				}
			}
			return true
		})

	}

	return
}
