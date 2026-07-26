package auth

import (
	"context"
	"gin-ikamers-api/internal/user"
	"google.golang.org/api/idtoken"
	"time"
)

type Service struct {
	users          user.Repository
	authRepo       Repository
	tokens         *TokenService
	googleClientID string
	defaultRoleID  int64
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewService(users user.Repository, authRepo Repository, tokens *TokenService, googleClientID string, defaultRoleID int64) *Service {
	return &Service{
		users:          users,
		authRepo:       authRepo,
		tokens:         tokens,
		googleClientID: googleClientID,
		defaultRoleID:  defaultRoleID,
	}
}

func (s *Service) LoginWithPassword(ctx context.Context, email, password, ua, ip string) (*TokenPair, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil || u == nil {
		return nil, ErrInvalidCredentials
	}
	if !u.IsActive {
		return nil, ErrUserInactive
	}
	if u.PasswordHash == nil || !VerifyPassword(*u.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, u, ua, ip)
}

func (s *Service) RegisterWithPassword(ctx context.Context, email, password string, roleID int64) (*user.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	hashPtr := &hash

	newUser, err := s.users.Create(ctx, &user.User{
		RoleID: roleID, Email: email, PasswordHash: hashPtr, IsActive: true,
	})
	if err != nil {
		return nil, err
	}

	if err := s.authRepo.LinkIdentity(ctx, &Identity{
		UserID: newUser.ID, Provider: "local", ProviderUserID: email, Email: &email,
	}); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *Service) LoginWithGoogle(ctx context.Context, idTokenStr, ua, ip string) (*TokenPair, error) {
	payload, err := idtoken.Validate(ctx, idTokenStr, s.googleClientID)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	googleSubject := payload.Subject
	googleEmail, _ := payload.Claims["email"].(string)

	// 1. Identity Google sudah pernah di-link?
	userID, err := s.authRepo.FindUserIDByProviderIdentity(ctx, "google", googleSubject)
	if err == nil && userID > 0 {
		u, err := s.users.FindByID(ctx, userID)
		if err != nil || u == nil {
			return nil, ErrUserNotFound
		}
		if !u.IsActive {
			return nil, ErrUserInactive
		}
		return s.issueTokens(ctx, u, ua, ip)
	}

	// 2. Email sudah ada (dari register local)? Link identity Google-nya.
	existingUser, err := s.users.FindByEmail(ctx, googleEmail)
	if err == nil && existingUser != nil {
		if err := s.authRepo.LinkIdentity(ctx, &Identity{
			UserID: existingUser.ID, Provider: "google",
			ProviderUserID: googleSubject, Email: &googleEmail,
		}); err != nil {
			return nil, err
		}
		if !existingUser.IsActive {
			return nil, ErrUserInactive
		}
		return s.issueTokens(ctx, existingUser, ua, ip)
	}

	// 3. User baru — create user + link identity
	newUser, err := s.users.Create(ctx, &user.User{
		RoleID: s.defaultRoleID, Email: googleEmail, IsActive: true,
	})
	if err != nil {
		return nil, err
	}
	if err := s.authRepo.LinkIdentity(ctx, &Identity{
		UserID: newUser.ID, Provider: "google",
		ProviderUserID: googleSubject, Email: &googleEmail,
	}); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, newUser, ua, ip)
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	tokenHash := HashToken(refreshTokenStr)
	userID, err := s.authRepo.FindActiveRefreshToken(ctx, tokenHash)
	if err != nil || userID == 0 {
		return nil, ErrInvalidCredentials
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return nil, ErrUserNotFound
	}
	if !u.IsActive {
		return nil, ErrUserInactive
	}

	_ = s.authRepo.RevokeRefreshToken(ctx, tokenHash)

	return s.issueTokens(ctx, u, "", "")
}

func (s *Service) Logout(ctx context.Context, userID int64) error {
	return s.authRepo.RevokeAllRefreshTokens(ctx, userID)
}

func (s *Service) VerifyAccessToken(tokenStr string) (*Claims, error) {
	return s.tokens.ParseAccessToken(tokenStr)
}

func (s *Service) issueTokens(ctx context.Context, u *user.User, ua, ip string) (*TokenPair, error) {
	access, err := s.tokens.GenerateAccessToken(u.ID, u.RoleName)
	if err != nil {
		return nil, err
	}
	raw, hash, exp, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.authRepo.SaveRefreshToken(ctx, u.ID, hash, exp, ua, ip); err != nil {
		return nil, err
	}
	_ = s.users.UpdateLastLogin(ctx, u.ID, time.Now())
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresAt:    exp,
	}, nil
}
