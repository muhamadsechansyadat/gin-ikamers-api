package user

import (
	"context"
	"time"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, u *User) (*User, error)
	UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateEmail(ctx context.Context, userID int64, newEmail string) error
}
