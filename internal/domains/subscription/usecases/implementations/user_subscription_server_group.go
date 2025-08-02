package implementations

import (
	"context"
	"fmt"

	serverEntities "linke/internal/domains/server/entities"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/logger"
)

// ============= Server Group Management for User Subscriptions =============

// UpdateSubscriptionServerGroupsRequest represents the request to update server groups for a subscription
type UpdateSubscriptionServerGroupsRequest struct {
	SubscriptionID uint   `json:"subscription_id" binding:"required" example:"1"`
	ServerGroupIDs []uint `json:"server_group_ids" binding:"required" example:"[1,2,3]"`
}

// GetAvailableServerGroups returns available server groups for a user's subscription
func (s *UserSubscriptionService) GetAvailableServerGroups(ctx context.Context, userID uint, subscriptionID uint) ([]*serverEntities.ServerGroup, error) {
	// Get the subscription to verify ownership
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}

	// Get the subscription plan to see default server groups
	plan, err := s.subscriptionPlanService.GetSubscriptionPlan(ctx, subscription.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Get all server groups (no status filtering for now as ServerGroup doesn't have status field)
	var availableGroups []*serverEntities.ServerGroup
	query := s.db.WithContext(ctx)

	// If plan has specific server group restrictions, apply them
	// Note: Plan model doesn't have DefaultServerGroupIDs field yet, will implement later
	// For now, return all server groups
	_ = plan // Use plan variable to avoid unused variable error

	if err := query.Find(&availableGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to get available server groups: %w", err)
	}

	return availableGroups, nil
}

// GetSubscriptionServerGroups returns the server groups assigned to a subscription
func (s *UserSubscriptionService) GetSubscriptionServerGroups(ctx context.Context, userID uint, subscriptionID uint) ([]*serverEntities.ServerGroup, error) {
	// Get the subscription to verify ownership and get assigned groups
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}

	// Get assigned server group IDs
	assignedGroupIDs := subscription.GetServerGroupIDs()
	if len(assignedGroupIDs) == 0 {
		return []*serverEntities.ServerGroup{}, nil
	}

	// Get server group details
	var serverGroups []*serverEntities.ServerGroup
	if err := s.db.WithContext(ctx).
		Where("id IN (?)", assignedGroupIDs).
		Find(&serverGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to get assigned server groups: %w", err)
	}

	return serverGroups, nil
}

// UpdateSubscriptionServerGroups updates the server groups for a user's subscription
func (s *UserSubscriptionService) UpdateSubscriptionServerGroups(ctx context.Context, userID uint, req *UpdateSubscriptionServerGroupsRequest) (*entities.UserSubscription, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription entities.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, req.SubscriptionID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Verify ownership
	if subscription.UserID != userID {
		tx.Rollback()
		return nil, fmt.Errorf("subscription does not belong to user")
	}

	// Verify subscription is active
	if !subscription.IsActive() {
		tx.Rollback()
		return nil, fmt.Errorf("can only update server groups for active subscriptions")
	}

	// Get subscription plan to validate server group access
	plan, err := s.subscriptionPlanService.GetSubscriptionPlan(ctx, subscription.SubscriptionPlanID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Validate requested server groups
	if len(req.ServerGroupIDs) > 0 {
		// Check if all requested groups exist
		var existingGroups []serverEntities.ServerGroup
		if err := tx.Where("id IN (?)", req.ServerGroupIDs).
			Find(&existingGroups).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to validate server groups: %w", err)
		}

		if len(existingGroups) != len(req.ServerGroupIDs) {
			tx.Rollback()
			return nil, fmt.Errorf("some server groups are invalid")
		}

		// Note: Plan server group restrictions will be implemented later
		_ = plan // Use plan variable
	}

	// Store old server groups for change log
	oldGroupIDs := subscription.GetServerGroupIDs()

	// Update server groups
	if err := subscription.SetServerGroupIDs(req.ServerGroupIDs); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set server group IDs: %w", err)
	}

	// Update subscription in database
	if err := tx.Model(&subscription).Update("server_group_ids", subscription.ServerGroupIDs).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update subscription server groups",
			logger.Uint("subscription_id", req.SubscriptionID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to update subscription server groups: %w", err)
	}

	// Log server group update
	logger.Info("Server groups updated by user",
		logger.Uint("subscription_id", req.SubscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Int("old_count", len(oldGroupIDs)),
		logger.Int("new_count", len(req.ServerGroupIDs)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit server group update: %w", err)
	}

	logger.Info("Subscription server groups updated successfully",
		logger.Uint("subscription_id", req.SubscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Any("old_groups", oldGroupIDs),
		logger.Any("new_groups", req.ServerGroupIDs))

	return &subscription, nil
}

// GetUserAccessibleServers returns all shadowsocks servers accessible to a user based on their subscriptions
func (s *UserSubscriptionService) GetUserAccessibleServers(ctx context.Context, userID uint) ([]*serverEntities.ShadowsocksServer, error) {
	// Get all active subscriptions for the user
	activeSubscriptions, _, err := s.GetUserSubscriptions(ctx, &GetUserSubscriptionsRequest{
		UserID: userID,
		Status: entities.UserSubscriptionStatusActive,
		Limit:  1000, // Get all active subscriptions
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if len(activeSubscriptions) == 0 {
		return []*serverEntities.ShadowsocksServer{}, nil
	}

	// Collect all server group IDs from active subscriptions
	var allGroupIDs []uint
	groupIDMap := make(map[uint]bool)

	for _, subscription := range activeSubscriptions {
		groupIDs := subscription.GetServerGroupIDs()
		for _, groupID := range groupIDs {
			if !groupIDMap[groupID] {
				allGroupIDs = append(allGroupIDs, groupID)
				groupIDMap[groupID] = true
			}
		}
	}

	if len(allGroupIDs) == 0 {
		return []*serverEntities.ShadowsocksServer{}, nil
	}

	// Get all servers from accessible server groups
	var accessibleServers []*serverEntities.ShadowsocksServer
	if err := s.db.WithContext(ctx).
		Preload("ServerGroup").
		Where("group_id IN (?) AND show = ?", allGroupIDs, 1). // show=1 means visible
		Order("group_id, sort").
		Find(&accessibleServers).Error; err != nil {
		return nil, fmt.Errorf("failed to get accessible servers: %w", err)
	}

	return accessibleServers, nil
}

// GetUserServersBySubscription returns servers accessible through a specific subscription
func (s *UserSubscriptionService) GetUserServersBySubscription(ctx context.Context, userID uint, subscriptionID uint) ([]*serverEntities.ShadowsocksServer, error) {
	// Get the subscription to verify ownership
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}

	if !subscription.IsActive() {
		return []*serverEntities.ShadowsocksServer{}, nil
	}

	// Get server group IDs for this subscription
	groupIDs := subscription.GetServerGroupIDs()
	if len(groupIDs) == 0 {
		return []*serverEntities.ShadowsocksServer{}, nil
	}

	// Get servers from assigned server groups
	var servers []*serverEntities.ShadowsocksServer
	if err := s.db.WithContext(ctx).
		Preload("ServerGroup").
		Where("group_id IN (?) AND show = ?", groupIDs, 1). // show=1 means visible
		Order("group_id, sort").
		Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscription servers: %w", err)
	}

	return servers, nil
}

// ValidateUserServerAccess validates if a user has access to a specific server
func (s *UserSubscriptionService) ValidateUserServerAccess(ctx context.Context, userID uint, serverID uint) (bool, *entities.UserSubscription, error) {
	// Get the server to find its server group
	var server serverEntities.ShadowsocksServer
	if err := s.db.WithContext(ctx).First(&server, serverID).Error; err != nil {
		return false, nil, fmt.Errorf("failed to get server: %w", err)
	}

	if !server.IsVisible() {
		return false, nil, nil
	}

	// Get all active subscriptions for the user
	activeSubscriptions, _, err := s.GetUserSubscriptions(ctx, &GetUserSubscriptionsRequest{
		UserID: userID,
		Status: entities.UserSubscriptionStatusActive,
		Limit:  1000,
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	// Check if any subscription has access to this server's group
	for _, subscription := range activeSubscriptions {
		if subscription.HasAccessToServerGroup(server.GroupID) {
			return true, subscription, nil
		}
	}

	return false, nil, nil
}

// GetServerGroupUsageStats returns usage statistics for server groups in user's subscriptions
func (s *UserSubscriptionService) GetServerGroupUsageStats(ctx context.Context, userID uint) (map[uint]*ServerGroupUsageStats, error) {
	// Get all active subscriptions for the user
	activeSubscriptions, _, err := s.GetUserSubscriptions(ctx, &GetUserSubscriptionsRequest{
		UserID: userID,
		Status: entities.UserSubscriptionStatusActive,
		Limit:  1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if len(activeSubscriptions) == 0 {
		return make(map[uint]*ServerGroupUsageStats), nil
	}

	// Collect all server group IDs
	var allGroupIDs []uint
	groupIDMap := make(map[uint]bool)

	for _, subscription := range activeSubscriptions {
		groupIDs := subscription.GetServerGroupIDs()
		for _, groupID := range groupIDs {
			if !groupIDMap[groupID] {
				allGroupIDs = append(allGroupIDs, groupID)
				groupIDMap[groupID] = true
			}
		}
	}

	if len(allGroupIDs) == 0 {
		return make(map[uint]*ServerGroupUsageStats), nil
	}

	// Get usage statistics for each server group
	stats := make(map[uint]*ServerGroupUsageStats)

	for _, groupID := range allGroupIDs {
		// Get servers in this group
		var serverCount int64
		if err := s.db.WithContext(ctx).
			Model(&serverEntities.ShadowsocksServer{}).
			Where("group_id = ? AND show = ?", groupID, 1).
			Count(&serverCount).Error; err != nil {
			logger.Error("Failed to count servers in group", logger.Uint("group_id", groupID), logger.Error2("error", err))
			continue
		}

		// Traffic logging is not implemented yet, using placeholder values
		totalTraffic := int64(0)
		sessionCount := int64(0)

		stats[groupID] = &ServerGroupUsageStats{
			ServerGroupID: groupID,
			ServerCount:   int(serverCount),
			TotalTraffic:  totalTraffic,
			SessionCount:  int(sessionCount),
		}
	}

	return stats, nil
}

// ServerGroupUsageStats represents usage statistics for a server group
type ServerGroupUsageStats struct {
	ServerGroupID uint  `json:"server_group_id"`
	ServerCount   int   `json:"server_count"`
	TotalTraffic  int64 `json:"total_traffic"`
	SessionCount  int   `json:"session_count"`
}
