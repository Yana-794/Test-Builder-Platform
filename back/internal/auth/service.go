package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	
)

const (
	sessionLifetime = 7 * 24 * time.Hour
	maxPasswordBytes = 128
)

type Service struct {
	repo     *Repository
	dummyHash string
}

func NewService(repo *Repository) (*Service, error) {
	// Используется, чтобы при неизвестном email тоже выполнять
	// дорогую проверку пароля, а не сразу возвращать ответ.
	dummyHash, err := hashPassword("dummy-password-for-timing-check")
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:      repo,
		dummyHash: dummyHash,
	}, nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))

	if email == "" || len(email) > 254 ||
		strings.ContainsAny(email, "\r\n") {
		return "", ErrInvalidInput
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidInput
	}

	return email, nil
}

func (s *Service) Register(
	ctx context.Context,
	email string,
	displayName string,
	password string,
) (User, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return User{}, ErrInvalidInput
	}

	displayName = strings.TrimSpace(displayName)

	if utf8.RuneCountInString(displayName) < 1 ||
		utf8.RuneCountInString(displayName) > 100 {
		return User{}, ErrInvalidInput
	}

	if utf8.RuneCountInString(password) < 12 ||
		len(password) > maxPasswordBytes {
		return User{}, ErrInvalidInput
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}

	return s.repo.CreateUser(
		ctx,
		normalizedEmail,
		displayName,
		passwordHash,
	)
}

func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
) (LoginResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil || password == "" || len(password) > maxPasswordBytes {
		_ = verifyPassword(password, s.dummyHash)
		return LoginResult{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, normalizedEmail)

	if errors.Is(err, ErrUserNotFound) {
		_ = verifyPassword(password, s.dummyHash)
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}

	if !verifyPassword(password, user.PasswordHash) {
		return LoginResult{}, ErrInvalidCredentials
	}

	// Каждый успешный вход получает новый случайный токен.
	rawSecret := make([]byte, 32)
	if _, err := rand.Read(rawSecret); err != nil {
		return LoginResult{}, err
	}
	sessionToken := base64.RawURLEncoding.EncodeToString(rawSecret)

	csrfBytes := make([]byte, 32)
	if _, err := rand.Read(csrfBytes); err != nil {
		return LoginResult{}, err
	}
	csrfToken := base64.RawURLEncoding.EncodeToString(csrfBytes)

	tokenHash := sha256.Sum256([]byte(sessionToken))
	expiresAt := time.Now().UTC().Add(sessionLifetime)

	if err := s.repo.CreateSession(
		ctx,
		user.ID,
		tokenHash[:],
		csrfBytes,
		expiresAt,
	); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:         user,
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Service) Authenticate(
	ctx context.Context,
	sessionToken string,
) (AuthenticatedSession, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(sessionToken)
	if err != nil || len(decoded) != 32 {
		return AuthenticatedSession{}, ErrUnauthorized
	}

	tokenHash := sha256.Sum256([]byte(sessionToken))

	stored, err := s.repo.GetSessionByTokenHash(ctx, tokenHash[:])
	if errors.Is(err, ErrSessionNotFound) {
		return AuthenticatedSession{}, ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedSession{}, err
	}

	if len(stored.CSRFToken) != 32 {
		return AuthenticatedSession{}, ErrUnauthorized
	}

	return AuthenticatedSession{
		User:      stored.User,
		CSRFToken: base64.RawURLEncoding.EncodeToString(stored.CSRFToken),
		ExpiresAt: stored.ExpiresAt,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	sessionToken string,
) error {
	decoded, err := base64.RawURLEncoding.DecodeString(sessionToken)
	if err != nil || len(decoded) != 32 {
		return ErrUnauthorized
	}

	tokenHash := sha256.Sum256([]byte(sessionToken))
	return s.repo.RevokeSession(ctx, tokenHash[:])
}

