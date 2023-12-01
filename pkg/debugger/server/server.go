package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	wsserver "github.com/filecoin-project/mir/pkg/debugger/server/wsServer"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/gorilla/websocket"
)

// files to be served
//go:embed frontend

var frontend embed.FS

type Server struct {
	port     string
	WsServer *wsserver.WsServer // TODO: should actually be private
	logger   logging.Logger
}

func getBaseFS() fs.FS {
	if fs, err := fs.Sub(frontend, "frontend"); err != nil {
		panic(err)
	} else {
		return fs
	}
}

func (s *Server) setupHttpRoutes() {
	fmt.Println("Setting up routes")
	http.HandleFunc("/ws", s.WsServer.HandleWs)
	http.Handle("/", http.FileServer(http.FS(getBaseFS())))
}

func (s *Server) StartServer() {
	s.logger.Log(logging.LevelInfo, "Starting servers...")
	s.setupHttpRoutes()
	if err := http.ListenAndServe(s.port, nil); err != nil {
		s.logger.Log(logging.LevelError, "ListenAndServe: ", err)
	}
}

func (s *Server) HasWSConnections() bool {
	return s.WsServer.HasConnections()
}

func (s *Server) SendMessage(msg []byte) {
	s.WsServer.SendChan <- wsserver.WsMessage{
		MessageType: websocket.TextMessage,
		Payload:     msg,
	}
}

func (s *Server) ReceiveAction() wsserver.WsMessage {
	msg := <-s.WsServer.RecvChan
	s.logger.Log(logging.LevelDebug, fmt.Sprintf("Received msg: %v", msg))
	return msg
}

func (s *Server) Close() {
	s.WsServer.Close()
}

func NewServer(port string, logger logging.Logger) *Server {
	wss := wsserver.NewWsServer(logger)

	s := &Server{
		port:     port,
		WsServer: wss,
		logger:   logger,
	}

	return s
}
