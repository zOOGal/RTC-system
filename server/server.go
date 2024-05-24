package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// online user map
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// broadcasting channel
	Message chan string
}

// server interface
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// send msg to all the users online
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg

}

func (this *Server) Handler(conn net.Conn) {
	// current connection
	// fmt.Println("Connection established")
	user := NewUser(conn, this)

	// user is online, add the user to usermap
	user.Online()

	// listen channel: if user is active
	isLive := make(chan bool)

	// receive msg from user
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn read error: ", err)
				return
			}

			// get user msg
			msg := string(buf[:n-1])

			user.ProcessMsg(msg)

			isLive <- true
		}
	}()

	// handler congestion
	for {
		select {
		case <-isLive:
			// do nothing

			// overtime: kick user out of the system
		case <-time.After(time.Second * 10):
			user.SendMsg("You're offline")
			close(user.C)
			conn.Close()

			// exit handler
			return

		}
	}

}

// start server interface
func (this *Server) Start() {
	// listen to  socket
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen error: ", err)
		return
	}
	// close listening to the socket
	defer listener.Close()

	// start listening
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener acception error: ", err)
			continue
		}

		// handler
		go this.Handler(conn)

	}

}
