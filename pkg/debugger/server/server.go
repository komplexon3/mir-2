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
//go:embed frontend/build

var frontend embed.FS

type Server struct {
	port            string
	EventWsServer   *wsserver.WsServer // TODO: should actually be private
	GodModeWsServer *wsserver.WsServer // TODO: should actually be private
	logger          logging.Logger
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
	http.HandleFunc("/ws", s.EventWsServer.HandleWs)
	http.HandleFunc("/godmodews", s.GodModeWsServer.HandleWs)
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
	return s.EventWsServer.HasConnections()
}

func (s *Server) SendEvent(msg []byte) {
	s.EventWsServer.SendChan <- wsserver.WsMessage{
		MessageType: websocket.TextMessage,
		Payload:     msg,
	}
}

func (s *Server) ReceiveEventAction() wsserver.WsMessage {
	msg := <-s.EventWsServer.RecvChan
	s.logger.Log(logging.LevelDebug, fmt.Sprintf("Received msg: %v", msg))
	return msg
}

func (s *Server) SendGodModeMessage(msg []byte) {
	s.GodModeWsServer.SendChan <- wsserver.WsMessage{
		MessageType: websocket.TextMessage,
		Payload:     msg,
	}
}

func (s *Server) ReceiveGodModeMessage() wsserver.WsMessage {
	msg := <-s.GodModeWsServer.RecvChan
	s.logger.Log(logging.LevelDebug, fmt.Sprintf("Received msg: %v", msg))
	return msg
}

func (s *Server) Close() {
	s.EventWsServer.Close()
}

func NewServer(port string, logger logging.Logger) *Server {
	ewss := wsserver.NewWsServer(logger)
	gmwss := wsserver.NewWsServer(logger)

	s := &Server{
		port:            port,
		EventWsServer:   ewss,
		GodModeWsServer: gmwss,
		logger:          logger,
	}

	return s
}
