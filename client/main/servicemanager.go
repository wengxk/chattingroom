package main

import (
	"chattingroom/client/services"
	"chattingroom/common/messages"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// ServiceManager 服务管理，初始化服务必要元素，为入口提供服务
type ServiceManager struct {
	Conn         net.Conn
	UserService  *services.UserService
	SMSService   *services.SMSService
	doneChan     chan struct{}
	lastReadTime int64
	idleTime     int64
	cancelCtx    context.CancelFunc
}

func NewServiceManage(conn net.Conn) (sm *ServiceManager) {
	return &ServiceManager{
		Conn:         conn,
		UserService:  &services.UserService{},
		SMSService:   &services.SMSService{},
		doneChan:     make(chan struct{}),
		lastReadTime: time.Now().Unix(),
		idleTime:     120,
	}
}

func (this *ServiceManager) closeDoneChan() {
	select {
	case <-this.doneChan:
	default:
		close(this.doneChan)
		this.cancelCtx()
		this.Conn.Close()
		fmt.Println("close")
	}
}

func (this *ServiceManager) Serve() {
	services.ConnRWer = services.NewConnSequentialRWer(this.Conn, 20, 20)
	ctx := context.Background()
	ctx, this.cancelCtx = context.WithCancel(ctx)
	services.ConnRWer.Start(ctx)
	go func() {
		for {
			select {
			case <-this.doneChan:
				return
			default:
			}
			mes := <-services.ConnRWer.ResponseMessageOfNoneRequestChan
			this.lastReadTime = time.Now().Unix()
			go this.SMSService.HandleMessageOfNoneRequest(mes)
		}
	}()

	go this.checkAlive()
	go this.beating()
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
				fmt.Println("当前会话已超时，请重新登录")
				return
			}
		} else {
			this.closeDoneChan()
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

	ticker := time.NewTicker(120 * time.Second)
	for {
		select {
		case <-this.doneChan:
			return
		default:
		}
		_, ok := <-ticker.C
		if ok {
			services.ConnRWer.MessageSenderChan <- mes
		}
	}

	return
}
