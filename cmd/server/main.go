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

var handlers map[int]mtp.Handler

var mrkt *market.Market

func init() {
	productService := mem.NewProductService()
	mrkt = market.New(productService)
}

func handleProductListRequest(mw *mtp.MessageWriter, msg *mtp.Message) {
	// doing logic
	products, err := mrkt.Products()
	if err != nil {
		writeError(mw, err)
		return
	}

	// preparing response
	pld, err := json.Marshal(products)
	if err != nil {
		writeError(mw, err)
		return
	}
	res := &mtp.Message{
		Type:    MessageProductList,
		Payload: string(pld),
	}
	mw.WriteMessage(res)
}

func handleProductAdd(mw *mtp.MessageWriter, msg *mtp.Message) {
	// processing request
	product, err := market.ParseProduct(msg.Payload)
	if err != nil {
		writeError(mw, err)
		return
	}
	// doing logic
	product, err = mrkt.AddProduct(product)
	if err != nil {
		writeError(mw, err)
		return
	}
	// preparing response
	pld, err := json.Marshal(product)
	if err != nil {
		writeError(mw, err)
		return
	}
	res := &mtp.Message{
		Type:    MessageProduct,
		Payload: string(pld),
	}
	mw.WriteMessage(res)
}

func writeError(mw *mtp.MessageWriter, err error) {
	log.Println("ERROR:", err)
	res := &mtp.Message{
		Type:    MessageError,
		Payload: err.Error(),
	}
	mw.WriteMessage(res)
}

func main() {
	srv := &mtp.Server{
		Addr:        ":8080",
		IdleTimeout: 30 * time.Second,
	}

	mtp.HandleFunc(MessageProductListRequest, handleProductListRequest)
	mtp.HandleFunc(MessageProductAdd, handleProductAdd)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()
	log.Println("Server started at 0.0.0.0:8080")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	srv.Shutdown()
	log.Println("Server stopped.")
}
