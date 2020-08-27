package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ortymid/t1-tcp/market"
	"github.com/ortymid/t1-tcp/mtp"
)

const (
	MessageInvalid            int = 0
	MessageError              int = 1
	MessageProductRequest     int = 100
	MessageProductListRequest int = 101
	MessageProduct            int = 102
	MessageProductList        int = 103
	MessageProductAdd         int = 104
)

func main() {
	c, err := mtp.Dial(":8080")
	if err != nil {
		panic(err)
	}
	for {
		fmt.Println("===========================================")
		fmt.Println("Enter a number representing a message type:")
		fmt.Println("101 - Request Product List")
		fmt.Println("102 - Add Product")

		fmt.Print("Message type: ")
		typ := readInt()

		msg := &mtp.Message{Type: typ}

		switch msg.Type {
		case MessageProductListRequest:
			msg.Payload = ""
		case MessageProductAdd:
			fmt.Print("\nEnter product name: ")
			name := readString()
			fmt.Print("Enter product price: ")
			price := readInt()
			product := &market.Product{Name: name, Price: price}
			pld, err := json.Marshal(product)
			if err != nil {
				panic(err)
			}
			msg.Payload = string(pld)
		default:
			fmt.Println("Unknown message type. Exiting.")
			os.Exit(0)
		}

		err = c.SendMessage(msg)
		if err != nil {
			fmt.Println("ERROR sending message:", err)
			continue
		}

		res, err := c.ReceiveMessage()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed. Exiting.")
				os.Exit(0)
			}
			fmt.Println("ERROR receiving message:", err)
			continue
		}
		fmt.Println("\nServer response:")
		fmt.Println(res)
	}
}

func readInt() int {
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return i
}

func readString() string {
	var s string
	_, err := fmt.Scanf("%s", &s)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return s
}
