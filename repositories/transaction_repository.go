package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/anggakrnwn/kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productName string
		var productID, price, stock int

		err := tx.QueryRow("SELECT id, name, price, stock FROM products WHERE id=$1 FOR UPDATE", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		if stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product '%s'. Available: %d, Requested: %d",
				productName, stock, item.Quantity)
		}

		subtotal := item.Quantity * price
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount, created_at) VALUES ($1, NOW()) RETURNING id",
		totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// PERBAIKAN: Insert transaction details dengan loop yang benar
	for i := range details {
		details[i].TransactionID = transactionID

		var detailID int
		err = tx.QueryRow(
			"INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4) RETURNING id",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal,
		).Scan(&detailID)
		if err != nil {
			return nil, err
		}
		details[i].ID = detailID
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Get the complete transaction with timestamp
	var createdAt time.Time
	err = repo.db.QueryRow("SELECT created_at FROM transactions WHERE id = $1", transactionID).Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		CreatedAt:   createdAt,
		Details:     details,
	}, nil
}

// New method for sales summary
func (repo *TransactionRepository) GetTodaySalesSummary() (*models.SalesSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(t.total_amount), 0) as total_revenue,
			COALESCE(COUNT(t.id), 0) as total_transaksi
		FROM transactions t
		WHERE DATE(t.created_at) = CURRENT_DATE
	`

	var summary models.SalesSummary
	err := repo.db.QueryRow(query).Scan(&summary.TotalRevenue, &summary.TotalTransactions)
	if err != nil {
		return nil, err
	}

	// Get best selling product today
	bestSellingQuery := `
		SELECT 
			p.name as product_name,
			SUM(td.quantity) as total_quantity
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE DATE(t.created_at) = CURRENT_DATE
		GROUP BY p.name
		ORDER BY total_quantity DESC
		LIMIT 1
	`

	var bestProductName sql.NullString
	var bestProductQty sql.NullInt64
	err = repo.db.QueryRow(bestSellingQuery).Scan(&bestProductName, &bestProductQty)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if bestProductName.Valid && bestProductQty.Valid {
		summary.BestSellingProduct = &models.BestSellingProduct{
			Name:     bestProductName.String,
			Quantity: int(bestProductQty.Int64),
		}
	}

	return &summary, nil
}

// Method for date range report
func (repo *TransactionRepository) GetSalesReport(startDate, endDate string) (*models.SalesSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(t.total_amount), 0) as total_revenue,
			COALESCE(COUNT(t.id), 0) as total_transaksi
		FROM transactions t
		WHERE DATE(t.created_at) BETWEEN $1 AND $2
	`

	var summary models.SalesSummary
	err := repo.db.QueryRow(query, startDate, endDate).Scan(&summary.TotalRevenue, &summary.TotalTransactions)
	if err != nil {
		return nil, err
	}

	// Get best selling product for date range
	bestSellingQuery := `
		SELECT 
			p.name as product_name,
			SUM(td.quantity) as total_quantity
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE DATE(t.created_at) BETWEEN $1 AND $2
		GROUP BY p.name
		ORDER BY total_quantity DESC
		LIMIT 1
	`

	var bestProductName sql.NullString
	var bestProductQty sql.NullInt64
	err = repo.db.QueryRow(bestSellingQuery, startDate, endDate).Scan(&bestProductName, &bestProductQty)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if bestProductName.Valid && bestProductQty.Valid {
		summary.BestSellingProduct = &models.BestSellingProduct{
			Name:     bestProductName.String,
			Quantity: int(bestProductQty.Int64),
		}
	}

	return &summary, nil
}
