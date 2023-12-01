package wsserver

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/filecoin-project/mir/pkg/logging"
)

type WsMessage struct {
	MessageType int
	Payload     []byte
}

type connection struct {
	id       string
	server   *WsServer
	conn     *websocket.Conn
	sendChan chan WsMessage
	context  context.Context
	cancel   context.CancelFunc
}

type WsServer struct {
	SendChan        chan WsMessage
	RecvChan        chan WsMessage
	connections     map[string]connection
	connectionsLock sync.Mutex
	logger          logging.Logger
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (wss *WsServer) HandleWs(w http.ResponseWriter, r *http.Request) {
	// cors *
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wss.logger.Log(logging.LevelError, "Upgrade failed: ", err)
		return
	}
	wss.connectionsLock.Lock()
	id := ""
	for {
		id = uuid.New().String()
		if _, ok := wss.connections[id]; !ok {
			break
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	conn := connection{id, wss, ws, make(chan WsMessage), ctx, cancel}
	wss.connections[id] = conn
	wss.connectionsLock.Unlock()
	go conn.wsConnection()
}

func (wsc *connection) closeConnection() {
	wsc.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	wsc.conn.Close()
	wsc.server.connectionsLock.Lock()
	wsc.cancel()
	close(wsc.sendChan)
	delete(wsc.server.connections, wsc.id)
	wsc.server.connectionsLock.Unlock()
}

func (wsc *connection) close() {
	wsc.cancel()
}

func (wsc *connection) wsConnection() {
	wsc.server.logger.Log(logging.LevelInfo, "New WS connection", "total connections: ", len(wsc.server.connections))
	//  close connection
	defer wsc.closeConnection()

	// sending msgs
	go func() {
		for {
			select {
			case <-wsc.context.Done():
				return
			case message := <-wsc.sendChan:
				err := wsc.conn.WriteMessage(message.MessageType, message.Payload)
				if err != nil {
					wsc.server.logger.Log(logging.LevelError, "Error sending WS message: ", err)
				}
			}
		}
	}()

	// "receiving" msgs
	for {
		msgType, payload, err := wsc.conn.ReadMessage()

		// catching all erros but actually only interested in "going away" errors
		if msgType != websocket.TextMessage || err != nil {
			wsc.server.logger.Log(logging.LevelWarn, "WS connection closed")
			return
		}

		wsc.server.RecvChan <- WsMessage{msgType, payload}
	}
}

func (wss *WsServer) sendHandler() {
	for {
		message := <-wss.SendChan
		wss.connectionsLock.Lock()
		for _, conn := range wss.connections {
			conn.sendChan <- message
		}
		wss.connectionsLock.Unlock()
	}
}

func (wss *WsServer) HasConnections() bool {
	wss.connectionsLock.Lock()
	defer wss.connectionsLock.Unlock()
	return len(wss.connections) > 0
}

func (wss *WsServer) Close() {
	wss.connectionsLock.Lock()
	defer wss.connectionsLock.Unlock()
	for _, conn := range wss.connections {
		conn.close()
	}
	close(wss.SendChan)
}

func NewWsServer(logger logging.Logger) *WsServer {
	wss := &WsServer{
		connections: make(map[string]connection),
		SendChan:    make(chan WsMessage),
		RecvChan:    make(chan WsMessage),
		logger:      logger,
	}

	go wss.sendHandler()

	return wss
}
