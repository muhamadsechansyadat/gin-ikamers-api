package auth

import "time"

type Identity struct {
	ID             int64
	UserID         int64
	Provider       string
	ProviderUserID string
	Email          *string
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	UserAgent *string
	IPAddress *string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
