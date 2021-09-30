package ws

import (
	"encoding/json"
	"net/url"
	"public/libs_go/gateway/master"
	"reflect"
	"sync"
	"sync/atomic"

	"public/libs_go/logs"

	"github.com/gorilla/websocket"
	sutil "github.com/souliot/siot-util"
)

var (
	Ws *Hub
)

type cType int

const (
	WsAuth cType = iota
	WsLog
)

func InitWS() {
	Ws = newHub()
}

type wsAuth struct {
	Authorization string
	Appid         string
	Secret        string
}

func (m *wsAuth) valid() (b bool) {
	return true
}

type Hub struct {
	clients    *sync.Map
	keyClients *sync.Map
}

func newHub() *Hub {
	return &Hub{
		clients:    new(sync.Map),
		keyClients: new(sync.Map),
	}
}

func (m *Hub) AddClient(mt cType, c *websocket.Conn, id string, key string) {
	wc := newWsClient(mt, c, id, key)
	_, loaded := m.clients.LoadOrStore(id, wc)
	if !loaded {
		wc.Do()
	}

	if v, ok := m.keyClients.Load(key); ok {
		if _, loaded := v.(*sync.Map).LoadOrStore(id, wc); !loaded {
			id_cache := new(sync.Map)
			id_cache.Store(id, wc)
			m.keyClients.Store(key, id_cache)
		}
	} else {
		id_cache := new(sync.Map)
		id_cache.Store(id, wc)
		m.keyClients.Store(key, id_cache)
	}
}

func (m *Hub) DelClient(c *wsClient) {
	m.clients.Delete(c.id)
	if v, ok := m.keyClients.Load(c.key); ok {
		id_cache := v.(*sync.Map)
		id_cache.Delete(c.id)
		m.keyClients.Store(c.key, id_cache)
	}
}

func (m *Hub) Pub(msg Message) {
	m.clients.Range(func(k, v interface{}) bool {
		if c, ok := v.(*wsClient); ok {
			if c.wsType == msg.DataType && c.auth {
				c.sendchan <- msg
			}
		}
		return true
	})
}

func (m *Hub) SendMessageToID(id string, msg Message) {
	if v, loaded := m.clients.Load(id); loaded {
		if c, ok := v.(*wsClient); ok {
			if c.wsType == msg.DataType && c.auth && !c.isClosed() {
				c.sendchan <- msg
			}
		}
	}
}

func (m *Hub) SendMessageToKEY(key string, msg Message) {
	if v, loaded := m.keyClients.Load(key); loaded {
		if c, ok := v.(*sync.Map); ok {
			c.Range(func(id, iws interface{}) bool {
				if ws, ok := iws.(*wsClient); ok {
					if ws.wsType == msg.DataType && ws.auth && !ws.isClosed() {
						ws.sendchan <- msg
					}
				}
				return true
			})
		}
	}
}

type Message struct {
	Data     interface{} `json:"data"`
	DataType cType       `json:"data_type"`
}

func NewMessage(mt cType, data interface{}) Message {
	return Message{
		Data:     data,
		DataType: mt,
	}
}

type wsClient struct {
	id        string
	key       string
	ws        *Hub
	wsType    cType
	conn      *websocket.Conn
	auth      bool
	closeFlag int32
	sendchan  chan Message
	readchan  chan Message
	closeOnce sync.Once
}

func newWsClient(mt cType, c *websocket.Conn, id, key string) *wsClient {
	return &wsClient{
		id:       id,
		key:      key,
		ws:       Ws,
		wsType:   mt,
		conn:     c,
		auth:     true,
		sendchan: make(chan Message),
		readchan: make(chan Message),
	}
}

func (m *wsClient) Do() {
	go m.handelLoop()
	go m.readLoop()
	go m.writeLoop()
}

func (m *wsClient) close() {
	m.closeOnce.Do(func() {
		atomic.StoreInt32(&m.closeFlag, 1)
		close(m.sendchan)
		close(m.readchan)
		m.conn.Close()

		m.ws.DelClient(m)
	})
}

func (m *wsClient) isClosed() bool {
	return atomic.LoadInt32(&m.closeFlag) == 1
}

func (m *wsClient) handelLoop() {
	defer func() {
		recover()
		m.close()
	}()
	for {
		select {
		case p := <-m.readchan:
			if m.isClosed() {
				return
			}
			//
			if !m.auth {
				wa := new(wsAuth)
				bs := reflect.ValueOf(p.Data).Bytes()
				err := json.Unmarshal(bs, wa)
				if err != nil {
					continue
				}
				// 检测认证信息
				valid := wa.valid()
				if valid {
					m.auth = true
					if m.id == "" {
						u, err := url.Parse("//" + m.conn.RemoteAddr().String())
						if err != nil {
							m.id = master.GetID()
						} else {
							m.id = sutil.GetMd5String(u.Hostname())
						}
					}
					m.sendchan <- Message{
						Data:     sutil.Actionsuccess,
						DataType: WsAuth,
					}
				} else {
					continue
				}
			}
		}
	}
}

func (m *wsClient) readLoop() {
	defer func() {
		recover()
		m.close()
	}()
	for {
		_, data, err := m.conn.ReadMessage()
		if err != nil {
			return
		}
		m.readchan <- NewMessage(m.wsType, data)
	}
}

func (m *wsClient) writeLoop() {
	defer func() {
		recover()
		m.close()
	}()
	for {
		select {
		case p := <-m.sendchan:
			if m.isClosed() {
				return
			}
			err := m.conn.WriteJSON(p.Data)
			if err != nil {
				logs.Error("发送数据错误：", err)
				return
			}
		}
	}
}
