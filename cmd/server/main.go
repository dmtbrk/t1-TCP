package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ortymid/t1-tcp/market"
	"github.com/ortymid/t1-tcp/market/mem"
	"github.com/ortymid/t1-tcp/mmtp"
)

var mrkt *market.Market

func init() {
	productService := mem.NewProductService()
	mrkt = market.New(productService)
}

func handler(mw *mmtp.MessageWriter, msg *mmtp.Message) {
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
}

func main() {
	srv := &mmtp.Server{
		Addr:        ":8080",
		IdleTimeout: 5 * time.Second,
		Handler:     mmtp.HandlerFunc(handler),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()
	log.Println("Server started at 0.0.0.0:8080")

	<-done
	srv.Shutdown()
	log.Println("Server stopped.")
}

func writeError(mw *mmtp.MessageWriter, err error) {
	log.Println("ERROR:", err)
	res := &mmtp.Message{
		Type:    mmtp.MessageError,
		Payload: err,
	}
	mw.WriteMessage(res)
}
