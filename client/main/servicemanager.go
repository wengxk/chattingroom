package main

import (
	"chattingroom/client/services"
	"net"
)

// ServiceManager 服务管理，初始化服务必要元素，为入口提供服务
type ServiceManager struct {
	Conn        net.Conn
	UserService *services.UserService
	SMSService  *services.SMSService
}

func NewServiceManage(conn net.Conn) (sm *ServiceManager) {
	return &ServiceManager{
		Conn:        conn,
		UserService: &services.UserService{},
		SMSService:  &services.SMSService{},
	}
}

func (this *ServiceManager) Serve() {
	services.ConnRWer = services.NewConnSequentialRWer(this.Conn, 20, 20)
	services.ConnRWer.Start()

	go func() {
		for {
			mes := <-services.ConnRWer.ResponseMessageOfNoneRequestChan
			go this.SMSService.HandleMessageOfNoneRequest(mes)
		}
	}()
}
