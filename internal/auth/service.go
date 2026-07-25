package auth

import (
	"context"
	"gin-ikamers-api/internal/user"
	"google.golang.org/api/idtoken"
	"time"
)

type Service struct {
	repo           user.Repository
	tokens         *TokenService
	googleClientID string
	defaultRoleID  int64
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewService(repo user.Repository, tokens *TokenService, googleClientID string, defaultRoleID int64) *Service {
	return &Service{
		repo:           repo,
		tokens:         tokens,
		googleClientID: googleClientID,
		defaultRoleID:  defaultRoleID,
	}
}

func (s *Service) LoginWithPassword(ctx context.Context, email, password, ua, ip string) (*TokenPair, error) {
	u, err := s.repo.FindByEmail(ctx, email)
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
	return s.repo.CreateWithIdentity(ctx,
		&user.User{RoleID: roleID, Email: email, PasswordHash: hashPtr, IsActive: true},
		&user.Identity{Provider: "local", ProviderUserID: email, Email: &email},
	)
}

func (s *Service) LoginWithGoogle(ctx context.Context, idTokenStr, ua, ip string) (*TokenPair, error) {
	payload, err := idtoken.Validate(ctx, idTokenStr, s.googleClientID)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	googleSubject := payload.Subject
	googleEmail, _ := payload.Claims["email"].(string)

	u, err := s.repo.FindByProviderIdentity(ctx, "google", googleSubject)
	if err == nil && u != nil {
		if !u.IsActive {
			return nil, ErrUserInactive
		}
		return s.issueTokens(ctx, u, ua, ip)
	}

	existingUser, err := s.repo.FindByEmail(ctx, googleEmail)
	if err == nil && existingUser != nil {
		if err := s.repo.LinkIdentity(ctx, &user.Identity{
			UserID:         existingUser.ID,
			Provider:       "google",
			ProviderUserID: googleSubject,
			Email:          &googleEmail,
		}); err != nil {
			return nil, err
		}
		if !existingUser.IsActive {
			return nil, ErrUserInactive
		}
		return s.issueTokens(ctx, existingUser, ua, ip)
	}

	newUser, err := s.repo.CreateWithIdentity(ctx,
		&user.User{
			RoleID:   s.defaultRoleID,
			Email:    googleEmail,
			IsActive: true,
		},
		&user.Identity{
			Provider:       "google",
			ProviderUserID: googleSubject,
			Email:          &googleEmail,
		},
	)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, newUser, ua, ip)
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	tokenHash := HashToken(refreshTokenStr)
	userID, err := s.repo.FindActiveRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	u, err := s.repo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return nil, ErrUserNotFound
	}

	if !u.IsActive {
		return nil, ErrUserInactive
	}

	_ = s.repo.RevokeRefreshToken(ctx, tokenHash)

	access, err := s.tokens.GenerateAccessToken(u.ID, u.RoleName)
	if err != nil {
		return nil, err
	}

	raw, hash, exp, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveRefreshToken(ctx, u.ID, hash, exp, "", ""); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresAt:    exp,
	}, nil
}

func (s *Service) Logout(ctx context.Context, userID int64) error {
	return s.repo.RevokeAllRefreshTokens(ctx, userID)
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
	if err := s.repo.SaveRefreshToken(ctx, u.ID, hash, exp, ua, ip); err != nil {
		return nil, err
	}
	_ = s.repo.UpdateLastLogin(ctx, u.ID, time.Now())
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresAt:    exp,
	}, nil
}
