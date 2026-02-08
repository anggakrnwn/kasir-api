package services

import (
	"github.com/anggakrnwn/kasir-api/models"
	"github.com/anggakrnwn/kasir-api/repositories"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(items []models.CheckoutItem) (*models.Transaction, error) {
	return s.repo.CreateTransaction(items)
}

func (s *TransactionService) GetTodaySalesSummary() (*models.SalesSummary, error) {
	return s.repo.GetTodaySalesSummary()
}

func (s *TransactionService) GetSalesReport(startDate, endDate string) (*models.SalesSummary, error) {
	return s.repo.GetSalesReport(startDate, endDate)
}
