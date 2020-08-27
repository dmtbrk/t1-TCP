package mtp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	Addr        string
	Handler     Handler
	IdleTimeout time.Duration

	listener   net.Listener
	conns      map[net.Conn]bool
	inShutdown bool

	mu sync.Mutex
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
	srv.listener = ln
	return srv.Serve()
}

func (srv *Server) Serve() error {
	for !srv.inShutdown { // http uses atomicBool
		conn, err := srv.listener.Accept()
		if err != nil {
			if srv.inShutdown {
				break
			}
			log.Println("ERROR: accepting connection:", err)
			continue
		}
		conn = newConn(conn, srv.IdleTimeout)
		srv.addConn(conn)
		log.Println("INFO: connection accepted:", conn.RemoteAddr())

		go srv.Handle(conn)
	}
	return nil
}

func (srv *Server) Handle(conn net.Conn) {
	defer func() {
		conn.Close()
		srv.deleteConn(conn)
		log.Println("INFO: connection closed:", conn.RemoteAddr())
	}()

	br := bufio.NewReader(conn)
	mw := NewMessageWriter(conn)
	for {
		msg, err := ReadMessage(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Printf("INFO: deadline exceeded: %v", conn.RemoteAddr())
				break
			}
			log.Printf("ERROR: parsing message from %v: %v", conn.RemoteAddr(), err)
			continue
		}

		srv.Handler.ServeMMTP(mw, msg)
	}
}

func (srv *Server) Shutdown() {
	log.Println("INFO: shutting down...")
	srv.inShutdown = true // http uses atomicBool
	srv.listener.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("INFO: waiting for %d connections to disconnect", len(srv.conns))
		default:
			if len(srv.conns) == 0 {
				return
			}
		}
	}
}

func (srv *Server) addConn(conn net.Conn) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.conns == nil {
		srv.conns = make(map[net.Conn]bool)
	}

	srv.conns[conn] = true
}

func (srv *Server) deleteConn(conn net.Conn) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	delete(srv.conns, conn)
}

type conn struct {
	net.Conn
	IdleTimeout time.Duration
}

func newConn(c net.Conn, idleTimeout time.Duration) *conn {
	return &conn{Conn: c, IdleTimeout: idleTimeout}
}

func (c *conn) Write(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Write(p)
}

func (c *conn) Read(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Read(p)
}

func (c *conn) updateDeadline() {
	t := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(t)
}

type MessageWriter struct {
	w io.Writer
}

func NewMessageWriter(w io.Writer) *MessageWriter {
	return &MessageWriter{w: w}
}

func (mw *MessageWriter) WriteMessage(msg *Message) error {
	_, err := msg.WriteTo(mw.w)
	return err
}

type Handler interface {
	ServeMMTP(mw *MessageWriter, msg *Message)
}

type HandlerFunc func(mw *MessageWriter, msg *Message)

func (f HandlerFunc) ServeMMTP(mw *MessageWriter, msg *Message) {
	f(mw, msg)
}

func ListenAndServe(addr string, handler Handler) error {
	srv := &Server{
		Addr:        addr,
		IdleTimeout: 5 * time.Second,
		Handler:     handler,
	}
	return srv.ListenAndServe()
}
