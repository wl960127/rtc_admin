package signaler

import (
	"go-admin/pkg/util"
	"go-admin/service/p2p"
	"strings"
	"sync"
)

var signaler *Signaler
var once sync.Once

// PeerInfo 同伴信息
type PeerInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UserAgent string `json:"user_agent"`
}

// Peer peer
type Peer struct {
	info PeerInfo
	conn *p2p.WebSocketConn
}

// Session 会话
type Session struct {
	id   string
	from Peer
	to   Peer
}

// Signaler 信令
type Signaler struct {
	peers    map[string]Peer
	sessions map[string]Session
	// expresMap *uti
}

// NewSignaler aaa
func NewSignaler() *Signaler {
	once.Do(func() {
		signaler = &Signaler{
			peers:    make(map[string]Peer),
			sessions: make(map[string]Session),
		}
	})
	return signaler
}

// NotifyPeersUpdate .
func (s *Signaler) NotifyPeersUpdate(conn *p2p.WebSocketConn, peers map[string]Peer) {
	infos := []PeerInfo{}
	for _, peer := range peers {
		infos = append(infos, peer.info)
	}
	request := make(map[string]interface{})
	request["type"] = "peers"
	request["data"] = infos
	for _, peer := range peers {
		peer.conn.Send(util.Marshal(request))
	}

}

// HandleNewWebSocket .
func (s *Signaler) HandleNewWebSocket(conn *p2p.WebSocketConn) {
	conn.On("message", func(message []byte) {
		request, err := util.Unmarshal(string(message))
		if err != nil {
			util.Errorf("Unmarshal error %v", err)
			return
		}
		util.Infof("收到信息 %v", request)
		var data map[string]interface{} = nil
		tmp, found := request["data"]
		if !found {
			util.Errorf("******************* No data sturct found ***********************")
			return
		}
		data = tmp.(map[string]interface{})
		switch request["type"] {
		case "new":
			peer := Peer{
				conn: conn,
				info: PeerInfo{
					ID:   data["id"].(string),
					Name: data["name"].(string),
				},
			}
			s.peers[peer.info.ID] = peer
			s.NotifyPeersUpdate(conn, s.peers)
			break
		case "leave":
			{
			}
			break
		case "offer":
			fallthrough
		case "answer":
			fallthrough
		case "candidate":
			{
				if to, ok := data["to"]; !ok || to == nil {
					util.Errorf("No to id found!")
					return
				}

				to := data["to"].(string)
				if peer, ok := s.peers[to]; !ok {
					msg := map[string]interface{}{
						"type": "error",
						"data": map[string]interface{}{
							"request": request["type"],
							"reason":  "Peer [" + to + "] not found ",
						},
					}
					conn.Send(util.Marshal(msg))
					return
				} else {
					peer.conn.Send(util.Marshal(request))
				}
			}
			break
		case "bye":
			{
				if id, ok := data["session_id"]; !ok || id == nil {
					util.Errorf("No session_id found!")
					return
				}

				sessionID := data["session_id"].(string)
				ids := strings.Split(sessionID, "-")
				if len(ids) != 2 {
					msg := map[string]interface{}{
						"type": "error",
						"data": map[string]interface{}{
							"request": request["type"],
							"reason":  "Invalid session [" + sessionID + "]",
						},
					}
					conn.Send(util.Marshal(msg))
					return
				}
				if peer, ok := s.peers[ids[0]]; !ok {
					msg := map[string]interface{}{
						"type": "error",
						"data": map[string]interface{}{
							"request": request["type"],
							"reason":  "Peer [" + ids[0] + "] not found.",
						},
					}
					conn.Send(util.Marshal(msg))
					return
				} else {
					bye := map[string]interface{}{
						"type": "bye",
						"data": map[string]interface{}{
							"to":         ids[0],
							"session_id": sessionID,
						},
					}
					peer.conn.Send(util.Marshal(bye))
				}

				if peer, ok := s.peers[ids[1]]; !ok {
					msg := map[string]interface{}{
						"type": "error",
						"data": map[string]interface{}{
							"request": request["type"],
							"reason":  "Peer [" + ids[0] + "] not found ",
						},
					}
					conn.Send(util.Marshal(msg))
					return
				} else {
					bye := map[string]interface{}{
						"type": "bye",
						"data": map[string]interface{}{
							"to":         ids[1],
							"session_id": sessionID,
						},
					}
					peer.conn.Send(util.Marshal(bye))
				}
			}
			break
		case "heartbeat":
			heartbeat := map[string]interface{}{
				"type": "heartbeat",
				"data": map[string]interface{}{},
			}
			conn.Send(util.Marshal(heartbeat))
			break
		default:
			util.Errorf("******************* Unkown request %v ***********************", request)
			break
		}

	})

	conn.On("close", func(code int, text string) {
		util.Infof("On Close %v", conn)
		var peerID string = ""
		for _, peer := range s.peers {

			if peer.conn == conn {
				peerID = peer.info.ID
			} else {
				leave := map[string]interface{}{
					"type": "leave",
					"data": peer.info.ID,
				}
				peer.conn.Send(util.Marshal(leave))
			}
		}

		util.Infof("Remove peer %s", peerID)
		if peerID == "" {
			util.Infof("Leve peer id not found")
			return
		}
		delete(s.peers, peerID)

		s.NotifyPeersUpdate(conn, s.peers)

	})
}
