package implementations

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/domains/auth/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	authEntities "linke/internal/domains/auth/entities"
	referralEntities "linke/internal/domains/referral/entities"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db                  *gorm.DB
	userService         interfaces.UserService
	jwtService          interfaces.JWTService
	inviteCodeService   interfaces.InviteCodeService
	loginSecurityService interfaces.LoginSecurityService
}


func NewAuthService(db *gorm.DB, userService interfaces.UserService, jwtService interfaces.JWTService, inviteCodeService interfaces.InviteCodeService, loginSecurityService interfaces.LoginSecurityService) *AuthService {
	return &AuthService{
		db:                  db,
		userService:         userService,
		jwtService:          jwtService,
		inviteCodeService:   inviteCodeService,
		loginSecurityService: loginSecurityService,
	}
}

// Register creates a new user account with email and password
func (a *AuthService) Register(ctx context.Context, req *interfaces.RegisterRequest) (*interfaces.AuthResponse, error) {
	// Validate invite code if provided
	var inviteCode *referralEntities.InviteCode
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
	user := &userEntities.User{
		Email:    req.Email,
		Name:     name,
		Username: username,
		Password: string(hashedPassword),
		Provider: userEntities.ProviderLocal,
		Status:   userEntities.UserStatusActive,
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
		} else {
			logger.Info("Invite code used successfully and referral created",
				logger.String("email", req.Email),
				logger.String("invite_code", inviteCode.Code),
				logger.Uint("user_id", user.ID),
				logger.Uint("referrer_id", inviteCode.CreatedByID),
			)
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

	return &interfaces.AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user with email and password
func (a *AuthService) Login(ctx context.Context, req *interfaces.LoginRequest) (*interfaces.AuthResponse, error) {
	// Extract IP and User-Agent from context (set by middleware)
	ip, _ := ctx.Value("client_ip").(string)
	userAgent, _ := ctx.Value("user_agent").(string)
	if ip == "" {
		ip = "unknown"
	}
	if userAgent == "" {
		userAgent = "unknown"
	}

	// Check if account is locked
	if a.loginSecurityService != nil {
		isLocked, lockout, err := a.loginSecurityService.IsAccountLocked(ctx, req.Email)
		if err != nil {
			logger.Error("Failed to check account lockout status",
				logger.String("email", req.Email),
				logger.Error2("error", err))
			// Continue with login attempt despite error
		} else if isLocked {
			logger.Warn("Login attempt on locked account",
				logger.String("email", req.Email),
				logger.String("ip", ip),
				logger.Duration("remaining_lock_time", lockout.GetRemainingLockTime()))
			
			// Record the failed attempt
			a.recordLoginAttempt(ctx, req.Email, ip, userAgent, authEntities.LoginFailureAccountLocked, false, nil)
			
			return nil, fmt.Errorf("account is temporarily locked due to multiple failed login attempts. Please try again in %v", 
				lockout.GetRemainingLockTime().Round(time.Minute))
		}
	}

	// Get user by email (first check without status filter for better error messages)
	user, err := a.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Warn("Login attempt with non-existent email",
			logger.String("email", req.Email),
			logger.String("ip", ip))
		
		// Record failed attempt
		a.recordLoginAttempt(ctx, req.Email, ip, userAgent, authEntities.LoginFailureUserNotFound, false, nil)
		
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is using local authentication
	if !user.IsLocalAccount() {
		logger.Warn("Login attempt for OAuth user with local credentials",
			logger.String("email", req.Email),
			logger.String("provider", user.Provider),
			logger.String("ip", ip))
		
		// Record failed attempt
		a.recordLoginAttempt(ctx, req.Email, ip, userAgent, authEntities.LoginFailureOAuthMismatch, false, &user.ID)
		
		return nil, fmt.Errorf("this account uses %s authentication. Please use the appropriate login method", user.Provider)
	}

	// Check user status
	if !user.IsActive() {
		var reason string
		switch user.Status {
		case userEntities.UserStatusInactive:
			reason = authEntities.LoginFailureAccountInactive
		case userEntities.UserStatusBanned:
			reason = authEntities.LoginFailureAccountBanned
		default:
			reason = authEntities.LoginFailureAccountInactive
		}

		logger.Warn("Login attempt for inactive user",
			logger.String("email", req.Email),
			logger.String("status", user.Status),
			logger.String("ip", ip))
		
		// Record failed attempt
		a.recordLoginAttempt(ctx, req.Email, ip, userAgent, reason, false, &user.ID)
		
		return nil, fmt.Errorf("account is %s. Please contact support", user.Status)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn("Failed login attempt with incorrect password",
			logger.String("email", req.Email),
			logger.String("ip", ip),
			logger.Uint("user_id", user.ID))
		
		// Record failed attempt
		a.recordLoginAttempt(ctx, req.Email, ip, userAgent, authEntities.LoginFailureInvalidCredentials, false, &user.ID)
		
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

	// Record successful login attempt
	a.recordLoginAttempt(ctx, req.Email, ip, userAgent, authEntities.LoginSuccessLocal, true, &user.ID)

	logger.Info("User logged in successfully",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
		logger.String("ip", ip))

	return &interfaces.AuthResponse{
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

	// Revoke all existing tokens for security
	if err := a.jwtService.RevokeAllUserTokens(userID, authEntities.BlacklistReasonPasswordChange); err != nil {
		logger.Warn("Failed to revoke tokens after password change",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		// Don't fail the password change if token revocation fails
	}

	logger.Info("Password changed successfully",
		logger.Uint("user_id", userID),
	)

	return nil
}

// AdminResetPassword resets a user's password by admin (generates a new temporary password)
func (a *AuthService) AdminResetPassword(ctx context.Context, adminUserID, targetUserID uint, newPassword string) error {
	// Get admin user to verify permissions
	adminUser, err := a.userService.GetUserByID(ctx, adminUserID)
	if err != nil {
		return fmt.Errorf("admin user not found")
	}

	if !adminUser.IsAdmin() {
		return fmt.Errorf("insufficient permissions: admin role required")
	}

	// Get target user
	user, err := a.userService.GetUserByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found")
	}

	// Only allow password reset for local accounts
	if !user.IsLocalAccount() {
		return fmt.Errorf("password reset is only available for local accounts, user uses %s authentication", user.Provider)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash new password for admin reset",
			logger.Uint("admin_id", adminUserID),
			logger.Uint("target_user_id", targetUserID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to process new password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := a.userService.UpdateUser(ctx, user); err != nil {
		logger.Error("Failed to update password during admin reset",
			logger.Uint("admin_id", adminUserID),
			logger.Uint("target_user_id", targetUserID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update password")
	}

	// Revoke all existing tokens for security
	if err := a.jwtService.RevokeAllUserTokens(targetUserID, authEntities.BlacklistReasonPasswordChange); err != nil {
		logger.Warn("Failed to revoke tokens after admin password reset",
			logger.Uint("admin_id", adminUserID),
			logger.Uint("target_user_id", targetUserID),
			logger.Error2("error", err))
		// Don't fail the password reset if token revocation fails
	}

	logger.Info("Password reset by admin",
		logger.Uint("admin_id", adminUserID),
		logger.String("admin_email", adminUser.Email),
		logger.Uint("target_user_id", targetUserID),
		logger.String("target_email", user.Email),
	)

	return nil
}

// ValidateToken validates a JWT token and returns user info from cached JWT claims
func (a *AuthService) ValidateToken(tokenString string) (*userEntities.User, error) {
	claims, err := a.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Build user object from JWT claims (cached data, no database query needed)
	user := &userEntities.User{
		ID:       claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Provider: claims.Provider,
		Status:   claims.Status,
		Role:     claims.Role,
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
	err := a.db.WithContext(ctx).Model(&userEntities.User{}).Where("username = ?", username).Count(&count).Error
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

// recordLoginAttempt is a helper method to record login attempts
func (a *AuthService) recordLoginAttempt(ctx context.Context, email, ip, userAgent, reason string, success bool, userID *uint) {
	if a.loginSecurityService == nil {
		return
	}

	if err := a.loginSecurityService.RecordLoginAttempt(ctx, email, ip, userAgent, reason, success, userID); err != nil {
		logger.Error("Failed to record login attempt",
			logger.String("email", email),
			logger.String("ip", ip),
			logger.String("success", fmt.Sprintf("%t", success)),
			logger.Error2("error", err))
	}
}

// Logout revokes the current JWT token
func (a *AuthService) Logout(ctx context.Context, tokenString string, userID uint) error {
	if err := a.jwtService.RevokeToken(tokenString, &userID, authEntities.BlacklistReasonLogout); err != nil {
		logger.Error("Failed to revoke token during logout",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to logout securely")
	}

	logger.Info("User logged out successfully",
		logger.Uint("user_id", userID))

	return nil
}