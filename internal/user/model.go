package user

type User struct {
	ID              int64
	UUID            string
	RoleID          int64
	RoleName        string
	Email           string
	PasswordHash    *string
	IsActive        bool
	IsEmailVerified bool
}
