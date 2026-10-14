package auth
import (
	"context"
	"errors"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanUser(row pgx.Row) (User, error) {
	var user User
	var verified pgtype.Timestamptz

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}

	if verified.Valid {
		t := verified.Time
		user.EmailVerifiedAt = &t
	}

	return user, nil
}

func (r *Repository) CreateUser(
	ctx context.Context,
	email string,
	displayName string,
	passwordHash string,
) (User, error) {
	const query = `
		INSERT INTO users (email, display_name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING
			id, email, display_name, password_hash,
			email_verified_at, created_at, updated_at
	`

	user, err := scanUser(
		r.pool.QueryRow(ctx, query, email, displayName, passwordHash),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_users_email" {
			return User{}, ErrEmailExists
		}

		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	const query = `
		SELECT
			id, email, display_name, password_hash,
			email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash []byte,
	csrfToken []byte,
	expiresAt time.Time,
) error {
	const query = `
		INSERT INTO auth_sessions
			(user_id, token_hash, csrf_token, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		userID,
		tokenHash,
		csrfToken,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *Repository) GetSessionByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (StoredSession, error) {
	const query = `
		SELECT
			u.id, u.email, u.display_name,
			u.email_verified_at, u.created_at, u.updated_at,
			s.csrf_token, s.expires_at
		FROM auth_sessions AS s
		JOIN users AS u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()
	`

	var stored StoredSession
	var verified pgtype.Timestamptz

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&stored.User.ID,
		&stored.User.Email,
		&stored.User.DisplayName,
		&verified,
		&stored.User.CreatedAt,
		&stored.User.UpdatedAt,
		&stored.CSRFToken,
		&stored.ExpiresAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return StoredSession{}, ErrSessionNotFound
	}
	if err != nil {
		return StoredSession{}, fmt.Errorf("get session: %w", err)
	}

	if verified.Valid {
		t := verified.Time
		stored.User.EmailVerifiedAt = &t
	}

	return stored, nil
}

func (r *Repository) RevokeSession(
	ctx context.Context,
	tokenHash []byte,
) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = now()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`

	if _, err := r.pool.Exec(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}