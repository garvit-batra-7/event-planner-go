package auth

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"event-planner-go/services/auth/internal/jwt"   // <-- add this
	"event-planner-go/services/auth/internal/user"  // <-- your user repo/model
	"event-planner-go/pkg/shared"
	"event-planner-go/services/auth/internal/config"
)

type Service struct {
	userRepo user.Repository
	config   *config.Config
	logger   *shared.Logger
}

func NewService(userRepo user.Repository, config *config.Config, logger *shared.Logger) *Service {
	return &Service{
		userRepo: userRepo,
		config:   config,
		logger:   logger,
	}
}

// Login authenticates a user and returns a token
func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.Error("Database error during login", "error", err.Error())
		return nil, fmt.Errorf("internal error")
	}

	if existingUser == nil {
		s.logger.Warn("Login attempt with non-existent email", "email", req.Email)
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		s.logger.LogUser(shared.LevelWarn, fmt.Sprintf("%d", existingUser.ID), "login_failed", "Invalid password attempt")
		return nil, errors.New("invalid email or password")
	}

	// Generate token
	token, err := jwt.GenerateToken(existingUser.ID, s.config.Auth.JWTSecret, s.config.Auth.JWTExpiration)
	if err != nil {
		s.logger.Error("Failed to generate JWT", "error", err.Error())
		return nil, fmt.Errorf("failed to generate token")
	}

	s.logger.LogUser(shared.LevelInfo, fmt.Sprintf("%d", existingUser.ID), "login", "User logged in successfully",
		"email", req.Email,
	)

	return &AuthResponse{
		Token:     token,
		ExpiresIn: s.config.Auth.JWTExpiration.Seconds(),
		User: UserDTO{
			ID:    existingUser.ID,
			Email: existingUser.Email,
			Name:  existingUser.Name,
		},
	}, nil
}

// Register creates a new user account
func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Validate password length
	if len(req.Password) < s.config.Auth.PasswordMinLength {
		s.logger.Warn("Password too short", "length", len(req.Password))
		return nil, fmt.Errorf("password must be at least %d characters", s.config.Auth.PasswordMinLength)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.Error("Database error during registration", "error", err.Error())
		return nil, fmt.Errorf("internal error")
	}

	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err.Error())
		return nil, fmt.Errorf("internal error")
	}

	// Create user
	newUser := &user.User{            // <-- use user.User
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err.Error(), "email", req.Email)
		return nil, fmt.Errorf("failed to create user")
	}

	s.logger.LogUser(shared.LevelInfo, fmt.Sprintf("%d", newUser.ID), "register", "User registered successfully",
		"email", req.Email,
	)

	// Generate token for newly registered user
	token, err := jwt.GenerateToken(newUser.ID, s.config.Auth.JWTSecret, s.config.Auth.JWTExpiration)
	if err != nil {
		s.logger.Error("Failed to generate JWT", "error", err.Error())
		return nil, fmt.Errorf("failed to generate token")
	}

	return &AuthResponse{
		Token:     token,
		ExpiresIn: s.config.Auth.JWTExpiration.Seconds(),
		User: UserDTO{
			ID:    newUser.ID,
			Email: newUser.Email,
			Name:  newUser.Name,
		},
	}, nil
}

func (s *Service) GetByID(id int) (*user.User, error) {
    return s.userRepo.GetByID(id)
}
