package mtp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Message struct {
	Type    int
	Payload string
}

// Write writes a message in wire format. It implements the io.WriterTo interface.
func (m *Message) WriteTo(w io.Writer) (n int64, err error) {
	// Wrap the writer in a bufio Writer if it's not already buffered.
	var bw *bufio.Writer
	if _, ok := w.(io.ByteWriter); !ok {
		bw = bufio.NewWriter(w)
		w = bw
	}

	// Write the type.
	nt, err := fmt.Fprintf(w, "%d\r\n", m.Type)
	n += int64(nt)
	if err != nil {
		return 0, err
	}

	// Write the payload.
	np, err := fmt.Fprintf(w, "%s\r\n", m.Payload)
	n += int64(np)
	if err != nil && err != io.EOF {
		return 0, err
	}

	// Finish.
	if bw, ok := w.(*bufio.Writer); ok {
		err = bw.Flush()
		if err != nil {
			return 0, err
		}
	}
	return n, nil
}

func ReadMessage(br *bufio.Reader) (*Message, error) {
	msg := new(Message)
	
	s, err := readLine(br)
	if err != nil {
		return nil, err
	}
	msg.Type, err = parseMessageType(s)
	if err != nil {
		return nil, err
	}

	s, err = readLine(br)
	msg.Payload = s

	return msg, nil
}

func parseMessageType(s string) (int, error) {
	return strconv.Atoi(s)
}

// readLine reads a line and returns it ommiting new line characters.
func readLine(br *bufio.Reader) (string, error) {
	var line []byte
	for {
		l, isPrefix, err := br.ReadLine()
		if err != nil {
			return "", err
		}

		// First ReadLine call produced a full line. Don't need to copy.
		if line == nil && !isPrefix {
			return string(l), nil
		}

		// More than one ReadLine call is needed.
		line = append(line, l...)
		if !isPrefix {
			break
		}
	}
	return string(line), nil
}

// type MessageType int

// const (
// 	MessageInvalid MessageType = iota
// 	MessageRequestItem
// 	MessageRequestList
// 	MessageResponseItem
// 	MessageResponseList
// 	MessageError
// )
