package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Client 创建用户结构体
type Client struct {
	C    chan string
	Name string
	Addr string
}

// 创建在线用户栈
var onlineMap map[string]Client

// 创建全局 channl
var message = make(chan string)

// WriteMsgToClient 监听消息并发送给客户端
func WriteMsgToClient(clnt Client, conn net.Conn) {
	// 一有消息就写给当前用户
	for msg := range clnt.C {
		conn.Write([]byte(msg + "\n"))
	}
}

func makeMsg(clnt Client, msg string) string {
	buf := "[" + clnt.Addr + " " + clnt.Name + "]:" + msg
	return buf
}

// HandleConnect 用户控制句柄
func HandleConnect(conn net.Conn) {
	isLogOut := make(chan bool)
	isAlive := make(chan bool)
	defer conn.Close()
	// 获取 ip 和port
	netAddr := conn.RemoteAddr().String()

	// create user client stuct, 默认用户名时 ip + port
	clnt := Client{make(chan string), netAddr, netAddr}

	// 将新连接用添加到map
	onlineMap[netAddr] = clnt

	// 创建给用户发送消息的go程
	go WriteMsgToClient(clnt, conn)

	// 发送用户上线消息
	message <- makeMsg(clnt, "login")

	// 创建一个go程， 广播用户消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				fmt.Println("检测到客户端推出", clnt.Name)
				isLogOut <- true
				return
			}
			if err != nil {
				fmt.Println("conn.Read err", err)
				return
			}
			// 将读到的信息广播
			msg := string(buf[:n])
			msg = msg[:len(msg)-1]
			if msg == "who" {
				rdat := ""
				for _, val := range onlineMap {
					rdat += val.Addr + ":" + val.Name + "\n"
				}
				conn.Write([]byte(rdat))
			} else if strings.HasPrefix(msg, "rename") {
				clnt.Name = msg[7:]
				onlineMap[netAddr] = clnt
				conn.Write([]byte("renameOk: " + clnt.Name))
			} else {
				message <- makeMsg(clnt, msg)
			}
			isAlive <- true
		}
	}()

	// 监控退出
	for {
		select {
		case <-isLogOut:
			// 清理用户信息
			delete(onlineMap, clnt.Addr)
			message <- makeMsg(clnt, "log out")
			return
		case <-isAlive:
			// 刷新for
		case <-time.After(time.Second * 15):
			delete(onlineMap, clnt.Addr)
			message <- makeMsg(clnt, "time out and killed")
			fmt.Println("kill ", clnt.Name)
			return
		}
	}
}

// Mannger 发送数据的管家
func Mannger() {
	// 初始化 onlineMap
	onlineMap = make(map[string]Client)

	// 监听全局中是否有数据
	for {
		msg := <-message
		for _, clnt := range onlineMap {
			clnt.C <- msg
		}
	}
}
func main() {
	// 创建监听套接字
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println("listem err")
	}
	defer listener.Close()

	// 创建管理者go 程, 管理map。 和全局chan
	go Mannger()

	// 循环监听客户端
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept Err")
		}
		// 启动 go 程
		go HandleConnect(conn)
	}
}
