package p2p

import (
	"errors"
	"go-admin/pkg/util"
	"net"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

const pingPeriod = 5 * time.Second

type WebSocketConn struct {
	emission.Emitter
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
}


func NewWebSocketConn(socket *websocket.Conn) *WebSocketConn {
	var conn WebSocketConn
	conn.Emitter = *emission.NewEmitter()
	conn.socket = socket
	conn.mutex = new(sync.Mutex)
	conn.closed = false
	conn.socket.SetCloseHandler(func(code int, text string) error {
		util.Warnf("%s [%d]", text, code)
		conn.Emit("close", code, text)
		conn.closed = true
		return nil
	})
	return &conn
}

/***/
func (conn *WebSocketConn) RendMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = conn.socket
	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				util.Warnf("******************* conn ReadMessage Error ***********************")
				if c, k := err.(*websocket.CloseError); k {
					conn.Emit("close", c.Code, c.Text)
				} else {
					if c, k := err.(*net.OpError); k {
						conn.Emit("close", 1008, c.Error())
					}
				}
				close(stop)
				break
			}
			in <- msg
		}
	}()

	for {
		select {
		case _ = <-pingTicker.C:
			// util.Infof("heart-package")
			if err := conn.Send("{}"); err != nil {
				util.Errorf("******************* heartbeat ***********************")
				pingTicker.Stop()
				return
			}
		case message := <-in:
			{
				util.Infof("Recivied data: %s", message)
				conn.Emit("message", []byte(message))
			}
		case <-stop:
			return
		}
	}
}

/*
* Send |message| to the connection.
 */
func (conn *WebSocketConn) Send(message string) error {
	util.Infof("Send data: %s", message)
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if conn.closed {
		return errors.New("websocket: write closed")
	}
	return conn.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

// Close 断开socket
func (conn *WebSocketConn) Close()  {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if conn.closed == false {
		util.Infof("Close ws conn now ",conn)
		conn.socket.Close()
		conn.closed = true
	}else {
		util.Warnf("Transport already closed :", conn)
	}
}