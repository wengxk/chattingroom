package services

import (
	"chattingroom/common/messages"
	"chattingroom/common/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// DaemonService 专门用于接收服务器的主动推送信息,例如用户状态改变信息的接收和聊天短消息的接收处理
type DaemonService struct {
	Conn   net.Conn
	UserID int
}

func (this *DaemonService) request() (err error) {
	mes := messages.Message{
		Type: messages.DaemonRequestMessageType,
	}
	requestmes := messages.DaemonRequestMessage{
		UserID: this.UserID,
	}

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

	resultmes := messages.DaemonResponseMessage{}
	err = json.Unmarshal([]byte(mes.Data), &resultmes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resultmes.Code == 200 {
		fmt.Println("daemon service created", this.UserID)
	} else {
		fmt.Println("daemon service create failed", this.UserID)
		return errors.New(resultmes.Error)
	}
	return
}

// 专门用来接收服务器主动推送的消息
func (this *DaemonService) ProcessServerMessage() {

	err := this.request()
	if err != nil {
		fmt.Println(err)
		return
	}

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}

	for {
		fmt.Println("客户端正在读取服务器消息")
		mes, err := mt.ReceiveMessage()
		if err != nil {
			fmt.Println("客户端读取服务器消息出错", err)
			return
		}
		this.handleServerPushMessage(&mes)
	}
}

func (this *DaemonService) handleServerPushMessage(message *messages.Message) (err error) {
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

func (this *DaemonService) handleUserStateChangeMessage(message *messages.Message) (err error) {
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

func (this *DaemonService) handleShortReceiverMessage(message *messages.Message) (err error) {
	var mes messages.ShortMessageReceiverMessage
	err = json.Unmarshal([]byte(message.Data), &mes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("您收到来自用户", mes.SrcUser.UserID, "的消息:", mes.Content)
	return
}
