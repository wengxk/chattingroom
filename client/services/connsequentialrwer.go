package services

import (
	"chattingroom/common/messages"
	"chattingroom/common/utils"
	"context"
	"fmt"
	"io"
	"net"
)

var (
	ConnRWer *ConnSequentialRWer
)

// ConnSequentialRWer 对同一个tcp连接的顺序读取封装
type ConnSequentialRWer struct {
	Conn                             net.Conn               // TCP连接
	MessageSenderChan                chan *messages.Message // 发送消息管道
	ResponseMessageOfNoneRequestChan chan *messages.Message // 接收消息管道，没有对应请求的消息，服务器主动推送的消息
	ResponseOfRequestContainer       utils.ConcurrentMap    // 存储各个服务请求的管道
}

func NewConnSequentialRWer(conn net.Conn, senderCap int, receiverCap int) (crwer *ConnSequentialRWer) {
	return &ConnSequentialRWer{
		Conn:                             conn,
		MessageSenderChan:                make(chan *messages.Message, senderCap),
		ResponseMessageOfNoneRequestChan: make(chan *messages.Message, receiverCap),
		ResponseOfRequestContainer:       utils.NewConcurrentMap(),
	}
}

func (this *ConnSequentialRWer) Start(ctx context.Context) {
	go func() {
		mt := utils.MessageTransfer{
			Conn: this.Conn,
		}
		for {
			select {
			case <-ctx.Done():
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
	}()

	go func() {
		mt := utils.MessageTransfer{
			Conn: this.Conn,
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			mes, err := mt.ReceiveMessage()
			if err == io.EOF {
				fmt.Println("conn closed", err)
				return
			}
			if err != nil {
				fmt.Println("conn read message err", err)
				return
			}
			if mes != nil {
				reqchan, ok := this.ResponseOfRequestContainer.Get(mes.UUID)
				if !ok {
					this.ResponseMessageOfNoneRequestChan <- mes
					continue
				}
				reqchan <- mes
			}
		}
	}()
}
