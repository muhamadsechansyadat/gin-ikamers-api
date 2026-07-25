package user

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

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
                SELECT u.id, u.uuid, u.role_id, r.name, u.email, u.password_hash, u.is_active, u.is_email_verified                                  
                FROM users u                                                                                                                        
                JOIN roles r ON r.id = u.role_id
                WHERE lower(u.email) = lower($1) AND u.deleted_at IS NULL                                                                           
                LIMIT 1                                                                                                                             
        `
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.UUID, &u.RoleID, &u.RoleName, &u.Email,
		&u.PasswordHash, &u.IsActive, &u.IsEmailVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	query := `
                SELECT u.id, u.uuid, u.role_id, r.name, u.email, u.password_hash, u.is_active, u.is_email_verified
                FROM users u                                                                                                                        
                JOIN roles r ON r.id = u.role_id
                WHERE u.id = $1 AND u.deleted_at IS NULL                                                                                            
                LIMIT 1                                                                                                                           
        `
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.UUID, &u.RoleID, &u.RoleName, &u.Email,
		&u.PasswordHash, &u.IsActive, &u.IsEmailVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) FindByProviderIdentity(ctx context.Context, provider, providerUserID string) (*User, error) {
	query := `
                SELECT u.id, u.uuid, u.role_id, r.name, u.email, u.password_hash, u.is_active, u.is_email_verified                                  
                FROM users u                                                                                                                        
                JOIN roles r ON r.id = u.role_id
                JOIN user_identities i ON i.user_id = u.id                                                                                          
                WHERE i.provider = $1 AND i.provider_user_id = $2 AND u.deleted_at IS NULL                                                        
                LIMIT 1                                                                                                                             
        `
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&u.ID, &u.UUID, &u.RoleID, &u.RoleName, &u.Email,
		&u.PasswordHash, &u.IsActive, &u.IsEmailVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) CreateWithIdentity(ctx context.Context, u *User, id *Identity) (*User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
                INSERT INTO users (role_id, email, password_hash, is_active)                                                                        
                VALUES ($1, $2, $3, $4)                                                                                                           
                RETURNING id, uuid                                                                                                                  
        `, u.RoleID, u.Email, u.PasswordHash, u.IsActive).Scan(&u.ID, &u.UUID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
                INSERT INTO user_identities (user_id, provider, provider_user_id, email)
                VALUES ($1, $2, $3, $4)                                                                                                             
        `, u.ID, id.Provider, id.ProviderUserID, id.Email)
	if err != nil {
		return nil, err
	}

	if err := tx.QueryRowContext(ctx, `SELECT name FROM roles WHERE id = $1`, u.RoleID).Scan(&u.RoleName); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) LinkIdentity(ctx context.Context, id *Identity) error {
	_, err := r.db.ExecContext(ctx, `
                INSERT INTO user_identities (user_id, provider, provider_user_id, email)                                                            
                VALUES ($1, $2, $3, $4)                                                                                                             
        `, id.UserID, id.Provider, id.ProviderUserID, id.Email)
	return err
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_login_at = $1 WHERE id = $2`, at, userID)
	return err
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
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *PostgresRepository) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
