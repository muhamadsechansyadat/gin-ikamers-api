package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"gin-ikamers-api/internal/platform/mailer"
	"gin-ikamers-api/internal/platform/ratelimit"
	"gin-ikamers-api/internal/user"
	"google.golang.org/api/idtoken"
	"math/big"
	"strings"
	"time"
)

type Service struct {
	users          user.Repository
	authRepo       Repository
	tokens         *TokenService
	googleClientID string
	defaultRoleID  int64
	mailer         mailer.Mailer
	limiter        *ratelimit.Limiter
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewService(users user.Repository, authRepo Repository, tokens *TokenService,
	mailer mailer.Mailer, googleClientID string, defaultRoleID int64, limiter *ratelimit.Limiter) *Service {
	return &Service{
		users: users, authRepo: authRepo, tokens: tokens,
		mailer: mailer, googleClientID: googleClientID, defaultRoleID: defaultRoleID, limiter: limiter,
	}
}

const (
	otpLength     = 6
	otpTTLMinutes = 15
)

func (s *Service) LoginWithPassword(ctx context.Context, email, password, ua, ip string) (*TokenPair, error) {
	key := "login:email" + strings.ToLower(strings.TrimSpace(email))

	res, err := s.limiter.Allow(ctx, key, ratelimit.Per(5, 15*time.Minute))
	if err == nil && !res.Allowed {
		return nil, ErrTooManyLoginAttempts
	}

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

	_ = s.limiter.Reset(ctx, key)
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

func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword *string, newPassword string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return ErrUserNotFound
	}

	hasPassword := u.PasswordHash != nil && *u.PasswordHash != ""

	if hasPassword {
		if currentPassword == nil || *currentPassword == "" {
			return ErrCurrentPasswordRequired
		}
		if !VerifyPassword(*u.PasswordHash, *currentPassword) {
			return ErrInvalidCredentials
		}
		if VerifyPassword(*u.PasswordHash, newPassword) {
			return ErrPasswordSameAsCurrent
		}
	}

	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}
	_ = s.authRepo.RevokeAllRefreshTokens(ctx, userID)
	return nil
}

func (s *Service) RequestEmailChange(ctx context.Context, userID int64, newEmail string, currentPassword *string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return ErrUserNotFound
	}

	hasPassword := u.PasswordHash != nil && *u.PasswordHash != ""
	if hasPassword {
		if currentPassword == nil || *currentPassword == "" {
			return ErrCurrentPasswordRequired
		}
		if !VerifyPassword(*u.PasswordHash, *currentPassword) {
			return ErrInvalidCredentials
		}
	}

	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == strings.ToLower(u.Email) {
		return ErrEmailAlreadyUsed
	}
	existing, _ := s.users.FindByEmail(ctx, newEmail)
	if existing != nil {
		return ErrEmailAlreadyUsed
	}

	otp, err := generateNumericOTP(otpLength)
	if err != nil {
		return err
	}
	otpHash := hashOTP(otp)
	expiresAt := time.Now().Add(otpTTLMinutes * time.Minute)

	_ = s.authRepo.InvalidateActiveEmailChangeVerifications(ctx, userID)
	if err := s.authRepo.CreateEmailChangeVerification(ctx, userID, newEmail, otpHash, expiresAt); err != nil {
		return err
	}

	body, err := mailer.RenderEmailChangeOTP(otp, otpTTLMinutes)
	if err != nil {
		return err
	}
	return s.mailer.Send(ctx, newEmail, "Confirm Your Email Change", body)
}

func (s *Service) ConfirmEmailChange(ctx context.Context, userID int64, otp string) error {
	v, err := s.authRepo.FindActiveEmailChangeVerification(ctx, userID)
	if err != nil {
		return err
	}
	if v == nil {
		return ErrNoActiveVerification
	}
	if v.OTPHash != hashOTP(otp) {
		return ErrInvalidOTP
	}

	existing, _ := s.users.FindByEmail(ctx, v.NewEmail)
	if existing != nil && existing.ID != userID {
		_ = s.authRepo.MarkEmailChangeVerificationUsed(ctx, v.ID)
		return ErrEmailAlreadyUsed
	}

	if err := s.users.UpdateEmail(ctx, userID, v.NewEmail); err != nil {
		return err
	}
	_ = s.authRepo.MarkEmailChangeVerificationUsed(ctx, v.ID)
	_ = s.authRepo.RevokeAllRefreshTokens(ctx, userID)
	return nil
}

func generateNumericOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[n.Int64()]
	}
	return string(b), nil
}

func hashOTP(otp string) string {
	h := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(h[:])
}
