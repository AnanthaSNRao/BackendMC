package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpriedToken = errors.New("token has expried")
	ErrInvalidToken = errors.New("Invalid token")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpriedAt time.Time `json:"expried_at"`
}

func NewPayload(userName string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  userName,
		IssuedAt:  time.Now(),
		ExpriedAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpriedAt) {
		return ErrExpriedToken
	}
	return nil
}
