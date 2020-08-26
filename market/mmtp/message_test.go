package mmtp

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/ortymid/t1-tcp/market"
)

func TestMessage_WriteTo(t *testing.T) {
	type fields struct {
		Type    MessageType
		Payload interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Write ProductRequest message",
			fields: fields{
				Type:    MessageProductRequest,
				Payload: 1,
			},
			want:    "1\n1\n",
			wantErr: false,
		},
		{
			name: "Write Product message",
			fields: fields{
				Type:    MessageProduct,
				Payload: &market.Product{ID: 1, Name: "p", Price: 100},
			},
			want:    "2\n1,\"p\",100\n",
			wantErr: false,
		},
		{
			name: "Write ProductListRequest message",
			fields: fields{
				Type:    MessageProductListRequest,
				Payload: nil,
			},
			want:    "3\n",
			wantErr: false,
		},
		{
			name: "Write ProductList message",
			fields: fields{
				Type: MessageProductList,
				Payload: []*market.Product{
					{ID: 1, Name: "p1", Price: 100},
					{ID: 2, Name: "p2", Price: 200},
				},
			},
			want:    "4\n1,\"p1\",100\n2,\"p2\",200\n\n",
			wantErr: false,
		},
		{
			name: "Write ProductAdd message",
			fields: fields{
				Type:    MessageProductAdd,
				Payload: &market.Product{ID: 1, Name: "p", Price: 100},
			},
			want:    "5\n1,\"p\",100\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				Type:    tt.fields.Type,
				Payload: tt.fields.Payload,
			}
			w := &bytes.Buffer{}
			_, err := msg.WriteTo(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := w.String(); got != tt.want {
				t.Errorf("Message.WriteTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_ReadFrom(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantMsg *Message
		wantErr bool
	}{
		{
			name: "Read ProductRequest message",
			in:   "1\n1\n",
			wantMsg: &Message{
				Type:    MessageProductRequest,
				Payload: 1,
			},
			wantErr: false,
		},
		{
			name: "Read Product message",
			in:   "2\n1,\"p\",100\n",
			wantMsg: &Message{
				Type:    MessageProduct,
				Payload: &market.Product{ID: 1, Name: "p", Price: 100},
			},
			wantErr: false,
		},
		{
			name: "Read ProductListRequest message",
			in:   "3\n",
			wantMsg: &Message{
				Type:    MessageProductListRequest,
				Payload: nil,
			},
			wantErr: false,
		},
		{
			name: "Read ProductList message",
			in:   "4\n1,\"p1\",100\n2,\"p2\",200\n\n",
			wantMsg: &Message{
				Type: MessageProductList,
				Payload: []*market.Product{
					{ID: 1, Name: "p1", Price: 100},
					{ID: 2, Name: "p2", Price: 200},
				},
			},
			wantErr: false,
		},
		{
			name: "Read ProductAdd message",
			in:   "5\n1,\"p\",100\n",
			wantMsg: &Message{
				Type:    MessageProductAdd,
				Payload: &market.Product{ID: 1, Name: "p", Price: 100},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{}
			r := strings.NewReader(tt.in)
			_, err := msg.ReadFrom(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(msg, tt.wantMsg) {
				t.Errorf("Message = %v, want %v", msg, tt.wantMsg)
			}
		})
	}
}
