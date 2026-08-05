package profile

import (
	"gin-ikamers-api/internal/shared/storage"
	"strings"
	"time"
)

type UpdateProfileRequest struct {
	FullName  string  `json:"full_name" binding:"required,min=2,max=255"`
	Gender    *string `json:"gender" binding:"omitempty,oneof=male female other"`
	BirthDate *string `json:"birth_date" binding:"omitempty,datetime=2006-01-02"`
	Bio       *string `json:"bio" binding:"omitempty,max=1000"`
}

type ChangePasswordRequest struct {
	CurrentPassword    *string `json:"current_password" binding:"omitempty"`
	NewPassword        string  `json:"new_password" binding:"required,min=8,strong_password"`
	ConfirmNewPassword string  `json:"confirm_new_password" binding:"required,eqfield=NewPassword"`
}

type RequestEmailChangeRequest struct {
	NewEmail        string  `json:"new_email" binding:"required,email"`
	CurrentPassword *string `json:"current_password" binding:"omitempty"`
}

type ConfirmEmailChangeRequest struct {
	OTP string `json:"otp" binding:"required,len=6,numeric"`
}

type UserSummaryResponse struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ProfileResponse struct {
	UUID      string     `json:"uuid"`
	FullName  string     `json:"full_name"`
	Gender    *string    `json:"gender"`
	BirthDate *time.Time `json:"birth_date"`
	AvatarURL *string    `json:"avatar_url"`
	Bio       *string    `json:"bio"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	User UserSummaryResponse `json:"user"`
}

func ToResponse(p *Profile, userSummary UserSummaryResponse, s *storage.SupabaseClient) ProfileResponse {
	return ProfileResponse{
		UUID:      p.UUID,
		FullName:  p.FullName,
		Gender:    p.Gender,
		BirthDate: p.BirthDate,
		AvatarURL: buildAvatarURL(p.AvatarURL, s),
		Bio:       p.Bio,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		User:      userSummary,
	}
}

func buildAvatarURL(path *string, s *storage.SupabaseClient) *string {
	if path == nil || *path == "" {
		return nil
	}
	if strings.HasPrefix(*path, "http://") || strings.HasPrefix(*path, "https://") {
		return path
	}
	url := s.PublicURL(*path)
	return &url
}
