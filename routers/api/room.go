package api

import (
	"go-admin/pkg/signaler"
	"go-admin/service/p2p"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketServer a
type WebSocketServer struct {
	handleWebSocket  func(ws *p2p.WebSocketConn, request *http.Request)
	handleTurnServer func(writer http.ResponseWriter, request *http.Request)
	// Websocket upgrader
	upgrader websocket.Upgrader
}

// WsHandler wsq请求
func WsHandler(c *gin.Context) {
	//应答客户端告知升级为websocket
	wsSocket, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	wsTransport := p2p.NewWebSocketConn(wsSocket)
	signaler.NewSignaler().HandleNewWebSocket(wsTransport)
	wsTransport.RendMessage()

}
