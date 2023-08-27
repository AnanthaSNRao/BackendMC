// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package db

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        int64
	Owner     string
	Balance   int64
	Currency  string
	CreatedAt time.Time
}

type Entry struct {
	ID        int64
	AccountID int64
	// can be negative or positive
	Amount    int64
	CreatedAt time.Time
}

type Session struct {
	ID           uuid.UUID
	Username     string
	RefreshToken string
	UserAgent    string
	ClientIp     string
	IsBlocked    bool
	CreatedAt    time.Time
	ExpiredAt    time.Time
}

type Transfer struct {
	ID            int64
	FromAccountID int64
	ToAccountID   int64
	// must be positive
	Amount    int64
	CreatedAt time.Time
}

type User struct {
	Username          string
	HashedPassword    string
	FullName          string
	Email             string
	PasswordChangedAt time.Time
	CreatedAt         time.Time
}
