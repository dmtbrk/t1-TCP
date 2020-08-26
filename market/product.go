package market

import "fmt"

type ProductService interface {
	Products() ([]*Product, error)
	AddProduct(*Product) (*Product, error)
}

type Product struct {
	ID    int
	Name  string
	Price int
}

func (p *Product) String() string {
	return fmt.Sprintf("Product{ ID: %d, Name: %s, Price: %d }", p.ID, p.Name, p.Price)
}
