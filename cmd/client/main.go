package main

import (
	"encoding/json"
	"fmt"
	"io"

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
ConnectionLoop:
	for {
		c, err := mtp.Dial(":8080")
		if err != nil {
			panic(err)
		}
	MessageInputLoop:
		for {
			fmt.Println("===========================================")
			fmt.Println("Enter a number representing a message type:")
			fmt.Println("101 - Request Product List")
			fmt.Println("104 - Add Product")

			msg := &mtp.Message{}
			for {
				fmt.Print("Message type: ")
				msg.Type, err = readInt()
				if err != nil {
					fmt.Println(err)
					continue
				}

				switch msg.Type {
				case MessageProductListRequest:
					msg.Payload = ""
				case MessageProductAdd:
					fmt.Print("\nEnter product name: ")
					name, err := readString()
					if err != nil {
						fmt.Println(err)
						continue
					}

					fmt.Print("Enter product price: ")
					price, err := readInt()
					if err != nil {
						fmt.Println(err)
						continue
					}

					product := &market.Product{Name: name, Price: price}
					pld, err := json.Marshal(product)
					if err != nil {
						panic(err)
					}
					msg.Payload = string(pld)
				default:
					fmt.Println("Unknown message type. Try again.")
					continue
				}
				break
			}

			err = c.SendMessage(msg)
			if err != nil {
				fmt.Println("ERROR sending message:", err)
				continue MessageInputLoop
			}

			res, err := c.ReceiveMessage()
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed. Reconnecting.")
					continue ConnectionLoop
				}
				fmt.Println("ERROR receiving message:", err)
				continue MessageInputLoop
			}

			fmt.Println("\nServer response:")
			fmt.Println("Type:", res.Type)
			fmt.Println("Payload:", res.Payload)
		}
	}
}

func readInt() (int, error) {
	var i int
	_, err := fmt.Scanf("%d", &i)
	return i, err
}

func readString() (string, error) {
	var s string
	_, err := fmt.Scanf("%s", &s)
	return s, err
}
