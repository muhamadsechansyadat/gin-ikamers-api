package user

type UserResponse struct {
	ID              int64  `json:"id"`
	UUID            string `json:"uuid"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsActive        bool   `json:"is_active"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

func ToResponse(u *User) UserResponse {
	return UserResponse{
		ID:              u.ID,
		UUID:            u.UUID,
		Email:           u.Email,
		Role:            u.RoleName,
		IsActive:        u.IsActive,
		IsEmailVerified: u.IsEmailVerified,
	}
}
