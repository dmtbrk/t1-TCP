package market

import (
	"encoding/json"
	"fmt"
)

type ProductService interface {
	Products() ([]*Product, error)
	AddProduct(*Product) (*Product, error)
}

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func (p *Product) String() string {
	return fmt.Sprintf("Product{ ID: %d, Name: %s, Price: %d }", p.ID, p.Name, p.Price)
}

// ParseProduct parses json encoded Product.
func ParseProduct(s string) (*Product, error) {
	p := &Product{}
	err := json.Unmarshal([]byte(s), p)
	return p, err
}
