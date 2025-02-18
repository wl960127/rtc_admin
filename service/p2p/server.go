package p2p

// import (
// 	"errors"
// 	"go-admin/pkg/util"
// 	"sync"
// 	"time"

// 	"github.com/chuckpreslar/emission"
// 	"github.com/gorilla/websocket"
// )

// const (
// 	// Time allowed to write a message to the peer.

// 	writeWait = 10 * time.Second
// 	// Time allowed to read the next pong message from the peer.

// 	pongWait = 60 * time.Second
// 	// Send pings to peer with this period. Must be less than pongWait
// 	pingPeriod     = (pongWait * 9) / 10
// 	maxMessageSize = 512

// 	//JoinRoom 加入房间
// 	JoinRoom = "joinRoom"
// 	//Offer 消息
// 	Offer = "offer"
// 	//Answer 消息
// 	Answer = "answer"
// 	//Candidate 消息
// 	Candidate = "candidate"
// 	//HangUp 挂断
// 	//LeaveRoom 离开房间
// 	HangUp = "hangUp"
// 	LeaveRoom = "leaveRoom"
// 	//UpdateUserList 更新房间用户列表
// 	UpdateUserList = "updateUserList"
// )

// var manager *RoomManager
// var once sync.Once

// //wsConnection 客户端连接
// type wsConnection struct {
// 	wsSocket *websocket.Conn
// 	inChan   chan *wsMessage //读队列
// 	outChan  chan *wsMessage

// 	mutex     sync.Mutex // 避免重复关闭管道
// 	isClosed  bool
// 	closeChan chan byte //关闭通知
// }

// type wsMessage struct {
// 	messageType int
// 	data        []byte
// }

// // MsgHandler 消息处理
// func MsgHandler(wsSocket *websocket.Conn) {
// 	wsConn := &wsConnection{
// 		wsSocket:  wsSocket,
// 		inChan:    make(chan *wsMessage, 1000),
// 		outChan:   make(chan *wsMessage, 1000),
// 		closeChan: make(chan byte),
// 		isClosed:  false,
// 	}

// 	//处理器
// 	go wsConn.procLoop()
// 	//读协程
// 	go wsConn.wsReadLooop()
// 	//写协程
// 	go wsConn.wsWriteLoop()
// }

// func (wsConn *wsConnection) procLoop() {
// 	go func() {
// 		for {
// 			time.Sleep(2 * time.Second)
// 			// util.Infof("-----heartbeat-----")
// 			//发送空包
// 			heartPackage := map[string]interface{}{
// 				//消息类型
// 				"msgType": "heartPackage",
// 				//空数据包
// 				"data": "",
// 			}
// 			if err := wsConn.wsWrite(websocket.TextMessage, []byte(util.Marshal(heartPackage))); err != nil {
// 				util.Errorf("heartbeat fail")
// 				wsConn.wsClose()
// 				break
// 			}
// 		}
// 	}()
// 	// 同步处理模型,需要并行的话 可以每个请求一个goruntine
// 	for {
// 		msg, err := wsConn.wsRead()
// 		if err != nil {
// 			util.Errorf("read fail")
// 			break
// 		}
// 		util.Infof(string(msg.data))
// 		// 处理数据
// 		requestData, err := util.Unmarshal(string(msg.data))
// 		if err != nil {
// 			util.Errorf("解析Json数据Unmarshal错误 %v", err)
// 			return
// 		}
// 		//定义数据
// 		var data map[string]interface{} = nil
// 		// 获取"data"这个key的具体数据
// 		tmp, found := requestData["data"]
// 		//如果没有找到数据输出日志
// 		if !found {
// 			util.Errorf("没有发现数据!")
// 			return
// 		}
// 		data = tmp.(map[string]interface{})
// 		// 获取roomId
// 		roomID := data["roomId"].(string)
// 		util.Infof("房间Id: %v", roomID)

// 		//创建Room
// 		roomManager := GetRoomManager()
// 		room := roomManager.getRoom(roomID)
// 		//查询不到房间则创建一个房间
// 		if room == nil {
// 			room = roomManager.createRoom(roomID)
// 		}

// 		// 判断消息类型
// 		switch requestData["msgType"] {
// 		case JoinRoom:
// 			onJoinRoom(wsConn, data, room, roomManager)
// 			break
// 			//提议Offer消息
// 		case Offer:
// 			//直接执行下一个case并转发消息
// 			fallthrough
// 		//应答Answer消息
// 		case Answer:
// 			//直接执行下一个case并转发消息
// 			fallthrough
// 		//网络信息Candidate
// 		case Candidate:
// 			onCandidate(wsConn, data, room, roomManager, requestData)
// 			break
// 		//挂断消息
// 		case HangUp:
// 			util.Infof(" 挂断消息 " + util.Marshal(data))

// 			// onHangUp(conn,data,room, roomManager,request)
// 			break
// 		default:
// 			{
// 				util.Warnf("未知的请求 %v", requestData)
// 			}
// 			break
// 		}

// 		err = wsConn.wsWrite(msg.messageType, msg.data)
// 		if err != nil {
// 			util.Errorf("write fail")
// 			break
// 		}
// 	}

// }

// func onJoinRoom(conn *wsConnection, data map[string]interface{}, room *Room, roomManager *RoomManager) {
// 	//创建一个User
// 	user := User{
// 		//连接
// 		conn: conn,
// 		//User信息
// 		info: UserInfo{
// 			ID:   data["id"].(string),   //ID值
// 			Name: data["name"].(string), //名称
// 		},
// 	}
// 	util.Debugf("用户房间用户信息 %v", user.info)
// 	//把User放入数组里
// 	room.users[user.info.ID] = user

// 	//通知所有的User更新
// 	roomManager.notifyUsersUpdate(conn, room.users)
// }

// //offer/answer/candidate消息处理
// func onCandidate(conn *wsConnection, data map[string]interface{}, room *Room, roomManager *RoomManager, request map[string]interface{}) {
// 	//读取目标to属性值
// 	to := data["to"].(string)
// 	//查找User对象
// 	if user, ok := room.users[to]; !ok {
// 		util.Errorf("没有发现用户[" + to + "]")
// 		return
// 	} else {
// 		//发送信息给目标User
// 		util.Infof("发送信息给目标User %v",util.Marshal(request))
// 		if err := user.conn.wsWrite(websocket.TextMessage, []byte(util.Marshal(request))); err != nil {
// 			util.Errorf("onCandidate 通知失败")
// 		}
// 	}
// }

// //通知所有的用户更新
// func (roomManager *RoomManager) notifyUsersUpdate(conn *wsConnection, users map[string]User) {
// 	//更新信息
// 	infos := []UserInfo{}
// 	//迭代所有的User
// 	for _, userClient := range users {
// 		//添加至数组里
// 		infos = append(infos, userClient.info)
// 	}
// 	//创建发送消息数据结构
// 	request := make(map[string]interface{})
// 	//消息类型
// 	request["msgType"] = UpdateUserList
// 	//数据
// 	request["data"] = infos

// 	util.Infof(" 更新用户通知 " + util.Marshal(request))

// 	//迭代所有的User
// 	for _, user := range users {
// 		//将Json数据发送给每一个User
// 		if err := user.conn.wsWrite(websocket.TextMessage, []byte(util.Marshal(request))); err != nil {
// 			util.Errorf("通知用户更新失败")
// 		}
// 	}
// }

// func (wsConn *wsConnection) wsReadLooop() {
// 	for {
// 		msgType, data, err := wsConn.wsSocket.ReadMessage()
// 		if err != nil {
// 			goto error
// 		}
// 		util.Infof(string(data))
// 		req := &wsMessage{
// 			msgType,
// 			data,
// 		}

// 		select {
// 		case wsConn.inChan <- req:
// 		case <-wsConn.closeChan:
// 			goto closed
// 		}
// 	}
// error:
// 	wsConn.wsClose()
// closed:
// }

// func (wsConn *wsConnection) wsWriteLoop() {
// 	for {
// 		select {
// 		// 取一个应答
// 		case msg := <-wsConn.outChan:
// 			// 写给websocket
// 			if err := wsConn.wsSocket.WriteMessage(msg.messageType, msg.data); err != nil {
// 				goto error
// 			}
// 		case <-wsConn.closeChan:
// 			goto closed
// 		}
// 	}
// error:
// 	wsConn.wsClose()
// closed:
// }

// func (wsConn *wsConnection) wsWrite(messageType int, data []byte) error {
// 	select {
// 	case wsConn.outChan <- &wsMessage{messageType, data}:
// 	case <-wsConn.closeChan:
// 		return errors.New("websocket closed")
// 	}
// 	return nil
// }

// func (wsConn *wsConnection) wsRead() (*wsMessage, error) {
// 	select {
// 	case msg := <-wsConn.inChan:
// 		return msg, nil
// 	case <-wsConn.closeChan:
// 	}
// 	return nil, errors.New("websocket closed")
// }

// func (wsConn *wsConnection) wsClose() {
// 	wsConn.wsSocket.Close()

// 	wsConn.mutex.Lock()
// 	defer wsConn.mutex.Unlock()
// 	if !wsConn.isClosed {
// 		wsConn.isClosed = true
// 		close(wsConn.closeChan)
// 	}
// }

// //RoomManager 定义房间
// type RoomManager struct {
// 	rooms map[string]*Room
// }

// // GetRoomManager 实例化房间管理对象
// func GetRoomManager() *RoomManager {
// 	once.Do(func() {
// 		manager = &RoomManager{
// 			rooms: make(map[string]*Room),
// 		}
// 	})
// 	return manager
// }

// // Room 定义房间
// type Room struct {
// 	//所有用户
// 	users map[string]User
// 	//所有会话
// 	sessions map[string]Session
// 	ID       string
// }

// //NewRoom 实例化房间对象
// func NewRoom(id string) *Room {
// 	var room = &Room{
// 		users:    make(map[string]User),
// 		sessions: make(map[string]Session),
// 		ID:       id,
// 	}
// 	return room
// }

// //获取房间
// func (roomManager *RoomManager) getRoom(id string) *Room {
// 	return roomManager.rooms[id]
// }

// //创建房间
// func (roomManager *RoomManager) createRoom(id string) *Room {
// 	util.Infof("创建房间 " + id)
// 	roomManager.rooms[id] = NewRoom(id)
// 	return roomManager.rooms[id]
// }

// //删除房间
// func (roomManager *RoomManager) deleteRoom(id string) {
// 	delete(roomManager.rooms, id)
// }

// //UserInfo 用户信息
// type UserInfo struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// }

// //Session 会话信息
// type Session struct {
// 	id   string
// 	from User
// 	to   User
// }

// // User 用户
// type User struct {
// 	info UserInfo
// 	// conn
// 	conn *wsConnection
// }


// type WebSocketConn struct {
// 	emission.Emitter
// 	socket *websocket.Conn
// 	mutex  *sync.Mutex
// 	closed bool
// }