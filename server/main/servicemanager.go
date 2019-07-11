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

type ServiceManager struct {
	Conn          net.Conn
	userService   *services.UserService
	daemonService *services.DaemonService
}

func (this *ServiceManager) HandleConnection() (err error) {
	defer (func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	})()

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}
	mes, err := mt.ReceiveMessage()
	if err == io.EOF {
		fmt.Println("conn closed", err)
		return err
	}
	if err != nil {
		fmt.Println("conn read message err", err)
		return err
	}
	err = this.createService(&mes)
	if err != nil {
		return err
	}
	err = this.handleService(&mes)
	if err != nil {
		fmt.Println("process message err", err)
		return err
	}
	for {
		mes, err := mt.ReceiveMessage()
		err = this.handleService(&mes)
		if err != nil {
			fmt.Println("process message err", err)
			return err
		}
	}
}

func (this *ServiceManager) createService(message *messages.Message) (err error) {
	switch message.Type {
	case messages.LoginMessageType:
		{
			this.userService = services.NewUserService(this.Conn)
		}
	case messages.RegistryMessageType:
		{
			this.userService = services.NewUserService(this.Conn)
		}
	case messages.GetOnlineUsersMessageType:
		{
			this.userService = services.NewUserService(this.Conn)
		}
	case messages.DaemonRequestMessageType:
		{
			this.daemonService = services.NewDaemonService(this.Conn)
		}
	default:
		return errors.New(fmt.Sprintln("未知消息类型，无法处理", message.Type))
	}
	return
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
	case messages.DaemonRequestMessageType:
		{
			err = this.daemonService.Accept(message)
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
