package api

import (
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


// WsHandler wsq请求
func WsHandler(c *gin.Context) {
	//应答客户端告知升级为websocket
	wsSocket, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// wsConn := &wsConnection{
	// 	wsSocket:  wsSocket,
	// 	inChan:    make(chan *wsMessage, 1000),
	// 	outChan:   make(chan *wsMessage, 1000),
	// 	closeChan: make(chan byte),
	// 	isClosed:  false,
	// }
	p2p.MsgHandler(wsSocket)
}
