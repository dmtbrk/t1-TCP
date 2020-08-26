package main

import (
	"fmt"
	"log"

	"github.com/ortymid/t1-tcp/market"
	"github.com/ortymid/t1-tcp/market/mem"
	"github.com/ortymid/t1-tcp/mmtp"
)

func main() {
	productService := mem.NewProductService()
	mrkt := market.New(productService)

	fmt.Println("Starting server at localhost:8080...")

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
}

func writeError(mw *mmtp.MessageWriter, err error) {
	log.Println("ERROR:", err)
	res := &mmtp.Message{
		Type:    mmtp.MessageError,
		Payload: err,
	}
	mw.WriteMessage(res)
}
