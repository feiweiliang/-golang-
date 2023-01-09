package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Print("连接服务端失败: ", err)
		return nil
	}
	client.conn = conn

	return client
}

func (this *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, this.conn)
	/**
	// 以上写法等价于以下写法
	for {
		buf := make([]byte, 4096)
		this.conn.Read(buf)
		fmt.Println(buf)
	}
	**/
}

func (this *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.查看当前在线用户")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 4 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法的数字<<<<")
		return false
	}
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {
		}

		//根据不同模式处理不同业务
		switch this.flag {
		case 1:
			//公聊模式
			fmt.Println(">>>>>进入公聊模式<<<<<")
			this.PublicChat()
		case 2:
			//私聊模式
			fmt.Println(">>>>>进入私聊模式<<<<<")
			this.PrivateChat()
		case 3:
			//改名
			fmt.Println(">>>>>进入更新用户名模式<<<<<")
			this.UpdateName()
		case 4:
			//查看在线用户
			fmt.Println(">>>>>当前在线用户如下<<<<<")
			this.GetAllUsers()
		}
	}
}

func (this *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名:")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error !!!")
		return false
	}
	return true
}

func (this *Client) GetAllUsers() {
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error !!!")
		return
	}
	return
}

func (this *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>>请输入聊天内容,exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//消息不为空就发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write: ", err)
				return
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容,exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	this.GetAllUsers()
	fmt.Println(">>>>>请输入聊天对象[用户名],exit退出.")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {

		fmt.Println(">>>>>请输入消息内容,exit退出.")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//消息不为空就发送给服务器
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write: ", err)
					return
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>请输入消息内容,exit退出.")
			fmt.Scanln(&chatMsg)
		}

		remoteName = ""
		this.GetAllUsers()
		fmt.Println(">>>>>请输入聊天对象[用户名],exit退出.")
		fmt.Scanln(&remoteName)
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>连接服务器失败...")
	}

	//单独开一个go程，去处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>>>连接服务器成功...")

	//启动客户端服务
	client.Run()
}
