package mtp

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestReadMessage(t *testing.T) {
	type args struct {
		r *bufio.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Message
		wantErr bool
	}{
		{
			name:    "Read valid message without payload",
			args:    args{r: bufio.NewReader(strings.NewReader("1\r\n\r\n"))},
			want:    &Message{Type: 1, Payload: ""},
			wantErr: false,
		},
		{
			name:    "Read valid message with payload",
			args:    args{r: bufio.NewReader(strings.NewReader("2\r\npayload\r\n"))},
			want:    &Message{Type: 2, Payload: "payload"},
			wantErr: false,
		},
		{
			name:    "Stop reading message after last CRLF",
			args:    args{r: bufio.NewReader(strings.NewReader("2\r\npayload\r\njunk"))},
			want:    &Message{Type: 2, Payload: "payload"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadMessage(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadMessage() Payload = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_WriteTo(t *testing.T) {
	type fields struct {
		Type    int
		Payload string
	}
	tests := []struct {
		name    string
		fields  fields
		wantN   int64
		wantW   string
		wantErr bool
	}{
		{
			name:    "Write message without payload",
			fields:  fields{Type: 1, Payload: ""},
			wantN:   5,
			wantW:   "1\r\n\r\n",
			wantErr: false,
		},
		{
			name:    "Write message with payload",
			fields:  fields{Type: 1, Payload: "payload"},
			wantN:   12,
			wantW:   "1\r\npayload\r\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Type:    tt.fields.Type,
				Payload: tt.fields.Payload,
			}
			w := &bytes.Buffer{}
			gotN, err := m.WriteTo(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("Message.WriteTo() = %v, want %v", gotN, tt.wantN)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Message.WriteTo() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
