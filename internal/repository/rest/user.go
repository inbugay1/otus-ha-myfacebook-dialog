package rest

import (
	"context"
	"errors"
	"fmt"

	"myfacebook-dialog/internal/myfacebookapiclient"
	"myfacebook-dialog/internal/repository"
)

type UserRepository struct {
	apiClient *myfacebookapiclient.Client
}

func NewUserRepository(apiClient *myfacebookapiclient.Client) *UserRepository {
	return &UserRepository{
		apiClient: apiClient,
	}
}

func (r *UserRepository) GetUserByToken(ctx context.Context, token string) (*repository.User, error) {
	user, err := r.apiClient.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, myfacebookapiclient.ErrNotFound) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("userrepository failed to get user by token: %w", err)
	}

	return &repository.User{
		ID: user.ID,
	}, nil
}
