package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// 定义一个Server的结构体
type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播
	Message chan string
}

// 监听Message广播消息channel的go程，一旦有消息就发送给全部在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//将消息发送给全部在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//当前连接的业务

	user := NewUser(conn, this)

	user.Online()

	//监听用户是否活跃
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
				return
			}
			//提取用户消息，去除换行
			msg := string(buf[:n-1])
			//广播消息
			user.DoMessage(msg)

			//用户发送任意消息，代表他是活跃的
			isLive <- true
		}

	}()

	//阻塞当前handler
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，应该重置定时器
			//不用做任何事情，为了激活select,更新定时器
			//当进入select后，<-isLive有数据没有阻塞，继续判断下一个case，从而运行了time.After导致定时器刷新
		case <-time.After(time.Second * 1000):
			//已经超时，强制关闭当前User
			user.SendMsg("timeout!!!")
			//销毁用的资源
			close(user.C)
			//关闭连接
			conn.Close()
			//退出当前handler
			return
		}
	}

}

// 创建一个Server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 启动服务器
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Print("net.Listen error:", err)
		return
	}

	//关闭socket
	defer listener.Close()

	//启动监听Message的go程
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Print("listener.Accept error:", err)
			return
		}
		//do handler
		go this.Handler(conn)
	}

}
