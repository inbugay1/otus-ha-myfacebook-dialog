package repository

import "context"

type User struct {
	ID string
}

type UserRepository interface {
	GetUserByToken(ctx context.Context, token string) (*User, error)
}
