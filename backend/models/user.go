package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"column:hashed_pwd;not null" json:"password_hash"` // Map JSON field
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// BeforeCreate Hook: Generates a UUID and hashes the password before saving to the database
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID
	fmt.Printf("Raw password during registration: %s\n", u.PasswordHash) // Debug log
	u.ID = uuid.New()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)

	return nil
}
