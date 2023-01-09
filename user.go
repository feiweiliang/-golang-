package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	//此处的conn是客户端的socket连接句柄，不是服务端的，记得区分
	conn net.Conn
	//方便操作Server
	server *Server
}

// 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		userAddr,
		userAddr,
		make(chan string),
		conn,
		server,
	}

	//启动监听当前user channel消息的go程
	go user.ListenMessage()

	return user
}

// 用户上线业务
func (this *User) Online() {
	//用户上线，添加到OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播用户上线消息
	this.server.BroadCast(this, " Online!!!")
}

// 用户下线业务
func (this *User) Offline() {
	//用户上线，从OnlineMap删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播用户下线消息
	this.server.BroadCast(this, " Outline!!!")

}

// 给当前用户对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "is online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|新名字
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("rename error!!! Double name\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("rename successful!!!\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式：to|张三|你好张三

		//1.获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("error...\n")
			return
		}
		//2.得到对方的user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("no user: " + remoteName + "\n")
			return
		}
		//3.获取消息内容，通过对方的user对象将消息内容发过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("null...\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说: " + content + "\n")

	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前的channel，一有消息就发送回客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
