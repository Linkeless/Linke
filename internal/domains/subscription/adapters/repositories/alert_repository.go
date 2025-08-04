package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
)

// alertRepository implements the AlertRepository interface
type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository instance
func NewAlertRepository(db *gorm.DB) interfaces.AlertRepository {
	return &alertRepository{
		db: db,
	}
}

// Alert Configurations

func (r *alertRepository) CreateAlertConfiguration(ctx context.Context, config *entities.AlertConfiguration) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *alertRepository) GetAlertConfiguration(ctx context.Context, id uint) (*entities.AlertConfiguration, error) {
	var config entities.AlertConfiguration
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *alertRepository) GetAlertConfigurations(ctx context.Context, filter interfaces.AlertConfigurationFilter) ([]*entities.AlertConfiguration, error) {
	query := r.db.WithContext(ctx).Model(&entities.AlertConfiguration{})

	// Apply filters
	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if filter.UsageType != "" {
		query = query.Where("usage_type = ?", filter.UsageType)
	}
	if filter.ThresholdType != "" {
		query = query.Where("threshold_type = ?", filter.ThresholdType)
	}
	if filter.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filter.IsEnabled)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if !filter.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	// Apply ordering
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDirection := "DESC"
	if filter.OrderDirection != "" {
		orderDirection = filter.OrderDirection
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDirection))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var configs []*entities.AlertConfiguration
	err := query.Find(&configs).Error
	return configs, err
}

func (r *alertRepository) UpdateAlertConfiguration(ctx context.Context, config *entities.AlertConfiguration) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *alertRepository) DeleteAlertConfiguration(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.AlertConfiguration{}, id).Error
}

// Usage Alerts

func (r *alertRepository) CreateUsageAlert(ctx context.Context, alert *entities.UsageAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *alertRepository) GetUsageAlert(ctx context.Context, id uint) (*entities.UsageAlert, error) {
	var alert entities.UsageAlert
	err := r.db.WithContext(ctx).First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *alertRepository) GetUsageAlerts(ctx context.Context, filter interfaces.UsageAlertFilter) ([]*entities.UsageAlert, error) {
	query := r.db.WithContext(ctx).Model(&entities.UsageAlert{})

	// Apply filters
	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if filter.AlertConfigurationID != 0 {
		query = query.Where("alert_configuration_id = ?", filter.AlertConfigurationID)
	}
	if filter.UsageType != "" {
		query = query.Where("usage_type = ?", filter.UsageType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.StartTime != nil {
		query = query.Where("fired_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("fired_at <= ?", *filter.EndTime)
	}
	if filter.IsResolved != nil {
		if *filter.IsResolved {
			query = query.Where("resolved_at IS NOT NULL")
		} else {
			query = query.Where("resolved_at IS NULL")
		}
	}
	if filter.IsActive != nil {
		if *filter.IsActive {
			query = query.Where("status = ? AND resolved_at IS NULL", entities.AlertStatusFired)
		} else {
			query = query.Where("status != ? OR resolved_at IS NOT NULL", entities.AlertStatusFired)
		}
	}
	if !filter.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	// Apply ordering
	orderBy := "fired_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDirection := "DESC"
	if filter.OrderDirection != "" {
		orderDirection = filter.OrderDirection
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDirection))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var alerts []*entities.UsageAlert
	err := query.Find(&alerts).Error
	return alerts, err
}

func (r *alertRepository) UpdateUsageAlert(ctx context.Context, alert *entities.UsageAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *alertRepository) DeleteUsageAlert(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.UsageAlert{}, id).Error
}

// Alert Status Operations

func (r *alertRepository) ResolveAlert(ctx context.Context, alertID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      entities.AlertStatusResolved,
			"resolved_at": &now,
			"updated_at":  now,
		}).Error
}

func (r *alertRepository) AcknowledgeAlert(ctx context.Context, alertID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":     entities.AlertStatusAcknowledged,
			"updated_at": now,
		}).Error
}

func (r *alertRepository) SuppressAlert(ctx context.Context, alertID uint, duration time.Duration) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":     entities.AlertStatusSuppressed,
			"updated_at": now,
			// Note: Suppression end time would need to be stored in metadata or separate field
		}).Error
}

// Alert Queries

func (r *alertRepository) GetActiveAlerts(ctx context.Context, subscriptionID uint) ([]*entities.UsageAlert, error) {
	var alerts []*entities.UsageAlert
	err := r.db.WithContext(ctx).
		Where("user_subscription_id = ? AND status = ? AND resolved_at IS NULL", subscriptionID, entities.AlertStatusFired).
		Order("fired_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *alertRepository) GetAlertsForConfiguration(ctx context.Context, configID uint, limit int) ([]*entities.UsageAlert, error) {
	var alerts []*entities.UsageAlert
	query := r.db.WithContext(ctx).
		Where("alert_configuration_id = ?", configID).
		Order("fired_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&alerts).Error
	return alerts, err
}

func (r *alertRepository) GetActiveAlertConfigurations(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.AlertConfiguration, error) {
	query := r.db.WithContext(ctx).
		Where("user_subscription_id = ? AND is_enabled = ? AND deleted_at IS NULL", subscriptionID, true)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	var configs []*entities.AlertConfiguration
	err := query.Order("threshold ASC").Find(&configs).Error
	return configs, err
}

// Bulk Operations

func (r *alertRepository) ResolveAlertsForSubscription(ctx context.Context, subscriptionID uint, usageType string) error {
	now := time.Now()
	query := r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND status = ? AND resolved_at IS NULL", subscriptionID, entities.AlertStatusFired)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	return query.Updates(map[string]interface{}{
		"status":      entities.AlertStatusResolved,
		"resolved_at": &now,
		"updated_at":  now,
	}).Error
}

func (r *alertRepository) CleanupOldAlerts(ctx context.Context, olderThan time.Time) (int64, error) {
	// Delete resolved alerts older than the specified time
	result := r.db.WithContext(ctx).
		Where("status = ? AND resolved_at < ?", entities.AlertStatusResolved, olderThan).
		Delete(&entities.UsageAlert{})
	return result.RowsAffected, result.Error
}

// Additional helper methods for complex queries

// GetAlertsSummary returns a summary of alerts for a subscription
func (r *alertRepository) GetAlertsSummary(ctx context.Context, subscriptionID uint, period time.Duration) (*interfaces.AlertsSummary, error) {
	startTime := time.Now().Add(-period)

	var summary struct {
		TotalAlerts    int64 `gorm:"column:total_alerts"`
		ActiveAlerts   int64 `gorm:"column:active_alerts"`
		ResolvedAlerts int64 `gorm:"column:resolved_alerts"`
		CriticalAlerts int64 `gorm:"column:critical_alerts"`
		HighAlerts     int64 `gorm:"column:high_alerts"`
		MediumAlerts   int64 `gorm:"column:medium_alerts"`
		LowAlerts      int64 `gorm:"column:low_alerts"`
	}

	err := r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND fired_at >= ?", subscriptionID, startTime).
		Select(`
			COUNT(*) as total_alerts,
			SUM(CASE WHEN status = 'fired' AND resolved_at IS NULL THEN 1 ELSE 0 END) as active_alerts,
			SUM(CASE WHEN status = 'resolved' OR resolved_at IS NOT NULL THEN 1 ELSE 0 END) as resolved_alerts,
			SUM(CASE WHEN severity = 'critical' THEN 1 ELSE 0 END) as critical_alerts,
			SUM(CASE WHEN severity = 'error' THEN 1 ELSE 0 END) as high_alerts,
			SUM(CASE WHEN severity = 'warning' THEN 1 ELSE 0 END) as medium_alerts,
			SUM(CASE WHEN severity = 'info' THEN 1 ELSE 0 END) as low_alerts
		`).Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	// Get additional time-based counts
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var alertsToday, alertsThisWeek, alertsThisMonth int64

	// Count alerts today
	r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND fired_at >= ?", subscriptionID, todayStart).
		Count(&alertsToday)

	// Count alerts this week
	r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND fired_at >= ?", subscriptionID, weekStart).
		Count(&alertsThisWeek)

	// Count alerts this month
	r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND fired_at >= ?", subscriptionID, monthStart).
		Count(&alertsThisMonth)

	return &interfaces.AlertsSummary{
		TotalAlerts:     summary.TotalAlerts,
		ActiveAlerts:    summary.ActiveAlerts,
		ResolvedAlerts:  summary.ResolvedAlerts,
		CriticalAlerts:  summary.CriticalAlerts,
		HighAlerts:      summary.HighAlerts,
		MediumAlerts:    summary.MediumAlerts,
		LowAlerts:       summary.LowAlerts,
		AlertsToday:     alertsToday,
		AlertsThisWeek:  alertsThisWeek,
		AlertsThisMonth: alertsThisMonth,
	}, nil
}

// GetAlertTrends returns alert trends over time
func (r *alertRepository) GetAlertTrends(ctx context.Context, subscriptionID uint, period time.Duration, granularity string) ([]*interfaces.AlertTrendEntry, error) {
	startTime := time.Now().Add(-period)

	var dateFormat string
	switch granularity {
	case "hourly":
		dateFormat = "%Y-%m-%d %H:00:00"
	case "daily":
		dateFormat = "%Y-%m-%d"
	case "weekly":
		dateFormat = "%Y-%u" // Year-Week
	case "monthly":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d" // Default to daily
	}

	var results []struct {
		Period     string `gorm:"column:period"`
		AlertCount int64  `gorm:"column:alert_count"`
		Severity   string `gorm:"column:severity"`
	}

	err := r.db.WithContext(ctx).Model(&entities.UsageAlert{}).
		Where("user_subscription_id = ? AND fired_at >= ?", subscriptionID, startTime).
		Select(fmt.Sprintf("DATE_FORMAT(fired_at, '%s') as period, COUNT(*) as alert_count, severity", dateFormat)).
		Group("period, severity").
		Order("period ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert results to AlertTrendEntry
	trends := make([]*interfaces.AlertTrendEntry, len(results))
	for i, result := range results {
		timestamp, _ := time.Parse("2006-01-02", result.Period) // Basic parsing, should be improved
		trends[i] = &interfaces.AlertTrendEntry{
			Timestamp:  timestamp,
			AlertCount: result.AlertCount,
			Period:     result.Period,
			Severity:   result.Severity,
		}
	}

	return trends, nil
}

// GetTopAlertConfigurations returns configurations that generate the most alerts
func (r *alertRepository) GetTopAlertConfigurations(ctx context.Context, subscriptionID uint, limit int, period time.Duration) ([]*interfaces.TopAlertConfiguration, error) {
	startTime := time.Now().Add(-period)

	var results []struct {
		ConfigurationID      uint      `gorm:"column:configuration_id"`
		Name                 string    `gorm:"column:name"`
		UsageType            string    `gorm:"column:usage_type"`
		Threshold            float64   `gorm:"column:threshold"`
		AlertCount           int64     `gorm:"column:alert_count"`
		LastAlertFired       time.Time `gorm:"column:last_alert_fired"`
		AvgResolutionSeconds float64   `gorm:"column:avg_resolution_seconds"`
	}

	err := r.db.WithContext(ctx).
		Table("usage_alerts ua").
		Select(`
			ua.alert_configuration_id as configuration_id,
			ac.name,
			ua.usage_type,
			ac.threshold,
			COUNT(*) as alert_count,
			MAX(ua.fired_at) as last_alert_fired,
			AVG(TIMESTAMPDIFF(SECOND, ua.fired_at, ua.resolved_at)) as avg_resolution_seconds
		`).
		Joins("JOIN alert_configurations ac ON ua.alert_configuration_id = ac.id").
		Where("ua.user_subscription_id = ? AND ua.fired_at >= ?", subscriptionID, startTime).
		Group("ua.alert_configuration_id, ac.name, ua.usage_type, ac.threshold").
		Order("alert_count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	topConfigs := make([]*interfaces.TopAlertConfiguration, len(results))
	for i, result := range results {
		avgResolution := time.Duration(result.AvgResolutionSeconds) * time.Second
		topConfigs[i] = &interfaces.TopAlertConfiguration{
			ConfigurationID:       result.ConfigurationID,
			Name:                  result.Name,
			UsageType:             result.UsageType,
			Threshold:             result.Threshold,
			AlertCount:            result.AlertCount,
			LastAlertFired:        result.LastAlertFired,
			AverageResolutionTime: avgResolution.String(),
		}
	}

	return topConfigs, nil
}
