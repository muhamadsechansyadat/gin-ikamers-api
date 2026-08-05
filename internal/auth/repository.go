package auth

import (
	"context"
	"time"
)

type Repository interface {
	LinkIdentity(ctx context.Context, id *Identity) error
	FindUserIDByProviderIdentity(ctx context.Context, provider, providerUserID string) (int64, error)

	SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ua, ip string) error
	FindActiveRefreshToken(ctx context.Context, tokenHash string) (int64, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID int64) error

	CreateEmailChangeVerification(ctx context.Context, userID int64, newEmail, otpHash string, expiresAt time.Time) error
	FindActiveEmailChangeVerification(ctx context.Context, userID int64) (*EmailChangeVerification, error)
	MarkEmailChangeVerificationUsed(ctx context.Context, id int64) error
	InvalidateActiveEmailChangeVerifications(ctx context.Context, userID int64) error
}
