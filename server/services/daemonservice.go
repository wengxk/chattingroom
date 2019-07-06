package services

import (
	"chattingroom/common/messages"
	"chattingroom/common/utils"
	"encoding/json"
	"fmt"
	"net"
)

// DaemonService 专用于服务器主动推送消息,例如用户状态改变信息和聊天信息的主动推送
type DaemonService struct {
	Conn   net.Conn
	UserID int
}

func NewDaemonService(conn net.Conn) (deamonservice *DaemonService) {
	return &DaemonService{
		Conn: conn,
	}
}

func (this *DaemonService) Accept(message *messages.Message) (err error) {
	mes := messages.Message{
		Type: messages.DaemonResponseMessageType,
	}
	requesetmes := messages.DaemonRequestMessage{}
	resultmes := messages.DaemonResponseMessage{}

	err = json.Unmarshal([]byte(message.Data), &requesetmes)
	if err != nil {
		resultmes.Code = 500
		resultmes.Error = err.Error()
	} else {
		this.UserID = requesetmes.UserID
		Usermanager.Add(this)
		resultmes.Code = 200
	}

	data, err := json.Marshal(resultmes)
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
	fmt.Println("daemon service created", requesetmes.UserID)

	// 推送用户上线消息
	mes = messages.Message{
		Type: messages.UserStateChangeMessageType,
	}
	usmes := messages.UserStateChangeMessage{
		UserID:    requesetmes.UserID,
		UserState: messages.UserOnline,
	}
	data, err = json.Marshal(usmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)

	Usermanager.PushServerMessage(&mes)
	return
}

func (this *DaemonService) PushServerMessage(message *messages.Message) (err error) {
	data, err := json.Marshal(message)
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
