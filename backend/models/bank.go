package models

import (
	"github.com/google/uuid"
	"time"
)

// Bank models/bank.go
type Bank struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"unique;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// BankAccount models/bank_account.go
type BankAccount struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	BankID        uuid.UUID `gorm:"type:uuid;not null"`
	AgencyNumber  string    `gorm:"not null"`
	AccountNumber string    `gorm:"not null"`
	Balance       float64   `gorm:"type:numeric(15,2);default:0.00"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
