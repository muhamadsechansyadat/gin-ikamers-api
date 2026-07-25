package user

import (
	"context"
	"time"
)

type User struct {
	ID              int64
	UUID            string
	RoleID          int64
	RoleName        string // hasil join dari roles
	Email           string
	PasswordHash    *string // null untuk oauth user
	IsActive        bool
	IsEmailVerified bool
}

type Identity struct {
	ID             int64
	UserID         int64
	Provider       string
	ProviderUserID string
	Email          *string
}

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByProviderIdentity(ctx context.Context, provider, providerUserID string) (*User, error)
	CreateWithIdentity(ctx context.Context, u *User, id *Identity) (*User, error)
	LinkIdentity(ctx context.Context, id *Identity) error
	UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error

	SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ua, ip string) error
	FindActiveRefreshToken(ctx context.Context, tokenHash string) (userID int64, err error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID int64) error
}
