package main //@包主

import ( //@进口
	"encoding/json" //@编码json
	"log" //@日志
	"time" //@时间

	"github.com/gorilla/websocket" //@github com 大猩猩 websocket
)

// ClientList is a map used to help manage a map of clients //@客户列表是用于帮助管理客户地图的地图
type ClientList map[*Client]bool //@类型客户端列表映射客户端布尔

// Client is a websocket client, basically a frontend visitor //@客户端是一个 websocket 客户端，基本上是一个前端访问者
type Client struct { //@类型客户端结构
	// the websocket connection //@网络套接字连接
	connection *websocket.Conn //@连接 websocket conn

	// manager is the manager used to manage the client //@manager 是用来管理client的manager
	manager *Manager //@经理经理
	// egress is used to avoid concurrent writes on the WebSocket //@出口用于避免在网络套接字上并发写入
	egress chan Event //@出口陈事件
	// chatroom is used to know what room user is in //@聊天室用于了解用户所在的房间
	chatroom string //@聊天室字符串
}

var ( //@变量
	// pongWait is how long we will await a pong response from client //@pong wait 是我们等待客户端的 pong 响应的时间
	pongWait = 10 * time.Second //@乒乓等待时间秒
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time //@ping 间隔必须小于 pong wait 我们不能乘以得到时间
	// Because that can make decimals, so instead *9 / 10 to get 90% //@因为那可以使小数所以而不是得到
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response //@它必须小于 ping 频率的原因是因为它会在收到响应之前发送一个新的 ping
	pingInterval = (pongWait * 9) / 10 //@ping 间隔 pong 等待
)

// NewClient is used to initialize a new Client with all required values initialized //@new client 用于初始化一个新的客户端，并初始化所有需要的值
func NewClient(conn *websocket.Conn, manager *Manager) *Client { //@func new client conn websocket conn manager manager 客户端
	return &Client{ //@回头客
		connection: conn, //@连接conn
		manager:    manager, //@经理经理
		egress:     make(chan Event), //@出口 make chan 事件
	}
}

// readMessages will start the client to read messages and handle them //@读取消息将启动客户端读取消息并处理它们
// appropriatly. //@恰当地
// This is suppose to be ran as a goroutine //@这应该作为 goroutine 运行
func (c *Client) readMessages() { //@func c 客户端读取消息
	defer func() { //@延迟函数
		// Graceful Close the Connection once this //@优雅地关闭连接一次
		// function is done //@功能完成
		c.manager.removeClient(c) //@c经理删除客户c
	}()
	// Set Max Size of Messages in Bytes //@以字节为单位设置消息的最大大小
	c.connection.SetReadLimit(512) //@c连接设置读取限制
	// Configure Wait time for Pong response, use Current time + pongWait //@配置乒乓响应的等待时间使用当前时间乒乓等待
	// This has to be done here to set the first initial timer. //@这必须在此处完成以设置第一个初始计时器
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil { //@如果 err c connection set read deadline time now add pong wait err nil
		log.Println(err) //@日志打印错误
		return //@返回
	}
	// Configure how to handle Pong responses //@配置如何处理 pong 响应
	c.connection.SetPongHandler(c.pongHandler) //@c 连接设置 pong 处理程序 c pong 处理程序

	// Loop Forever //@永远循环
	for { //@为了
		// ReadMessage is used to read the next message in queue //@读取消息用于读取队列中的下一条消息
		// in the connection //@在连接
		_, payload, err := c.connection.ReadMessage() //@payload err c 连接读取消息

		if err != nil { //@如果错误为零
			// If Connection is closed, we will Recieve an error here //@如果连接关闭，我们将在此处收到错误消息
			// We only want to log Strange errors, but simple Disconnection //@我们只想记录奇怪的错误但简单的断开连接
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { //@if websocket is unexpected close error err websocket close going away websocket close 异常关闭
				log.Printf("error reading message: %v", err) //@记录 printf 错误读取消息 v err
			}
			break // Break the loop to close conn & Cleanup //@break 打破循环以关闭 conn 清理
		}
		// Marshal incoming data into a Event struct //@将传入数据编组到事件结构中
		var request Event //@变量请求事件
		if err := json.Unmarshal(payload, &request); err != nil { //@如果错误 json 解组有效负载请求错误 nil
			log.Printf("error marshalling message: %v", err) //@记录 printf 错误编组消息 v err
			break // Breaking the connection here might be harsh xD //@break 在这里断开连接可能会很刺耳 x d
		}
		// Route the Event //@路由事件
		if err := c.manager.routeEvent(request, c); err != nil { //@if err c manager 路由事件请求 c err nil
			log.Println("Error handeling Message: ", err) //@记录 println 错误处理消息 err
		}
	}
}

// pongHandler is used to handle PongMessages for the Client //@pong 处理程序用于为客户端处理 pong 消息
func (c *Client) pongHandler(pongMsg string) error { //@func c 客户端 pong 处理程序 pong 消息字符串错误
	// Current time + Pong Wait time //@当前时间乒乓等待时间
	log.Println("pong") //@日志打印乒乓
	return c.connection.SetReadDeadline(time.Now().Add(pongWait)) //@返回 c 连接设置读取截止时间现在添加乒乓等待
}

// writeMessages is a process that listens for new messages to output to the Client //@write messages是一个监听新消息输出到客户端的进程
func (c *Client) writeMessages() { //@func c 客户端写消息
	// Create a ticker that triggers a ping at given interval //@创建一个在给定时间间隔触发 ping 的自动收报机
	ticker := time.NewTicker(pingInterval) //@ticker time 新的 ticker ping 间隔
	defer func() { //@延迟函数
		ticker.Stop() //@股票止损
		// Graceful close if this triggers a closing //@如果这触发关闭，则优雅关闭
		c.manager.removeClient(c) //@c经理删除客户c
	}()

	for { //@为了
		select { //@选择
		case message, ok := <-c.egress: //@案例消息 ok c egress
			// Ok will be false Incase the egress channel is closed //@如果出口通道关闭，ok 将是 false
			if !ok { //@如果可以的话
				// Manager has closed this connection channel, so communicate that to frontend //@经理已关闭此连接通道，因此请将其传达给前端
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil { //@如果错误 c 连接写入消息 websocket 关闭消息 nil err nil
					// Log that the connection is closed and the reason //@记录连接关闭和原因
					log.Println("connection closed: ", err) //@日志 println 连接关闭错误
				}
				// Return to close the goroutine //@返回关闭 goroutine
				return //@返回
			}

			data, err := json.Marshal(message) //@数据错误 json 编组消息
			if err != nil { //@如果错误为零
				log.Println(err) //@日志打印错误
				return // closes the connection, should we really //@返回关闭连接我们真的应该
			}
			// Write a Regular text message to the connection //@将常规文本消息写入连接
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil { //@if err c connection write message websocket text message 数据 err nil
				log.Println(err) //@日志打印错误
			}
			log.Println("sent message") //@记录 println 发送的信息
		case <-ticker.C: //@案例代码 c
			log.Println("ping") //@日志打印ping
			// Send the Ping //@发送 ping
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err) //@日志 println writemsg 错误
				return // return to break this goroutine triggeing cleanup //@return return 中断这个 goroutine 触发清理
			}
		}

	}
}
