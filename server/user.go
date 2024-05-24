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

// User API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// start listeing
	go user.ListenMessage()

	return user
}

// user online
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// broadcast user online info
	this.server.Broadcast(this, "is online")
}

// user offline
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// broadcast user online info
	this.server.Broadcast(this, "is offline")
}

// send msg to current client
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// process user msg
func (this *User) ProcessMsg(msg string) {
	if msg == "who" {
		// serach current online users
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "is online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		_, existed := this.server.OnlineMap[newName]
		if existed {
			this.SendMsg("This name has been used\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("Username had been updated: " + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {

		// get receiver's name
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("Incorrect message type, please use \"to|xxx|abcdefg\"format\n")
			return
		}
		remoteUser, existed := this.server.OnlineMap[remoteName]
		if !existed {
			this.SendMsg("Can't find this user!")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("Msg is empty! PLease send again!\n")
			return
		}
		remoteUser.SendMsg(this.Name + "sent a message:" + content)

	} else {
		this.server.Broadcast(this, msg)
	}
}

// listen to user channel; send to client once messaage generated
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))

	}
}
