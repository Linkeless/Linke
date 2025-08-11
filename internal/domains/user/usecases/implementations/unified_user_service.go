package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/events"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// UnifiedUserService consolidates all user service functionality using composition
type UnifiedUserService struct {
	db            *gorm.DB
	logger        framework.Logger
	caching       CachingBehavior
	events        EventBehavior
	capabilities  ServiceCapabilities
}

// CachingBehavior encapsulates caching functionality
type CachingBehavior interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, patterns ...string) error
	InvalidateUser(ctx context.Context, user *entities.User) error
}

// EventBehavior encapsulates event publishing functionality
type EventBehavior interface {
	PublishUserEvent(ctx context.Context, eventType string, user *entities.User) error
}

// ServiceCapabilities defines what features are enabled
type ServiceCapabilities struct {
	CachingEnabled bool
	EventsEnabled  bool
	MetricsEnabled bool
}

// NewUnifiedUserService creates a new unified user service with configurable behaviors
func NewUnifiedUserService(
	db *gorm.DB,
	logger framework.Logger,
	opts ...ServiceOption,
) *UnifiedUserService {
	service := &UnifiedUserService{
		db:           db,
		logger:       logger,
		capabilities: ServiceCapabilities{},
	}

	// Apply options
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// ServiceOption configures the unified service
type ServiceOption func(*UnifiedUserService)

// WithCaching enables caching behavior
func WithCaching(cache cache.Cache) ServiceOption {
	return func(s *UnifiedUserService) {
		s.caching = &defaultCachingBehavior{cache: cache}
		s.capabilities.CachingEnabled = true
	}
}

// WithEvents enables event publishing
func WithEvents(eventBus events.EventBus) ServiceOption {
	return func(s *UnifiedUserService) {
		s.events = &defaultEventBehavior{eventBus: eventBus}
		s.capabilities.EventsEnabled = true
	}
}

// Core CRUD operations
func (s *UnifiedUserService) CreateUser(ctx context.Context, user *entities.User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		s.logger.Error("Failed to create user",
			logger.String("email", user.Email),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Handle side effects
	s.handleUserCreated(ctx, user)

	s.logger.Info("User created successfully",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)
	return nil
}

func (s *UnifiedUserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	// Try cache first if enabled
	if s.capabilities.CachingEnabled {
		var user entities.User
		if err := s.caching.Get(ctx, s.userCacheKey(id), &user); err == nil {
			return &user, nil
		}
	}

	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache result if enabled
	if s.capabilities.CachingEnabled {
		s.caching.Set(ctx, s.userCacheKey(id), user, 15*time.Minute)
	}

	return &user, nil
}

func (s *UnifiedUserService) UpdateUser(ctx context.Context, user *entities.User) error {
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Handle side effects
	s.handleUserUpdated(ctx, user)

	return nil
}

// Domain-specific operations
func (s *UnifiedUserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Try cache first if enabled
	if s.capabilities.CachingEnabled {
		var user entities.User
		if err := s.caching.Get(ctx, s.userEmailCacheKey(email), &user); err == nil {
			return &user, nil
		}
	}

	var user entities.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Cache result if enabled
	if s.capabilities.CachingEnabled {
		s.caching.Set(ctx, s.userEmailCacheKey(email), user, 15*time.Minute)
	}

	return &user, nil
}

func (s *UnifiedUserService) GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}
	return &user, nil
}

func (s *UnifiedUserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("id = ? AND status = 'active'", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}
	return &user, nil
}

func (s *UnifiedUserService) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("email = ? AND status = 'active'", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		return nil, fmt.Errorf("failed to get active user by email: %w", err)
	}
	return &user, nil
}

// Batch operations
func (s *UnifiedUserService) GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error) {
	var users []*entities.User
	if err := s.db.WithContext(ctx).Find(&users, ids).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	return users, nil
}

func (s *UnifiedUserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	result := &interfaces.BatchOperationResult{}
	
	for _, id := range ids {
		if err := s.SoftDeleteUser(ctx, id); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
		} else {
			result.DeletedCount++
		}
	}
	
	return result, nil
}

func (s *UnifiedUserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	result := &interfaces.BatchOperationResult{}
	
	for _, id := range ids {
		if err := s.RestoreUser(ctx, id); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
		} else {
			result.RestoredCount++
		}
	}
	
	return result, nil
}

// Soft delete operations
func (s *UnifiedUserService) SoftDeleteUser(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&entities.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	// Handle side effects
	if s.capabilities.CachingEnabled {
		s.caching.Delete(ctx, s.userCachePattern(id))
	}

	return nil
}

func (s *UnifiedUserService) RestoreUser(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}
	return nil
}

func (s *UnifiedUserService) HardDeleteUser(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Unscoped().Delete(&entities.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to hard delete user: %w", err)
	}
	return nil
}

// List operations
func (s *UnifiedUserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	if err := s.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if err := s.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (s *UnifiedUserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	if err := s.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&entities.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted users: %w", err)
	}

	if err := s.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list deleted users: %w", err)
	}

	return users, total, nil
}

func (s *UnifiedUserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.User{})
	switch strings.ToLower(provider) {
	case "google":
		query = query.Where("google_id IS NOT NULL")
	case "github":
		query = query.Where("github_id IS NOT NULL")
	case "telegram":
		query = query.Where("telegram_id IS NOT NULL")
	default:
		return nil, 0, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users by provider: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users by provider: %w", err)
	}

	return users, total, nil
}

// Search operations
func (s *UnifiedUserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	searchQuery := s.db.WithContext(ctx).Model(&entities.User{}).
		Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", 
			"%"+query+"%", "%"+query+"%", "%"+query+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	if err := searchQuery.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// Status and role management
func (s *UnifiedUserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Status = status
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Handle side effects
	s.handleUserUpdated(ctx, &user)

	return &user, nil
}

func (s *UnifiedUserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Role = role
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	// Handle side effects
	s.handleUserUpdated(ctx, &user)

	return &user, nil
}

// Statistics
func (s *UnifiedUserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	stats := &interfaces.UserStats{
		ByProvider: make(map[string]int64),
	}

	// Total users
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Active users
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("status = 'active'").Count(&stats.ActiveUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// Provider statistics
	var googleCount, githubCount, telegramCount int64
	s.db.WithContext(ctx).Model(&entities.User{}).Where("google_id IS NOT NULL").Count(&googleCount)
	s.db.WithContext(ctx).Model(&entities.User{}).Where("github_id IS NOT NULL").Count(&githubCount)
	s.db.WithContext(ctx).Model(&entities.User{}).Where("telegram_id IS NOT NULL").Count(&telegramCount)
	
	stats.ByProvider["google"] = googleCount
	stats.ByProvider["github"] = githubCount
	stats.ByProvider["telegram"] = telegramCount

	return stats, nil
}

// Private helper methods
func (s *UnifiedUserService) handleUserCreated(ctx context.Context, user *entities.User) {
	if s.capabilities.EventsEnabled && s.events != nil {
		s.events.PublishUserEvent(ctx, "user.created", user)
	}
}

func (s *UnifiedUserService) handleUserUpdated(ctx context.Context, user *entities.User) {
	if s.capabilities.CachingEnabled && s.caching != nil {
		s.caching.InvalidateUser(ctx, user)
	}
	
	if s.capabilities.EventsEnabled && s.events != nil {
		s.events.PublishUserEvent(ctx, "user.updated", user)
	}
}

// Cache key helpers
func (s *UnifiedUserService) userCacheKey(id uint) string {
	return fmt.Sprintf("user:id:%d", id)
}

func (s *UnifiedUserService) userEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func (s *UnifiedUserService) userCachePattern(id uint) string {
	return fmt.Sprintf("user:*:%d", id)
}

// Default behavior implementations
type defaultCachingBehavior struct {
	cache cache.Cache
}

func (c *defaultCachingBehavior) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, key, data, ttl)
}

func (c *defaultCachingBehavior) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *defaultCachingBehavior) Delete(ctx context.Context, patterns ...string) error {
	for _, pattern := range patterns {
		if err := c.cache.Delete(ctx, pattern); err != nil {
			return err
		}
	}
	return nil
}

func (c *defaultCachingBehavior) InvalidateUser(ctx context.Context, user *entities.User) error {
	keys := []string{
		fmt.Sprintf("user:id:%d", user.ID),
		fmt.Sprintf("user:email:%s", user.Email),
	}
	if user.Username != "" {
		keys = append(keys, fmt.Sprintf("user:username:%s", user.Username))
	}
	return c.Delete(ctx, keys...)
}

type defaultEventBehavior struct {
	eventBus events.EventBus
}

func (e *defaultEventBehavior) PublishUserEvent(ctx context.Context, eventType string, user *entities.User) error {
	event := events.NewUserEvent(eventType, user.ID, map[string]any{
		"user_id":  user.ID,
		"email":    user.Email,
		"username": user.Username,
		"name":     user.Name,
		"status":   user.Status,
		"role":     user.Role,
	})
	return e.eventBus.Publish(ctx, event)
}