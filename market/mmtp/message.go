package mmtp

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ortymid/t1-tcp/market"
)

type Message struct {
	Type    MessageType
	Payload interface{}
}

func (msg *Message) ReadFrom(r io.Reader) (n int64, err error) {
	br := bufio.NewReader(r)
	m, n, err := ReadMessage(br)
	if err != nil {
		return n, err
	}
	msg.Type = m.Type
	msg.Payload = m.Payload
	return n, err
}

func (msg *Message) WriteTo(w io.Writer) (n int64, err error) {
	nt, err := fmt.Fprintf(w, "%d\n", msg.Type) // write type
	n += int64(nt)
	if err != nil {
		return n, err
	}
	switch msg.Type {
	case MessageProductRequest:
		id, ok := msg.Payload.(int)
		if !ok {
			return n, fmt.Errorf("expected int payload type, got %T", msg.Payload)
		}
		npi, err := fmt.Fprintf(w, "%d\n", id)
		n += int64(npi)
		if err != nil {
			return n, err
		}
	case MessageProduct, MessageProductAdd:
		p, ok := msg.Payload.(*market.Product)
		if !ok {
			return n, fmt.Errorf("expected *market.Product payload type, got %T", msg.Payload)
		}
		np, err := writeProduct(w, p)
		n += int64(np)
		if err != nil {
			return n, err
		}
	case MessageProductList:
		ps, ok := msg.Payload.([]*market.Product)
		if !ok {
			return n, fmt.Errorf("expected []*market.Product payload type, got %T", msg.Payload)
		}
		for _, p := range ps {
			np, err := writeProduct(w, p) // write a product as csv
			n += int64(np)
			if err != nil {
				return n, err
			}
		}
		n3, err := w.Write([]byte{'\n'}) // write the end of the body
		n += int64(n3)
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func writeProduct(w io.Writer, p *market.Product) (int, error) {
	return fmt.Fprintf(w, "%d,%q,%d\n", p.ID, p.Name, p.Price)
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{ Type: %d, Payload: %v }", msg.Type, msg.Payload)
}

type MessageType int

const (
	MessageInvalid            MessageType = iota // invalid message type
	MessageProductRequest                        // request for a product by id
	MessageProduct                               // contains a single product (product request or create response)
	MessageProductListRequest                    // request for all products
	MessageProductList                           // contains a list of products
	MessageProductAdd                            // contains product data to add to the list
	MessageError                                 // contains an error
)

// ReadMessage reads and parses an incoming message from b.
func ReadMessage(br *bufio.Reader) (msg *Message, n int64, err error) {
	msg = new(Message)
	// reading the message type
	line, err := br.ReadString('\n')
	n += int64(len(line))
	if err != nil {
		return nil, n, err
	}
	line = strings.TrimSuffix(line, "\n")
	msg.Type, err = parseMessageType(line)
	if err != nil {
		return nil, n, err
	}

	// reading the message body
	switch msg.Type {
	case MessageProduct, MessageProductAdd:
		p, np, err := readProduct(br)
		n += np
		if err != nil {
			return nil, n, err
		}
		msg.Payload = p
	case MessageProductRequest:
		id, npi, err := readProductID(br)
		n += npi
		if err != nil {
			return nil, n, err
		}
		msg.Payload = id
	case MessageProductList:
		products, npl, err := readProductList(br)
		n += npl
		if err != nil {
			return nil, n, err
		}
		msg.Payload = products
	case MessageProductListRequest:
		msg.Payload = nil // may be filtering parameters some day
	}

	return msg, n, nil
}

func parseMessageType(line string) (MessageType, error) {
	typ, err := strconv.Atoi(line)
	if err != nil {
		return MessageInvalid, fmt.Errorf("bad message type: %w", err)
	}
	return MessageType(typ), nil
}

func readProductID(br *bufio.Reader) (id int, n int64, err error) {
	line, err := br.ReadString('\n')
	n += int64(len(line))
	if err != nil {
		return id, n, err
	}
	line = strings.TrimSuffix(line, "\n")
	id, err = strconv.Atoi(line)
	if err != nil {
		return id, n, fmt.Errorf("bad product id: %w", err)
	}
	return id, n, err
}

func readProduct(br *bufio.Reader) (p *market.Product, n int64, err error) {
	line, err := br.ReadString('\n')
	n += int64(len(line))
	if err != nil {
		return nil, n, err
	}

	p, err = parseProduct(line)
	return p, n, err
}

func readProductList(br *bufio.Reader) (ps []*market.Product, n int64, err error) {
	for {
		line, err := br.ReadString('\n')
		n += int64(len(line))
		if err != nil {
			return nil, n, err
		}
		line = strings.TrimSuffix(line, "\n")
		if len(line) == 0 {
			break // reached the end of the message body
		}

		p, err := parseProduct(line)
		if err != nil {
			return nil, n, err
		}
		ps = append(ps, p)
	}
	return ps, n, nil
}

func parseProduct(line string) (*market.Product, error) {
	cr := csv.NewReader(strings.NewReader(line))
	record, err := cr.Read()
	// err == io.EOF should never be met
	if err != nil {
		return nil, err
	}
	if len(record) != 3 {
		return nil, errors.New("expected exactly 3 items per row")
	}

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, err
	}
	price, err := strconv.Atoi(record[2])
	if err != nil {
		return nil, err
	}
	return &market.Product{ID: id, Name: record[1], Price: price}, nil
}
