package auth

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterRequest struct {
	Email                string `json:"email" binding:"required,email"`
	Password             string `json:"password" binding:"required,min=8,max=72,strong_password"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
}

type GoogleLoginRequest struct {
	IdToken string `json:"id_token" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
