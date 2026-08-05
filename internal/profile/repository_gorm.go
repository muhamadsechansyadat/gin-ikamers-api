package profile

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByUserID(ctx context.Context, userID int64) (*Profile, error) {
	var p Profile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&p).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *GormRepository) Upsert(ctx context.Context, userID int64, fullName string, gender *string, birthDate *time.Time, bio *string) (*Profile, error) {
	var p Profile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&p).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		p = Profile{
			UserID:    userID,
			FullName:  fullName,
			Gender:    gender,
			BirthDate: birthDate,
			Bio:       bio,
		}
		if err := r.db.WithContext(ctx).Create(&p).Error; err != nil {
			return nil, err
		}
		return &p, nil
	}
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"full_name":  fullName,
		"updated_at": time.Now(),
	}
	if gender != nil {
		updates["gender"] = *gender
	}
	if birthDate != nil {
		updates["birth_date"] = *birthDate
	}
	if bio != nil {
		updates["bio"] = *bio
	}
	if err := r.db.WithContext(ctx).Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindByUserID(ctx, userID)
}

func (r *GormRepository) UpdateAvatar(ctx context.Context, userID int64, avatarPath *string) error {
	return r.db.WithContext(ctx).
		Model(&Profile{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"avatar_url": avatarPath,
			"updated_at": time.Now(),
		}).Error
}
