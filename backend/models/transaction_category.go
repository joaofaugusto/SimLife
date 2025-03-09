package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// TransactionCategory models/transaction_category.go
type TransactionCategory struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name string    `gorm:"unique;not null"`
	Type string    `gorm:"not null;check:type IN ('credit', 'debit', 'transfer')"`
}

// Transaction models/transaction.go
type Transaction struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID           `gorm:"type:uuid;not null"`
	User            User                `gorm:"foreignKey:UserID"` // Add this
	CategoryID      uuid.UUID           `gorm:"type:uuid;not null"`
	Category        TransactionCategory `gorm:"foreignKey:CategoryID"` // Add this
	FromAccountID   *uuid.UUID          `gorm:"type:uuid;default:null"`
	FromAccount     BankAccount         `gorm:"foreignKey:FromAccountID"` // Add this
	ToAccountID     *uuid.UUID          `gorm:"type:uuid;default:null"`
	ToAccount       BankAccount         `gorm:"foreignKey:ToAccountID"` // Add this
	Amount          float64             `gorm:"type:numeric(15,2);not null"`
	Description     string
	TransactionDate time.Time `gorm:"autoCreateTime" json:"transaction_date"`
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(
		&struct {
			TransactionDate string `json:"transaction_date"`
			*Alias
		}{
			TransactionDate: t.TransactionDate.Format("2006-01-02T15:04:05.000Z07:00"),
			Alias:           (*Alias)(&t),
		},
	)
}
