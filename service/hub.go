
package main

import (
	"fmt"
	"strings"
	"time"
)

type tmessage struct {
	content    []byte
	fromuser   []byte
	touser     []byte
	mtype      int
	createtime string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	//注册连接
	connections map[*connection]bool

	// Inbound messages from the connections.
	//连接中的绑定消息
	broadcast chan *tmessage

	// Register requests from the connections.
	//添加新连接
	register chan *connection

	// Unregister requests from connections.
	//删除连接
	unregister chan *connection
}

var h = hub{
	//广播slice
	broadcast: make(chan *tmessage),
	//注册者slice
	register: make(chan *connection),
	//未注册者sclie
	unregister: make(chan *connection),
	//连接map
	connections: make(map[*connection]bool),
	//在线用户数

}

func (h *hub) run() {

	for {

		select {

		//注册者有数据，则插入连接map
		case c := <-h.register:
			h.connections[c] = true

			remoteIp := strings.Split(c.ws.RemoteAddr().String(), ":")[0]
			c.username = c.GetRandomString(5)
			c.createip = []byte(remoteIp)
			c.auth = true

			username := "<b>" + string(c.username) + "</b>"
			c.send <- []byte("<span style='color:red'>welcome [ " + username + " ] chat secret ^_^</span>")
			t := time.Now().Format(f_times)

			for tmp_c := range h.connections {
				tmp_c.send <- []byte("system " + t + " : [ " + username + " join ] " + fmt.Sprintf("%d", len(h.connections)))

			}

		//非注册者有数据，则删除连接map
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
			t := time.Now().Format(f_times)
			username := "<b>" + string(c.username) + "</b>"
			for tmp_c := range h.connections {
				tmp_c.send <- []byte("system " + t + " : [ " + username + " gone ] " + fmt.Sprintf("%d", len(h.connections)))
			}
		//广播有数据
		case m := <-h.broadcast:
			//递归所有广播连接
			for c := range h.connections {
				var send_flag = false

				//根据广播消息标识记录
				/*
					text2 := string(m.content)
					reg2 := regexp.MustCompile(`^@.*? `)
					s2 := reg2.FindAllString(text2, -1)
				*/
				var send_msg []byte
				if m.mtype == 1 { //系统消息
					send_msg = []byte("system " + m.createtime + " : [ " + string(m.content) + " ]")
				} else if m.mtype == 2 { //用户消息
					temp_msg := string(m.fromuser) + " " + m.createtime + " : [ " + string(m.content) + " ] "
					if string(m.fromuser) == string(c.username) {
						temp_msg = "<b>" + temp_msg + "</b>"
					}
					send_msg = []byte(temp_msg)
				} else {
					send_msg = []byte(string(m.content))
				}

				if string(m.touser) != "all" {
					temp_msg := string(m.fromuser) + " " + m.createtime + " whisper : [ " + string(m.content) + " ] "

					if string(c.username) == string(m.touser) || string(c.username) == string(m.fromuser) {
						send_flag = true
						temp_msg = "<i>" + temp_msg + "</i>"
					}

					send_msg = []byte(temp_msg)
					if send_flag {
						select {
						//发送数据给连接
						case c.send <- send_msg:
							fmt.Println("[" + fmt.Sprintf("%d", len(h.connections)) + "]" + string(c.createip) + " " + string(send_msg))
						//关闭连接
						default:
							close(c.send)
							delete(h.connections, c)
						}
					}
				} else {
					select {
					//发送数据给连接
					case c.send <- send_msg:
						//int 转换成string 不能使用string(12)否则打印出来是空
						//str1 := strconv.Itoa(i)
						//str2 := fmt.Sprintf("%d", i)
						fmt.Println("[" + fmt.Sprintf("%d", len(h.connections)) + "]" + string(c.createip) + " " + string(send_msg))
					//关闭连接
					default:
						close(c.send)
						delete(h.connections, c)
					}
				}

			}
		}
	}
}
