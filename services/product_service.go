package services

import (
	"errors"
	"strings"

	"github.com/anggakrnwn/kasir-api/models"
	"github.com/anggakrnwn/kasir-api/repositories"
)

var (
	ErrInvalidProductName  = errors.New("product name cannot be empty")
	ErrInvalidProductPrice = errors.New("product price must be greater than zero")
	ErrInvalidProductStock = errors.New("product stock cannot be negative")
)

type ProductService struct {
	repo *repositories.ProductRepository
}

func NewProductService(repo *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetAll(name string) ([]models.Product, error) {
	return s.repo.GetAll(name)

}

func (s *ProductService) Create(data *models.Product) error {

	if strings.TrimSpace(data.Name) == "" {
		return ErrInvalidProductName
	}
	if data.Price <= 0 {
		return ErrInvalidProductPrice
	}
	if data.Stock < 0 {
		return ErrInvalidProductStock
	}

	return s.repo.Create(data)
}

func (s *ProductService) GetByID(id int) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) Update(product *models.Product) error {
	return s.repo.Update(product)
}

func (s *ProductService) Delete(id int) error {
	return s.repo.Delete(id)
}
