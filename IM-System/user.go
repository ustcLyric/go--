package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户名
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听
	go user.ListenMessage()
	return user
}

// 监听user channel
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 上线
func (this *User) Online() {
	// 用户上线加入
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// 广播上线消息
	this.server.BroadCast(this, "上线提醒")
}

// 下线
func (this *User) Offline() {
	// 用户下线
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	// 广播上线消息
	this.server.BroadCast(this, "下线提醒")
}

// 处理业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前存在用户返回
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 6 && msg[:7] == "rename|" {
		this.server.mapLock.Lock()
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("当前用户名以被占用\n")
		} else {
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.Name = newName
			this.sendMsg("您的名字已更新\n")
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 4 && msg[:3] == "to|" {
		toUserName := strings.Split(msg, "|")[1]
		if toUserName == "" {
			this.sendMsg("当前格式不正确 参考格式 to|接受人｜消息\n")
			return
		}

		toUser, ok := this.server.OnlineMap[toUserName]
		if !ok {
			this.sendMsg("接受消息者不存在\n")
			return
		}
		context := strings.Split(msg, "|")[2]
		if context == "" {
			this.sendMsg("当前格式不正确 参考格式 to|接受人｜消息\n")
			return
		}
		toUser.sendMsg("[" + this.Name + "] 对你说:" + context + "\n")

	} else {
		this.server.BroadCast(this, msg)
	}

}

func (this *User) sendMsg(onlineMsg string) {
	this.conn.Write([]byte(onlineMsg))
}
