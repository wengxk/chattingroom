package utils

import (
	"chattingroom/common/messages"
	"fmt"
	"net"
)

type ConnSequentialRWer struct {
	conn                net.Conn
	messageSenderChan   chan *messages.Message
	messageReceiverChan chan *messages.Message
}

func NewConnSequentialRWer(conn net.Conn, senderCap int, receiverCap int) (crwer *ConnSequentialRWer) {
	return &ConnSequentialRWer{
		conn:                conn,
		messageSenderChan:   make(chan *messages.Message, senderCap),
		messageReceiverChan: make(chan *messages.Message, receiverCap),
	}
}

func (this *ConnSequentialRWer) Start() {
	go func() {
		mt := MessageTransfer{
			Conn: this.conn,
		}
		for {
			mes := <-this.messageSenderChan
			err := mt.SendMessage(mes)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()

	go func() {
		mt := MessageTransfer{
			Conn: this.conn,
		}
		for {
			mes, err := mt.ReceiveMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			this.messageReceiverChan <- mes
		}
	}()
}

func (this *ConnSequentialRWer) Write(mes *messages.Message) {
	this.messageSenderChan <- mes
}

func (this *ConnSequentialRWer) Read() (mes *messages.Message) {
	mes = <-this.messageReceiverChan
	return mes
}
