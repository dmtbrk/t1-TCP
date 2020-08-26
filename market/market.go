package market

type Market struct {
	productService ProductService
}

func New(productService ProductService) *Market {
	return &Market{productService: productService}
}

func (m *Market) Products() ([]*Product, error) {
	return m.productService.Products()
}

func (m *Market) AddProduct(p *Product) (*Product, error) {
	return m.productService.AddProduct(p)
}
