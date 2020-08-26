package mmtp

import "net"

// Client represents a client connection to an MMTP server.
type Client struct {
	conn net.Conn
}

// Dial returns a new Client connected to an MMTP server at addr.
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	client := &Client{conn: conn}
	return client, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendMessage sends a message to the connection.
func (c *Client) SendMessage(msg *Message) {
	mw := MessageWriter{conn: c.conn}
	mw.WriteMessage(msg)
}

// ReceiveMessage reads the next message from the connection.
func (c *Client) ReceiveMessage() (*Message, error) {
	msg := new(Message)
	_, err := msg.ReadFrom(c.conn)
	return msg, err
}
