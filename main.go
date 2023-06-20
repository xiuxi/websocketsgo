package main //@包主

import ( //@进口
	"context" //@语境
	"fmt" //@调速器
	"log" //@日志
	"net/http" //@净http
)

func main() { //@主要功能

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine //@创建一个根 ctx 和一个可用于取消保留映射 goroutine 的取消函数
	rootCtx := context.Background() //@根ctx上下文背景
	ctx, cancel := context.WithCancel(rootCtx) //@ctx 使用取消根 ctx 取消上下文

	defer cancel() //@推迟取消

	setupAPI(ctx) //@设置 api 我 ctx

	// Serve on port :8080, fudge yeah hardcoded port //@在端口软糖上服务是的硬编码端口
	err := http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil) //@错误的 http 监听和服务 tl s 服务器 crt 服务器密钥 nil
	if err != nil { //@如果错误为零
		log.Fatal("ListenAndServe: ", err) //@记录致命的监听和服务错误
	}

}

// setupAPI will start all Routes and their Handlers //@设置 ap 我将启动所有路由及其处理程序
func setupAPI(ctx context.Context) { //@func setup ap i ctx context 上下文

	// Create a Manager instance used to handle WebSocket Connections //@创建用于处理 Web 套接字连接的管理器实例
	manager := NewManager(ctx) //@经理新经理ctx

	// Serve the ./frontend directory at Route / //@在路由中提供前端目录
	http.Handle("/", http.FileServer(http.Dir("./frontend"))) //@http 句柄 http 文件服务器 http dir 前端
	http.HandleFunc("/login", manager.loginHandler) //@http 句柄 func 登录管理器登录处理程序
	http.HandleFunc("/ws", manager.serveWS)

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, len(manager.clients)) //@fmt fprint w len 经理客户
	})
}
