package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger" 
	"linke/internal/shared/services"

	"gorm.io/gorm"
)

// EnhancedUserService demonstrates how to migrate to the new framework
// while maintaining backward compatibility with existing interfaces
type EnhancedUserService struct {
	// Embed the generic service for full framework support
	*services.BaseServiceImpl[entities.User, uint]
	
	// Add domain-specific mixins
	*services.LookupServiceMixin[entities.User, uint]
	*services.CacheableServiceMixin[entities.User, uint]
	*services.SearchableServiceMixin[entities.User, uint]
	
	// Legacy service for backward compatibility during transition
	legacyService *UserService
	
	// Domain-specific dependencies
	db     *gorm.DB
	logger framework.Logger
	cache  framework.Cache
}

// NewEnhancedUserService creates a new enhanced user service
func NewEnhancedUserService(
	db *gorm.DB,
	logger framework.Logger,
	cache framework.Cache,
	eventPub framework.EventPublisher,
	validator framework.Validator,
	repository framework.GenericRepository[entities.User, uint],
) *EnhancedUserService {
	// Create base service
	baseService := services.NewBaseService(
		"user-service",
		repository,
		logger,
		eventPub,
		validator,
	)

	// Create mixins
	lookupMixin := services.NewLookupServiceMixin(baseService, map[string]string{
		"email":       "email",
		"username":    "username", 
		"telegram_id": "telegram_id",
		"provider_id": "provider_id",
	})

	cacheableMixin := services.NewCacheableServiceMixin(baseService, cache)
	
	searchableMixin := services.NewSearchableServiceMixin(baseService, []string{
		"username", "email", "first_name", "last_name", "display_name",
	})

	// Create legacy service for backward compatibility
	legacyService := NewUserService(db, logger)

	return &EnhancedUserService{
		BaseServiceImpl:         baseService,
		LookupServiceMixin:      lookupMixin,
		CacheableServiceMixin:   cacheableMixin,
		SearchableServiceMixin:  searchableMixin,
		legacyService:           legacyService,
		db:                      db,
		logger:                  logger,
		cache:                   cache,
	}
}

// ===========================================
// Legacy Interface Compatibility Layer
// These methods maintain the original signatures while delegating to the new framework
// ===========================================

// CreateUser implements the legacy interface
func (s *EnhancedUserService) CreateUser(ctx context.Context, user *entities.User) error {
	req := &framework.CreateRequest[entities.User]{
		Data: user,
		Options: &framework.CreateOptions{
			PublishEvents:    true,
			EnableAuditLog:   true,
			ProcessWorkflows: true,
		},
	}
	
	_, err := s.Create(ctx, req)
	return err
}

// GetUserByID implements the legacy interface
func (s *EnhancedUserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	return s.GetByID(ctx, id)
}

// UpdateUser implements the legacy interface  
func (s *EnhancedUserService) UpdateUser(ctx context.Context, user *entities.User) error {
	req := &framework.UpdateRequest[entities.User]{
		Data: user,
		Options: &framework.UpdateOptions{
			PublishEvents:    true,
			EnableAuditLog:   true,
			ProcessWorkflows: true,
		},
	}
	
	_, err := s.Update(ctx, user.ID, req)
	return err
}

// GetUserByEmail uses the LookupServiceMixin
func (s *EnhancedUserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return s.GetByField(ctx, "email", email)
}

// GetUserByTelegramID uses the LookupServiceMixin  
func (s *EnhancedUserService) GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	return s.GetByField(ctx, "telegram_id", telegramID)
}

// GetActiveUserByID combines generic lookup with domain logic
func (s *EnhancedUserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if user.Status != "active" {
		return nil, fmt.Errorf("user is not active")
	}
	
	return user, nil
}

// GetActiveUserByEmail combines lookup mixin with domain logic
func (s *EnhancedUserService) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.GetByField(ctx, "email", email)
	if err != nil {
		return nil, err
	}
	
	if user.Status != "active" {
		return nil, fmt.Errorf("user is not active")
	}
	
	return user, nil
}

// GetUsersByIDs delegates to legacy implementation during transition
func (s *EnhancedUserService) GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error) {
	// TODO: Implement using generic framework batch operations
	return s.legacyService.GetUsersByIDs(ctx, ids)
}

// Soft delete operations map to generic framework
func (s *EnhancedUserService) SoftDeleteUser(ctx context.Context, id uint) error {
	return s.SoftDelete(ctx, id)
}

func (s *EnhancedUserService) RestoreUser(ctx context.Context, id uint) error {
	return s.Restore(ctx, id)
}

func (s *EnhancedUserService) HardDeleteUser(ctx context.Context, id uint) error {
	return s.HardDelete(ctx, id)
}

// List operations map to generic framework
func (s *EnhancedUserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{
		Limit:  limit,
		Offset: offset,
	}
	
	response, err := s.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	
	return response.Data, response.Total, nil
}

func (s *EnhancedUserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{
		Limit:  limit,
		Offset: offset,
	}
	
	response, err := s.ListDeleted(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	
	return response.Data, response.Total, nil
}

func (s *EnhancedUserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{
		Limit:   limit,
		Offset:  offset,
		Filters: map[string]any{"provider": provider},
	}
	
	response, err := s.ListWithFilters(ctx, map[string]any{"provider": provider}, req)
	if err != nil {
		return nil, 0, err
	}
	
	return response.Data, response.Total, nil
}

// Search operations use SearchableServiceMixin
func (s *EnhancedUserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{
		Limit:  limit,
		Offset: offset,
	}
	
	response, err := s.Search(ctx, query, req)
	if err != nil {
		return nil, 0, err
	}
	
	return response.Data, response.Total, nil
}

// Status and role management use generic framework
func (s *EnhancedUserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	return s.UpdateStatus(ctx, id, status)
}

func (s *EnhancedUserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	// Custom implementation for role updates
	updates := map[string]any{
		"role":            role,
		"role_updated_at": time.Now(),
	}
	
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update user role", 
			logger.Uint("user_id", id), 
			logger.String("role", role), 
			logger.ErrorField(err))
		return nil, fmt.Errorf("update user role: %w", err)
	}
	
	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	s.InvalidateCache(ctx, cacheKey)
	
	return s.GetByID(ctx, id)
}

// Statistics mapping
func (s *EnhancedUserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	baseStats, err := s.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}
	
	// Get provider counts
	providerCounts := make(map[string]int64)
	providers := []string{"local", "google", "github", "telegram"}
	
	for _, provider := range providers {
		count, err := s.CountByStatus(ctx, provider)
		if err != nil {
			s.logger.Warn("Failed to count users by provider", logger.String("provider", provider), logger.ErrorField(err))
			continue
		}
		providerCounts[provider] = count
	}
	
	// Get recent signups (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentSignups int64
	if err := s.db.Model(&entities.User{}).Where("created_at >= ?", thirtyDaysAgo).Count(&recentSignups).Error; err != nil {
		s.logger.Warn("Failed to count recent signups", logger.ErrorField(err))
		recentSignups = 0
	}
	
	return &interfaces.UserStats{
		TotalUsers:    baseStats.TotalCount,
		ActiveUsers:   baseStats.ActiveCount,
		InactiveUsers: baseStats.InactiveCount,
		BannedUsers:   baseStats.StatusCounts["banned"],
		DeletedUsers:  baseStats.DeletedCount,
		ByProvider:    providerCounts,
		RecentSignups: recentSignups,
	}, nil
}

// Batch operations map to generic framework
func (s *EnhancedUserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	response, err := s.BatchDelete(ctx, ids)
	if err != nil {
		return nil, err
	}
	
	return &interfaces.BatchOperationResult{
		DeletedCount: response.SuccessCount,
		FailedIDs:    response.FailedIDs,
	}, nil
}

func (s *EnhancedUserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	response, err := s.BatchRestore(ctx, ids)
	if err != nil {
		return nil, err
	}
	
	return &interfaces.BatchOperationResult{
		RestoredCount: response.SuccessCount,
		FailedIDs:     response.FailedIDs,
	}, nil
}

// ===========================================
// Enhanced Methods Using New Framework Features
// ===========================================

// GetUsersByEmails efficiently retrieves multiple users by email
func (s *EnhancedUserService) GetUsersByEmails(ctx context.Context, emails []string) ([]*entities.User, error) {
	if len(emails) == 0 {
		return []*entities.User{}, nil
	}
	
	var users []*entities.User
	if err := s.db.WithContext(ctx).Where("email IN (?)", emails).Find(&users).Error; err != nil {
		s.logger.Error("Failed to get users by emails", logger.Any("emails", emails), logger.ErrorField(err))
		return nil, fmt.Errorf("get users by emails: %w", err)
	}
	
	return users, nil
}

// GetUserProfile returns user profile with additional computed fields
func (s *EnhancedUserService) GetUserProfile(ctx context.Context, id uint) (*entities.UserProfile, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Create profile with computed fields
	profile := &entities.UserProfile{
		User:              *user,
		IsEmailVerified:   true, // Simplified - check if email verification exists
		AccountAge:        int(time.Since(user.CreatedAt).Hours() / 24), // days
		LastLoginDaysAgo:  0, // Simplified - would need last_login_at field
		HasProfilePicture: user.Avatar != "",
	}
	
	return profile, nil
}

// SearchUsersAdvanced uses the enhanced search capabilities
func (s *EnhancedUserService) SearchUsersAdvanced(ctx context.Context, req *interfaces.AdvancedUserSearchRequest) (*framework.SearchResponse[entities.User], error) {
	listReq := &framework.ListRequest{
		Limit:   req.Limit,
		Offset:  req.Offset,
		Filters: make(map[string]any),
	}
	
	// Add filters
	if req.Status != "" {
		listReq.Filters["status"] = req.Status
	}
	if req.Provider != "" {
		listReq.Filters["provider"] = req.Provider
	}
	if req.Role != "" {
		listReq.Filters["role"] = req.Role
	}
	if req.EmailVerified != nil {
		if *req.EmailVerified {
			listReq.Filters["email_verified_at IS NOT NULL"] = true
		} else {
			listReq.Filters["email_verified_at IS NULL"] = true
		}
	}
	
	if req.Query != "" {
		return s.SearchWithHighlight(ctx, req.Query, []string{"username", "email", "display_name"}, listReq)
	}
	
	// If no search query, just return filtered results
	listResponse, err := s.ListWithFilters(ctx, listReq.Filters, listReq)
	if err != nil {
		return nil, err
	}
	
	return &framework.SearchResponse[entities.User]{
		ListResponse: *listResponse,
		SearchTime:   5, // mock search time
		TotalMatches: listResponse.Total,
	}, nil
}

// WarmUserCache warms cache for frequently accessed users
func (s *EnhancedUserService) WarmUserCache(ctx context.Context, userIDs []uint) error {
	return s.WarmCache(ctx, userIDs)
}

// InvalidateUserCache invalidates cache for specific users
func (s *EnhancedUserService) InvalidateUserCache(ctx context.Context, userIDs []uint) error {
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = fmt.Sprintf("user:%d", id)
	}
	return s.InvalidateCache(ctx, keys...)
}

// ===========================================
// Domain-Specific Business Logic
// ===========================================

// ActivateUser activates a user account with domain-specific logic
func (s *EnhancedUserService) ActivateUser(ctx context.Context, id uint) (*entities.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if user.Status == "active" {
		return user, nil // Already active
	}
	
	// Update status with timestamp
	updates := map[string]any{
		"status":       "active",
		"activated_at": time.Now(),
	}
	
	if err := s.db.WithContext(ctx).Model(user).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to activate user", logger.Uint("user_id", id), logger.ErrorField(err))
		return nil, fmt.Errorf("activate user: %w", err)
	}
	
	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	s.InvalidateCache(ctx, cacheKey)
	
	// Publish user activated event
	// TODO: Implement event publishing
	
	s.logger.Info("User activated", logger.Uint("user_id", id))
	return s.GetByID(ctx, id)
}

// DeactivateUser deactivates a user account
func (s *EnhancedUserService) DeactivateUser(ctx context.Context, id uint, reason string) (*entities.User, error) {
	updates := map[string]any{
		"status":              "inactive",
		"deactivated_at":      time.Now(),
		"deactivation_reason": reason,
	}
	
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to deactivate user", logger.Uint("user_id", id), logger.ErrorField(err))
		return nil, fmt.Errorf("deactivate user: %w", err)
	}
	
	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	s.InvalidateCache(ctx, cacheKey)
	
	s.logger.Info("User deactivated", logger.Uint("user_id", id), logger.String("reason", reason))
	return s.GetByID(ctx, id)
}