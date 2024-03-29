package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("服务器开始监听，端口10001")
	ln, err := net.Listen("tcp", "0.0.0.0:10001")
	defer func() {
		ln.Close()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if err != nil {
		fmt.Println("net listen err", err)
		return
	}
	fmt.Println(ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("listen accept err", err)
		}
		fmt.Println(conn)
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// conn.SetDeadline(time.Now().Add(120 * time.Second))
	sm := &ServiceManager{
		Conn: conn,
	}
	defer func() {
		conn.Close()
		sm.closeDoneChan()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	err := sm.HandleConnection()
	if err != nil {
		fmt.Println(err)
		return
	}
}
