package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ortymid/t1-tcp/market"
	"github.com/ortymid/t1-tcp/market/mem"
	"github.com/ortymid/t1-tcp/market/mmtp"
)

func main() {
	productService := mem.NewProductService()
	mrkt := market.New(productService)

	// server
	go func() {
		err := mmtp.ListenAndServe(":8080", mmtp.HandlerFunc(func(mw *mmtp.MessageWriter, msg *mmtp.Message) {
			log.Println("INFO: server received:", msg)

			switch msg.Type {
			case mmtp.MessageProductAdd:
				// doing business
				product, err := mrkt.AddProduct(msg.Payload.(*market.Product))

				// writing the response
				if err != nil {
					writeError(mw, err)
					break
				}
				res := &mmtp.Message{
					Type:    mmtp.MessageProduct,
					Payload: product,
				}
				mw.WriteMessage(res)

			case mmtp.MessageProductListRequest:
				// doing business
				products, err := mrkt.Products()

				// writing the response
				if err != nil {
					writeError(mw, err)
					break
				}
				res := &mmtp.Message{
					Type:    mmtp.MessageProductList,
					Payload: products,
				}
				mw.WriteMessage(res)
			}
		}))
		if err != nil {
			panic(err)
		}
	}()

	// client:
	go func() {
		time.Sleep(500 * time.Millisecond) // wait for the server

		c, err := mmtp.Dial(":8080")
		if err != nil {
			panic(err)
		}

		// requesting all products
		req := &mmtp.Message{
			Type: mmtp.MessageProductListRequest,
		}
		log.Println("CLIENT: sending:\n", req)
		c.SendMessage(req)
		res, err := c.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		log.Println("CLIENT: received:\n", res)

		// adding a new product
		p := &market.Product{
			Name:  "New Product",
			Price: 1000,
		}
		req = &mmtp.Message{
			Type:    mmtp.MessageProductAdd,
			Payload: p,
		}
		log.Println("CLIENT: sending:\n", req)
		c.SendMessage(req)
		res, err = c.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		log.Println("CLIENT: received:\n", res)

		// requesting all products
		req = &mmtp.Message{
			Type: mmtp.MessageProductListRequest,
		}
		log.Println("CLIENT: sending:\n", req)
		c.SendMessage(req)
		res, err = c.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		log.Println("CLIENT: received:\n", res)
	}()

	fmt.Println("Press any key to stop.")
	stdr := bufio.NewReader(os.Stdin)
	stdr.ReadString('\n')
}

func writeError(mw *mmtp.MessageWriter, err error) {
	log.Println("ERROR:", err)
	res := &mmtp.Message{
		Type:    mmtp.MessageError,
		Payload: err,
	}
	mw.WriteMessage(res)
}
