package services

import (
	"chattingroom/common/messages"
	"encoding/json"
	"fmt"
)

// SMSService 短消息服务结构
type SMSService struct {
	UserID int
}

func (this *SMSService) HandleMessageOfNoneRequest(message *messages.Message) {
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

func (this *SMSService) handleUserStateChangeMessage(message *messages.Message) (err error) {
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

func (this *SMSService) handleShortReceiverMessage(message *messages.Message) (err error) {
	var mes messages.ShortMessageReceiverMessage
	err = json.Unmarshal([]byte(message.Data), &mes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("您收到来自用户", mes.SrcUser.UserID, "的消息:", mes.Content)
	return
}
