package profile

import (
	"context"
	"time"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID int64) (*Profile, error)
	Upsert(ctx context.Context, userID int64, fullName string, gender *string, birthDate *time.Time, bio *string) (*Profile, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarPath *string) error
}
