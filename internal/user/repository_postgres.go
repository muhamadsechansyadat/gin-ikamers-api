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

func (r *PostgresRepository) Create(ctx context.Context, u *User) (*User, error) {
	err := r.db.QueryRowContext(ctx, `
                INSERT INTO users (role_id, email, password_hash, is_active, is_email_verified)                                                                        
                VALUES ($1, $2, $3, $4, true)
                RETURNING id, uuid                                                                                                                  
        `, u.RoleID, u.Email, u.PasswordHash, u.IsActive).Scan(&u.ID, &u.UUID)
	if err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT name FROM roles WHERE id = $1`, u.RoleID).Scan(&u.RoleName); err != nil {
		return nil, err
	}

	return u, nil
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_login_at = $1 WHERE id = $2`, at, userID)
	return err
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		passwordHash, userID)
	return err
}

func (r *PostgresRepository) UpdateEmail(ctx context.Context, userID int64, newEmail string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = $1, updated_at = now(), is_email_verified = true WHERE id = $2`,
		newEmail, userID)
	return err
}
