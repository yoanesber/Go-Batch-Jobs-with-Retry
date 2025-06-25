package repository

import (
	"gorm.io/gorm" // Import GORM for ORM functionalities

	"github.com/yoanesber/go-batch-jobs-with-retry/internal/entity"
)

type TransactionRepository interface {
	GetAllTransactions(tx *gorm.DB, page int, limit int) ([]entity.Transaction, error)
}

type transactionRepository struct{}

func NewTransactionRepository() TransactionRepository {
	return &transactionRepository{}
}

func (r *transactionRepository) GetAllTransactions(tx *gorm.DB, page int, limit int) ([]entity.Transaction, error) {
	var transaction []entity.Transaction
	err := tx.Order("created_at ASC").
		Offset(page).
		Limit(limit).
		Find(&transaction).Error

	if err != nil {
		return nil, err
	}

	return transaction, nil
}
