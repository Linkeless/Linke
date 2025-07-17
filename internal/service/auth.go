package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db               *gorm.DB
	userService      *UserService
	jwtService       *JWTService
	inviteCodeService *InviteCodeService
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"` // Optional invite code
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User  *model.UserResponse `json:"user"`
	Token *TokenResponse      `json:"token"`
}

func NewAuthService(db *gorm.DB, userService *UserService, jwtService *JWTService, inviteCodeService *InviteCodeService) *AuthService {
	return &AuthService{
		db:               db,
		userService:      userService,
		jwtService:       jwtService,
		inviteCodeService: inviteCodeService,
	}
}

// Register creates a new user account with email and password
func (a *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Validate invite code if provided
	var inviteCode *model.InviteCode
	if req.InviteCode != "" {
		validatedCode, err := a.inviteCodeService.ValidateInviteCode(ctx, req.InviteCode)
		if err != nil {
			logger.Warn("Invalid invite code used during registration",
				logger.String("email", req.Email),
				logger.String("invite_code", req.InviteCode),
				logger.Error2("error", err),
			)
			return nil, fmt.Errorf("invalid invite code: %s", err.Error())
		}
		inviteCode = validatedCode
	}

	// Check if user already exists
	existingUser, err := a.userService.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to process password")
	}

	// Generate username and name from email
	emailParts := strings.Split(req.Email, "@")
	baseUsername := emailParts[0]
	
	// Generate a unique username by adding random numbers if needed
	username := a.generateUniqueUsername(ctx, baseUsername)
	
	// Generate name from email (capitalize first letter of username)
	name := baseUsername
	if len(baseUsername) > 0 {
		name = strings.ToUpper(string(baseUsername[0])) + baseUsername[1:]
	}

	// Create user
	user := &model.User{
		Email:    req.Email,
		Name:     name,
		Username: username,
		Password: string(hashedPassword),
		Provider: model.ProviderLocal,
		Status:   model.UserStatusActive,
	}

	// Set invite code information if provided
	if inviteCode != nil {
		user.InviteCodeID = &inviteCode.ID
		user.InviteCodeUsed = &inviteCode.Code
	}

	if err := a.userService.CreateUser(ctx, user); err != nil {
		logger.Error("Failed to create user during registration",
			logger.String("email", req.Email),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to create user account")
	}

	// Use the invite code if provided
	if inviteCode != nil {
		// Get IP address and user agent from context (can be enhanced later)
		ipAddress := "unknown"
		userAgent := "unknown"
		
		_, err := a.inviteCodeService.UseInviteCode(ctx, inviteCode.Code, user.ID, ipAddress, userAgent)
		if err != nil {
			logger.Error("Failed to use invite code during registration",
				logger.String("email", req.Email),
				logger.String("invite_code", inviteCode.Code),
				logger.Uint("user_id", user.ID),
				logger.Error2("error", err),
			)
			// Note: We don't return error here to avoid failing registration
			// if invite code usage fails after user creation
		}
	}

	// Generate JWT token
	token, err := a.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("Failed to generate token for new user",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to generate authentication token")
	}

	logger.Info("User registered successfully",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user with email and password
func (a *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Get user by email (first check without status filter for better error messages)
	user, err := a.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Warn("Login attempt with non-existent email",
			logger.String("email", req.Email),
		)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is using local authentication
	if !user.IsLocalAccount() {
		logger.Warn("Login attempt for OAuth user with local credentials",
			logger.String("email", req.Email),
			logger.String("provider", user.Provider),
		)
		return nil, fmt.Errorf("this account uses %s authentication. Please use the appropriate login method", user.Provider)
	}

	// Check user status
	if !user.IsActive() {
		logger.Warn("Login attempt for inactive user",
			logger.String("email", req.Email),
			logger.String("status", user.Status),
		)
		return nil, fmt.Errorf("account is %s. Please contact support", user.Status)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn("Failed login attempt with incorrect password",
			logger.String("email", req.Email),
			logger.Uint("user_id", user.ID),
		)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := a.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("Failed to generate token during login",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to generate authentication token")
	}

	logger.Info("User logged in successfully",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := a.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if user is using local authentication
	if !user.IsLocalAccount() {
		return fmt.Errorf("password change is only available for local accounts")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Warn("Password change attempt with incorrect old password",
			logger.Uint("user_id", userID),
		)
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash new password",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to process new password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := a.userService.UpdateUser(ctx, user); err != nil {
		logger.Error("Failed to update password",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update password")
	}

	logger.Info("Password changed successfully",
		logger.Uint("user_id", userID),
	)

	return nil
}

// ValidateToken validates a JWT token and returns user info
func (a *AuthService) ValidateToken(tokenString string) (*model.User, error) {
	claims, err := a.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get fresh user data from database (only active users)
	user, err := a.userService.GetActiveUserByID(context.Background(), claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found or inactive")
	}

	return user, nil
}

// generateUniqueUsername generates a unique username by checking database for conflicts
func (a *AuthService) generateUniqueUsername(ctx context.Context, baseUsername string) string {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Clean the base username (remove special characters, convert to lowercase)
	baseUsername = strings.ToLower(strings.ReplaceAll(baseUsername, ".", ""))
	baseUsername = strings.ReplaceAll(baseUsername, "+", "")
	baseUsername = strings.ReplaceAll(baseUsername, "_", "")
	
	// If base username is too short, pad it
	if len(baseUsername) < 3 {
		baseUsername = baseUsername + "user"
	}
	
	// Try the base username first
	if !a.usernameExists(ctx, baseUsername) {
		return baseUsername
	}
	
	// If base username exists, try with random numbers
	for attempts := 0; attempts < 10; attempts++ {
		randomNum := rand.Intn(9999) + 1 // 1-9999
		candidate := baseUsername + strconv.Itoa(randomNum)
		
		if !a.usernameExists(ctx, candidate) {
			return candidate
		}
	}
	
	// If all attempts failed, use timestamp
	timestamp := time.Now().Unix()
	return baseUsername + strconv.FormatInt(timestamp, 10)
}

// usernameExists checks if a username already exists in the database
func (a *AuthService) usernameExists(ctx context.Context, username string) bool {
	var count int64
	err := a.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		// If there's an error, assume username exists to be safe
		logger.Error("Error checking username existence",
			logger.String("username", username),
			logger.Error2("error", err),
		)
		return true
	}
	return count > 0
}