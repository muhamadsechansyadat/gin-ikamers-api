package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) LinkIdentity(ctx context.Context, id *Identity) error {
	_, err := r.db.ExecContext(ctx, `
                INSERT INTO user_identities (user_id, provider, provider_user_id, email)                         
                VALUES ($1, $2, $3, $4)                                                                          
        `, id.UserID, id.Provider, id.ProviderUserID, id.Email)
	return err
}

func (r *PostgresRepository) FindUserIDByProviderIdentity(ctx context.Context, provider, providerUserID string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `
                SELECT user_id FROM user_identities
                WHERE provider = $1 AND provider_user_id = $2
                LIMIT 1
        `, provider, providerUserID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *PostgresRepository) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ua, ip string) error {
	var uaPtr *string
	if ua != "" {
		uaPtr = &ua
	}
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	_, err := r.db.ExecContext(ctx, `
                INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
                VALUES ($1, $2, $3, $4, $5)
        `, userID, tokenHash, uaPtr, ipPtr, expiresAt)
	return err
}

func (r *PostgresRepository) FindActiveRefreshToken(ctx context.Context, tokenHash string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `                                                                       
                SELECT user_id FROM refresh_tokens                                                               
                WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
                LIMIT 1                                                                                          
        `, tokenHash).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`,
		tokenHash)
	return err
}

func (r *PostgresRepository) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revo
   IS NULL`, userID)
	return err
}
