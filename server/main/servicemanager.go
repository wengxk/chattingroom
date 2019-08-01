package main

import (
	"chattingroom/common/messages"
	"chattingroom/common/utils"
	"chattingroom/server/services"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/google/uuid"
)

// ServiceManager TCP连接服务管理
type ServiceManager struct {
	Conn              net.Conn
	userService       *services.UserService
	MessageSenderChan chan *messages.Message
	doneChan          chan struct{}
	lastReadTime      int64 // last readed mes time of unix
	idleTime          int64 // idle time
}

func (this *ServiceManager) closeDoneChan() {
	select {
	case <-this.doneChan:
	default:
		close(this.doneChan)
		this.userService.Offline()
		fmt.Println("close")
	}
}

func (this *ServiceManager) HandleConnection() (err error) {
	this.doneChan = make(chan struct{})
	this.lastReadTime = time.Now().Unix()
	this.idleTime = 120
	this.MessageSenderChan = make(chan *messages.Message, 50)
	this.userService = services.NewUserService(this.Conn, this.MessageSenderChan)
	go this.sequentialSendMessage()

	defer (func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	})()

	go this.checkAlive()
	go this.beating()

	mt := &utils.MessageTransfer{
		Conn: this.Conn,
	}
	for {
		select {
		case <-this.doneChan:
			return
		default:
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
		if mes != nil {
			go this.handleService(mes)
			this.lastReadTime = time.Now().Unix()
		}
	}

	return
}

func (this *ServiceManager) checkAlive() {
	ticker := time.NewTicker(120 * time.Second)
	for {
		select {
		case <-this.doneChan:
			return
		default:
		}
		_, ok := <-ticker.C
		if ok {
			if time.Now().Unix()-this.lastReadTime > this.idleTime {
				this.closeDoneChan()
				return
			}
		} else {
			this.closeDoneChan()
			return
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
		select {
		case <-this.doneChan:
			return
		default:
		}
		mes := <-this.MessageSenderChan
		err := mt.SendMessage(mes)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (this *ServiceManager) beating() (err error) {
	mes := &messages.Message{
		Type: messages.HeartBeatingMessageType,
		UUID: uuid.New().String(),
	}
	heart := messages.HeartBeatingMessage{}
	data, err := json.Marshal(heart)
	if err != nil {
		fmt.Println(err)
		return
	}
	mes.Data = string(data)

	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-this.doneChan:
			return
		default:
		}
		_, ok := <-ticker.C
		if ok {
			this.MessageSenderChan <- mes
		}
	}

	return
}
