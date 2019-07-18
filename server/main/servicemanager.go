package main

import (
	"chattingroom/common/messages"
	"chattingroom/common/utils"
	"chattingroom/server/services"
	"errors"
	"fmt"
	"io"
	"net"
)

// ServiceManager TCP连接服务管理
type ServiceManager struct {
	Conn              net.Conn
	userService       *services.UserService
	MessageSenderChan chan *messages.Message
}

func (this *ServiceManager) HandleConnection() (err error) {

	this.MessageSenderChan = make(chan *messages.Message, 50)
	this.userService = services.NewUserService(this.Conn, this.MessageSenderChan)

	go this.sequentialSendMessage()

	defer (func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	})()

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}

	for {
		mes, err := mt.ReceiveMessage()
		if err == io.EOF {
			fmt.Println("conn closed", err)
			return err
		}
		if err != nil {
			fmt.Println("conn read message err", err)
			return err
		}
		if mes != nil {
			go this.handleService(mes)
		}
	}
}

// 消息处理中心，消息转发
func (this *ServiceManager) handleService(message *messages.Message) (err error) {
	switch message.Type {
	case messages.LoginMessageType:
		{
			_, err = this.userService.Login(message)
		}
	case messages.RegistryMessageType:
		{
			err = this.userService.Register(message)
		}
	case messages.GetOnlineUsersMessageType:
		{
			err = this.userService.GetOnlineUsers(message)
		}
	case messages.LogoutMessageType:
		{
			this.userService.Logout(message)
		}
	case messages.HeartBeatingMessageType:
		{
			this.userService.HeartBeat(message)
		}
	case messages.ShortMessageSenderMessageType:
		{
			err = services.Usermanager.PushServerMessage(message)
		}
	default:
		return errors.New(fmt.Sprintln("未知消息类型，无法处理", message.Type))
	}
	return
}

func (this *ServiceManager) sequentialSendMessage() {
	mt := utils.MessageTransfer{
		Conn: this.Conn,
	}
	for {
		mes := <-this.MessageSenderChan
		err := mt.SendMessage(mes)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
