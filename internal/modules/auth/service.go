package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
	"uuid"

	"github.com/thakur-vijay/geotrack-backend/internal/modules/sessions"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("name, email and password are required")
)

type Service struct {
	userRepository *users.Repository
	sessionService *sessions.Service
}

func NewService(
	userRepository *users.Repository,
	sessionService *sessions.Service,
) *Service {
	return &Service{
		userRepository: userRepository,
		sessionService: sessionService,
	}
}

func (s *Service) Signup(
	ctx context.Context,
	request SignupRequest,
) (users.User, string, string, error) {
	start := time.Now()
	defer func() {
		log.Printf("[Auth] Signup: %v", time.Since(start))
	}()

	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.FullName == "" ||
		request.Email == "" ||
		strings.TrimSpace(request.Password) == "" {
		return users.User{}, "", "", ErrInvalidInput
	}

	// Hash password before storing it.
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return users.User{}, "", "", err
	}

	now := time.Now().UTC()

	user := users.User{
		ID:           uuid.New().String(),
		FullName:     request.FullName,
		Email:        request.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return users.User{}, "", "", err
	}

	accessToken, refreshToken, err :=
		s.sessionService.CreateSession(ctx, user.ID)
	if err != nil {
		return users.User{}, "", "", err
	}

	// Never expose the password hash in the returned user.
	user.PasswordHash = ""

	return user, accessToken, refreshToken, nil
}

func (s *Service) Login(
	ctx context.Context,
	request LoginRequest,
) (users.User, string, string, error) {
	start := time.Now()
	defer func() {
		log.Printf("[Auth] Login: %v", time.Since(start))
	}()

	email := strings.ToLower(strings.TrimSpace(request.Email))

	if email == "" || request.Password == "" {
		return users.User{}, "", "", ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, "", "", ErrInvalidCredentials
		}

		return users.User{}, "", "", err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(request.Password),
	); err != nil {
		return users.User{}, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err :=
		s.sessionService.CreateSession(ctx, user.ID)
	if err != nil {
		return users.User{}, "", "", err
	}

	user.PasswordHash = ""

	return *user, accessToken, refreshToken, nil
}
