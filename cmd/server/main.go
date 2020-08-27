package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ortymid/t1-tcp/market"
	"github.com/ortymid/t1-tcp/market/mem"
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

var mrkt *market.Market

func init() {
	productService := mem.NewProductService()
	mrkt = market.New(productService)
}

func handler(mw *mtp.MessageWriter, msg *mtp.Message) {
	log.Println("INFO: server received:", msg)

	switch msg.Type {
	case MessageProductAdd:
		// processing request
		product, err := market.ParseProduct(msg.Payload)
		if err != nil {
			writeError(mw, err)
			break
		}
		// doing logic
		product, err = mrkt.AddProduct(product)
		if err != nil {
			writeError(mw, err)
			break
		}
		// preparing response
		pld, err := json.Marshal(product)
		if err != nil {
			writeError(mw, err)
			break
		}
		res := &mtp.Message{
			Type:    MessageProduct,
			Payload: string(pld),
		}
		mw.WriteMessage(res)

	case MessageProductListRequest:
		// doing logic
		products, err := mrkt.Products()
		if err != nil {
			writeError(mw, err)
			break
		}

		// preparing response
		pld, err := json.Marshal(products)
		if err != nil {
			writeError(mw, err)
			break
		}
		res := &mtp.Message{
			Type:    MessageProduct,
			Payload: string(pld),
		}
		mw.WriteMessage(res)
	}
}

func main() {
	srv := &mtp.Server{
		Addr:        ":8080",
		IdleTimeout: 30 * time.Second,
		Handler:     mtp.HandlerFunc(handler),
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

func writeError(mw *mtp.MessageWriter, err error) {
	log.Println("ERROR:", err)
	res := &mtp.Message{
		Type:    MessageError,
		Payload: err.Error(),
	}
	mw.WriteMessage(res)
}
