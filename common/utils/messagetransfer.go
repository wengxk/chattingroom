package utils

import (
	"chattingroom/common/messages"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

type MessageTransfer struct {
	Conn net.Conn
	// buf  [4 * 1024]byte
}

// 读取tcp包，先读取长度，再按照长度接收消息
func (this *MessageTransfer) ReceiveMessage() (mes messages.Message, err error) {
	buf := make([]byte, 4*1024)
	//只要客户端连接没有关闭，这里就会一直阻塞等待读，直到客户端关闭，返回err io.EOF
	// 这里若使用this.buf[:4]，后面读取的长度会为0，this.buf是值类型，传参是值传参，然后再在这个副本上分出一个切片
	_, err = this.Conn.Read(buf[:4])
	if err != nil {
		fmt.Println(err)
		return
	}
	pkglength := binary.BigEndian.Uint32(buf[:4])
	// fmt.Println("read length", pkglength)

	n, err := this.Conn.Read(buf[:pkglength])
	if n != int(pkglength) {
		fmt.Println("length of received data err,excepted length", pkglength, "actual received length", n)
	}
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(buf[:pkglength], &mes)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// 发送tcp包，先发送包的长度，再发送消息
func (this *MessageTransfer) SendMessage(data []byte) (err error) {
	// 获取发送消息的长度
	var pkglength uint32
	pkglength = uint32(len(data))
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[0:4], pkglength)
	// 发送消息的长度
	n, err := this.Conn.Write(buf[0:4])
	if n != 4 || err != nil {
		fmt.Println("conn write length of message err", err)
		return err
	}
	// 发送消息
	n, err = this.Conn.Write(data)
	if n != int(pkglength) || err != nil {
		fmt.Println("conn write message err", err)
		return err
	}
	return
}
