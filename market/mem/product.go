package mem

import (
	"sync"

	"github.com/ortymid/t1-tcp/market"
)

type ProductService struct {
	mu       sync.RWMutex
	lastID   int
	products []*market.Product
}

func NewProductService() *ProductService {
	products := []*market.Product{
		{ID: 1, Name: "Banana", Price: 1500},
		{ID: 2, Name: "Carrot", Price: 1400},
	}
	return &ProductService{products: products, lastID: 2}
}

func (srv *ProductService) Products() ([]*market.Product, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	return srv.products, nil
}

func (srv *ProductService) AddProduct(p *market.Product) (*market.Product, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	srv.lastID++
	p.ID = srv.lastID
	srv.products = append(srv.products, p)
	return p, nil
}
