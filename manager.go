package main //@包主

import ( //@进口
	"context" //@语境
	"encoding/json" //@编码json
	"errors" //@错误
	"log" //@日志
	"net/http" //@净http
	"sync" //@同步
	"time" //@时间

	"github.com/gorilla/websocket" //@github com 大猩猩 websocket
)

var ( //@变量
	/**
	websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection //@websocket 升级器用于将传入的 http 请求升级为持久的 websocket 连接
	*/
	websocketUpgrader = websocket.Upgrader{ //@websocket 升级器 websocket 升级器
		// Apply the Origin Checker //@应用原点检查器
		CheckOrigin:     checkOrigin, //@检查原点 检查原点
		ReadBufferSize:  1024, //@读取缓冲区大小
		WriteBufferSize: 1024, //@写缓冲区大小
	}
)

var ( //@变量
	ErrEventNotSupported = errors.New("this event type is not supported") //@错误事件不支持错误新不支持此事件类型
)

// checkOrigin will check origin and return true if its allowed //@检查原点将检查原点并在允许的情况下返回 true
func checkOrigin(r *http.Request) bool { //@func check origin r http 请求 bool

	// Grab the request origin //@抓取请求来源
	origin := r.Header.Get("Origin") //@origin r header 获取原点

	switch origin { //@切换原点
	// Update this to HTTPS //@将此更新为 https
	case "https://localhost:8080": //@案例 https localhost
		return true //@返回真
	default: //@默认
		return false //@返回假
	}
}

// Manager is used to hold references to all Clients Registered, and Broadcasting etc //@经理用于保存对所有注册和广播等客户的引用
type Manager struct { //@类型管理器结构
	clients ClientList //@客户客户名单

	// Using a syncMutex here to be able to lcok state before editing clients //@在此处使用同步互斥锁，以便能够在编辑客户端之前锁定状态
	// Could also use Channels to block //@也可以使用通道来阻止
	sync.RWMutex //@同步读写互斥
	// handlers are functions that are used to handle Events //@处理程序是用于处理事件的函数
	handlers map[string]EventHandler //@处理程序映射字符串事件处理程序
	// otps is a map of allowed OTP to accept connections from //@otps 是允许 otp 接受来自的连接的映射
	otps RetentionMap //@otps保留地图
}

// NewManager is used to initalize all the values inside the manager //@new manager 用于初始化 manager 中的所有值
func NewManager(ctx context.Context) *Manager { //@func new manager ctx context 上下文管理器
	m := &Manager{ //@经理
		clients:  make(ClientList), //@客户制作客户名单
		handlers: make(map[string]EventHandler), //@处理程序使映射字符串事件处理程序
		// Create a new retentionMap that removes Otps older than 5 seconds //@创建一个新的保留映射，删除早于秒的 otps
		otps: NewRetentionMap(ctx, 5*time.Second), //@otps new retention map ctx 时间秒
	}
	m.setupEventHandlers() //@m 设置事件处理程序
	return m //@返回米
}

// setupEventHandlers configures and adds all handlers //@设置事件处理程序配置并添加所有处理程序
func (m *Manager) setupEventHandlers() { //@func m 管理器设置事件处理程序
	m.handlers[EventSendMessage] = SendMessageHandler //@m handlers event send message 发送消息处理器
	m.handlers[EventChangeRoom] = ChatRoomHandler //@m handlers event change room 聊天室处理程序
}

// routeEvent is used to make sure the correct event goes into the correct handler //@路由事件用于确保正确的事件进入正确的处理程序
func (m *Manager) routeEvent(event Event, c *Client) error { //@func m manager route event event event c 客户端错误
	// Check if Handler is present in Map //@检查地图中是否存在处理程序
	if handler, ok := m.handlers[event.Type]; ok { //@如果处理程序正常 m 处理程序事件类型正常
		// Execute the handler and return any err //@执行处理程序并返回任何错误
		if err := handler(event, c); err != nil { //@如果错误处理程序事件 c err nil
			return err //@返回错误
		}
		return nil //@返回零
	} else { //@别的
		return ErrEventNotSupported //@不支持返回错误事件
	}
}

// loginHandler is used to verify an user authentication and return a one time password //@登录处理程序用于验证用户身份验证并返回一次性密码
func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {

	type userLoginRequest struct { //@输入用户登录请求结构
		Username string `json:"username"` //@用户名字符串 json 用户名
		Password string `json:"password"` //@密码字符串 json 密码
	}

	var req userLoginRequest //@var req 用户登录请求
	err := json.NewDecoder(r.Body).Decode(&req) //@错误 json 新解码器 r 主体解码请求
	if err != nil { //@如果错误为零
		http.Error(w, err.Error(), http.StatusBadRequest) //@http error w err 错误 http status bad request
		return //@返回
	}

	// Authenticate user / Verify Access token, what ever auth method you use //@验证用户验证访问令牌您使用的任何身份验证方法
	if req.Username == "percy" && req.Password == "123" { //@如果需要用户名 percy 需要密码
		// format to return otp in to the frontend //@将 otp 返回到前端的格式
		type response struct { //@类型响应结构
			OTP string `json:"otp"` //@otp 字符串 json otp
		}

		// add a new OTP //@添加一个新的 otp
		otp := m.otps.NewOTP() //@otp m otps 新的 ot p

		resp := response{ //@响应响应
			OTP: otp.Key, //@otp 密钥
		}

		data, err := json.Marshal(resp) //@数据错误 json marshal resp
		if err != nil { //@如果错误为零
			log.Println(err) //@日志打印错误
			return //@返回
		}
		// Return a response to the Authenticated user with the OTP //@使用 otp 向经过身份验证的用户返回响应
		w.WriteHeader(http.StatusOK) //@w 写入标头 http 状态正常
		w.Write(data) //@w写数据
		return //@返回
	}

	// Failure to auth //@授权失败
	w.WriteHeader(http.StatusUnauthorized) //@w 写入标头 http 状态未经授权
}

// serveWS is a HTTP Handler that the has the Manager that allows connections //@serve ws 是一个 http 处理程序，它具有允许连接的管理器
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	// Grab the OTP in the Get param //@获取 get 参数中的 otp
	otp := r.URL.Query().Get("otp") //@otp r ur l 查询获取 otp
	if otp == "" { //@如果 otp
		// Tell the user its not authorized //@告诉用户它没有被授权
		w.WriteHeader(http.StatusUnauthorized) //@w 写入标头 http 状态未经授权
		return //@返回
	}

	// Verify OTP is existing //@验证 otp 是否存在
	if !m.otps.VerifyOTP(otp) { //@如果 m otps 验证 ot p otp
		w.WriteHeader(http.StatusUnauthorized) //@w 写入标头 http 状态未经授权
		return //@返回
	}

	log.Println("New connection") //@记录 println 新连接
	// Begin by upgrading the HTTP request //@首先升级 http 请求
	conn, err := websocketUpgrader.Upgrade(w, r, nil) //@conn err websocket 升级器升级 w r nil
	if err != nil { //@如果错误为零
		log.Println(err) //@日志打印错误
		return //@返回
	}

	// Create New Client //@创建新客户
	client := NewClient(conn, m) //@客户 新客户 conn m
	// Add the newly created client to the manager //@将新创建的客户端添加到管理器
	m.addClient(client) //@m 添加客户端客户端

	go client.readMessages() //@去客户端读取消息
	go client.writeMessages() //@去客户端写消息
}

// addClient will add clients to our clientList //@添加客户会将客户添加到我们的客户列表中
func (m *Manager) addClient(client *Client) { //@func m manager 添加客户客户客户
	// Lock so we can manipulate //@锁定以便我们可以操作
	m.Lock() //@米锁
	defer m.Unlock() //@延迟解锁

	// Add Client //@添加客户
	m.clients[client] = true //@m 客户 客户 真
}

// removeClient will remove the client and clean up //@删除客户端将删除客户端并清理
func (m *Manager) removeClient(client *Client) { //@func m manager 删除客户客户客户
	m.Lock() //@米锁
	defer m.Unlock() //@延迟解锁

	// Check if Client exists, then delete it //@检查客户端是否存在然后将其删除
	if _, ok := m.clients[client]; ok { //@如果没问题 m 客户 客户没问题
		// close connection //@紧密联系
		client.connection.Close() //@客户端连接关闭
		// remove //@消除
		delete(m.clients, client) //@删除 m 个客户 client
	}
}
