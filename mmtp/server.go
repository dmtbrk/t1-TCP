package mmtp

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	Addr        string
	Handler     Handler
	IdleTimeout time.Duration
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":8080"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("ERROR: accepting connection:", err)
			continue
		}
		log.Println("INFO: connection accepted:", conn.RemoteAddr())

		go srv.Handle(conn)
	}
}

func (srv *Server) Handle(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("ERROR: closing connection:", err)
		} else {
			log.Println("INFO: connection closed:", conn.RemoteAddr())
		}
	}()

	br := bufio.NewReader(conn)
	mw := &MessageWriter{conn: conn}
	for {
		msg, _, err := ReadMessage(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("ERROR: parsing message from %v: %v", conn.RemoteAddr(), err)
			continue
		}

		srv.Handler.ServeMMTP(mw, msg)
	}
}

type MessageWriter struct {
	conn net.Conn
}

func (mw *MessageWriter) WriteMessage(msg *Message) error {
	bw := bufio.NewWriter(mw.conn)
	_, err := msg.WriteTo(bw)
	if err != nil {
		return err
	}
	return bw.Flush()
}

type Handler interface {
	ServeMMTP(mw *MessageWriter, msg *Message)
}

type HandlerFunc func(mw *MessageWriter, msg *Message)

func (f HandlerFunc) ServeMMTP(mw *MessageWriter, msg *Message) {
	f(mw, msg)
}

func ListenAndServe(addr string, handler Handler) error {
	srv := &Server{Addr: addr, Handler: handler}
	return srv.ListenAndServe()
}
