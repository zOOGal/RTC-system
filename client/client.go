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
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	Client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	Client.conn = conn
	return Client

}

// deal with server's response
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)

}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1. Public conversation")
	fmt.Println("2. Private conversation")
	fmt.Println("3. Update username")
	fmt.Println("0. Exit")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("Please enter a valid number!")
		return false
	}
}

// look up current online users
func (client *Client) SelectUsers() {
	sendMsg := "Which user?\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write error: ", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println("Enter the user's name that you want to talk to: ")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("Enter msg: ")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error: ", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("Enter msg: ")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
	}
}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println("Enter msg. Press enter to exit.")
	fmt.Println(&chatMsg)

	for chatMsg != "exit" {
		// send to server

		// send msg if msg is not null
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error: ", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("please enter chat msg or enter exit")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) UpdateName() bool {
	fmt.Println("Please enter username: ")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		switch client.flag {
		case 1:
			client.PublicChat()
			break

		case 2:
			fmt.Println("Private mode...")
			break

		case 3:
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server Ip address(default 127.0.0.1)")
	flag.IntVar(&serverPort, "Port", 8888, "Set server port (default: 8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client != nil {
		fmt.Println("Fail to connect to the server...")
		return
	}

	// a seperate goroutine to deal with server's response
	go client.DealResponse()

	fmt.Println("Connected!")

	client.Run()

}
