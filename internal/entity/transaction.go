package entity

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents the transaction entity in the database.
type Transaction struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type       string     `gorm:"type:varchar(20);not null"`
	Amount     float64    `gorm:"type:decimal(10,2);not null"`
	Status     string     `gorm:"type:varchar(20);not null"`
	ConsumerID string     `gorm:"type:uuid;not null"`
	CreatedAt  *time.Time `gorm:"type:timestamptz"`
	UpdatedAt  *time.Time `gorm:"type:timestamptz"`
}
