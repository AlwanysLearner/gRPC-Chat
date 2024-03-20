package chat

import (
	"context"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"sync"
)

type ChatServerImplement struct {
	UnimplementedChatServer
}

// 创建一个全局的连接映射，用于存储每个用户对应的 WebSocket 连接
var OnlineMap = make(map[int64]*websocket.Conn)
var mapLock sync.RWMutex // 用于保护连接映射的读写操作
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有的WebSocket连接请求
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c *ChatServerImplement) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	productid, consumerid, msg := req.Producter, req.Consumer, req.Msg
	resp := new(ChatResponse)
	resp.MessageId = "0"
	resp.Success = false
	resp.Error = ""
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 升级 HTTP 连接到 WebSocket
		mapLock.RLock()
		_, ok := OnlineMap[productid]
		mapLock.RUnlock()
		if !ok {
			wsConn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				resp.Error = err.Error()
				return
			}
			mapLock.Lock()
			OnlineMap[productid] = wsConn
			mapLock.Unlock()
		}
		mapLock.RLock()
		u, ok := OnlineMap[consumerid]
		mapLock.RUnlock()
		if ok {
			resp.Success = true
			u.WriteMessage(websocket.TextMessage, []byte(msg))
		} else {
			resp.Error = "用户不在线"
		}
	})
	return resp, nil
}

func InitChat() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterChatServer(s, &ChatServerImplement{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
