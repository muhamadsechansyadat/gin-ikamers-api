package profile

import (
	"context"
	"errors"
	"fmt"
	"gin-ikamers-api/internal/shared/storage"
	"io"
	"time"
)

var ErrProfileNotFound = errors.New("profile not found")

type Service struct {
	repo    Repository
	storage *storage.SupabaseClient
}

func NewService(repo Repository, s *storage.SupabaseClient) *Service {
	return &Service{repo: repo, storage: s}
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*Profile, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) Upsert(ctx context.Context, userID int64, fullName string, gender *string, birthDate *time.Time, bio *string) (*Profile, error) {
	return s.repo.Upsert(ctx, userID, fullName, gender, birthDate, bio)
}

func (s *Service) UploadAvatar(ctx context.Context, userID int64, r io.Reader, contentType, ext string) (*Profile, error) {
	p, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProfileNotFound
	}

	if p.AvatarURL != nil && *p.AvatarURL != "" {
		_ = s.storage.Delete(ctx, *p.AvatarURL)
	}

	path := fmt.Sprintf("profile/%s-%d%s", p.UUID, time.Now().Unix(), ext)

	if err := s.storage.Upload(ctx, path, contentType, r); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAvatar(ctx, userID, &path); err != nil {
		_ = s.storage.Delete(ctx, path)
		return nil, err
	}

	p.AvatarURL = &path
	return p, nil
}

func (s *Service) DeleteAvatar(ctx context.Context, userID int64) (*Profile, error) {
	p, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProfileNotFound
	}

	if p.AvatarURL != nil && *p.AvatarURL != "" {
		_ = s.storage.Delete(ctx, *p.AvatarURL)
	}

	if err := s.repo.UpdateAvatar(ctx, userID, nil); err != nil {
		return nil, err
	}

	p.AvatarURL = nil
	return p, nil
}
