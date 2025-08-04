package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
)

// usageRepository implements the UsageRepository interface
type usageRepository struct {
	db *gorm.DB
}

// NewUsageRepository creates a new usage repository instance
func NewUsageRepository(db *gorm.DB) interfaces.UsageRepository {
	return &usageRepository{
		db: db,
	}
}

// Usage Records Management

func (r *usageRepository) CreateUsageRecord(ctx context.Context, record *entities.UsageRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *usageRepository) GetUsageRecord(ctx context.Context, id uint) (*entities.UsageRecord, error) {
	var record entities.UsageRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *usageRepository) GetUsageRecords(ctx context.Context, filter interfaces.UsageRecordFilter) ([]*entities.UsageRecord, error) {
	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{})

	// Apply filters
	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if filter.UsageType != "" {
		query = query.Where("usage_type = ?", filter.UsageType)
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	if filter.SourceID != "" {
		query = query.Where("source_id = ?", filter.SourceID)
	}
	if filter.IPAddress != "" {
		query = query.Where("ip_address = ?", filter.IPAddress)
	}
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}
	if !filter.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	// Apply ordering
	orderBy := "timestamp"
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

	var records []*entities.UsageRecord
	err := query.Find(&records).Error
	return records, err
}

func (r *usageRepository) UpdateUsageRecord(ctx context.Context, record *entities.UsageRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *usageRepository) DeleteUsageRecord(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.UsageRecord{}, id).Error
}

// Usage Aggregations

func (r *usageRepository) GetUsageAggregation(ctx context.Context, filter interfaces.UsageAggregationFilter) (*entities.UsageAggregation, error) {
	aggregations, err := r.GetUsageAggregations(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(aggregations) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return aggregations[0], nil
}

func (r *usageRepository) GetUsageAggregations(ctx context.Context, filter interfaces.UsageAggregationFilter) ([]*entities.UsageAggregation, error) {
	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{})

	// Apply filters
	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if filter.UsageType != "" {
		query = query.Where("usage_type = ?", filter.UsageType)
	}
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	// Build SELECT clause based on period
	selectClause := "user_subscription_id, usage_type"
	groupByClause := "user_subscription_id, usage_type"

	switch filter.Period {
	case interfaces.PeriodHourly:
		selectClause += ", DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00') as period_start"
		groupByClause += ", DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00')"
	case interfaces.PeriodDaily:
		selectClause += ", DATE(timestamp) as period_start"
		groupByClause += ", DATE(timestamp)"
	case interfaces.PeriodWeekly:
		selectClause += ", DATE_SUB(DATE(timestamp), INTERVAL WEEKDAY(timestamp) DAY) as period_start"
		groupByClause += ", DATE_SUB(DATE(timestamp), INTERVAL WEEKDAY(timestamp) DAY)"
	case interfaces.PeriodMonthly:
		selectClause += ", DATE_FORMAT(timestamp, '%Y-%m-01') as period_start"
		groupByClause += ", DATE_FORMAT(timestamp, '%Y-%m-01')"
	default:
		// Default to daily
		selectClause += ", DATE(timestamp) as period_start"
		groupByClause += ", DATE(timestamp)"
	}

	selectClause += `
		, SUM(amount) as total_amount
		, AVG(amount) as average_amount
		, MIN(amount) as min_amount
		, MAX(amount) as max_amount
		, COUNT(*) as record_count
		, MIN(timestamp) as first_usage
		, MAX(timestamp) as last_usage
	`

	query = query.Select(selectClause).Group(groupByClause)

	// Apply additional group by
	for _, groupBy := range filter.GroupBy {
		if groupBy != "usage_type" { // Already included
			query = query.Group(groupBy)
		}
	}

	// Apply ordering
	orderBy := "period_start"
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

	var results []struct {
		UserSubscriptionID uint      `gorm:"column:user_subscription_id"`
		UsageType          string    `gorm:"column:usage_type"`
		PeriodStart        time.Time `gorm:"column:period_start"`
		TotalAmount        int64     `gorm:"column:total_amount"`
		AverageAmount      float64   `gorm:"column:average_amount"`
		MinAmount          int64     `gorm:"column:min_amount"`
		MaxAmount          int64     `gorm:"column:max_amount"`
		RecordCount        int64     `gorm:"column:record_count"`
		FirstUsage         time.Time `gorm:"column:first_usage"`
		LastUsage          time.Time `gorm:"column:last_usage"`
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert results to UsageAggregation entities
	aggregations := make([]*entities.UsageAggregation, len(results))
	for i, result := range results {
		periodEnd := result.PeriodStart
		switch filter.Period {
		case interfaces.PeriodHourly:
			periodEnd = result.PeriodStart.Add(time.Hour)
		case interfaces.PeriodDaily:
			periodEnd = result.PeriodStart.AddDate(0, 0, 1)
		case interfaces.PeriodWeekly:
			periodEnd = result.PeriodStart.AddDate(0, 0, 7)
		case interfaces.PeriodMonthly:
			periodEnd = result.PeriodStart.AddDate(0, 1, 0)
		}

		aggregations[i] = &entities.UsageAggregation{
			UserSubscriptionID: result.UserSubscriptionID,
			UsageType:          result.UsageType,
			Period:             filter.Period,
			PeriodStart:        result.PeriodStart,
			PeriodEnd:          periodEnd,
			TotalAmount:        result.TotalAmount,
			AverageAmount:      result.AverageAmount,
			MinAmount:          result.MinAmount,
			MaxAmount:          result.MaxAmount,
			RecordCount:        result.RecordCount,
			FirstUsage:         result.FirstUsage,
			LastUsage:          result.LastUsage,
		}
	}

	return aggregations, nil
}

func (r *usageRepository) GetUsageSummary(ctx context.Context, filter interfaces.UsageSummaryFilter) (*entities.UsageSummary, error) {
	// Determine period boundaries
	var periodStart, periodEnd time.Time
	now := time.Now()

	if filter.PeriodStart != nil && filter.PeriodEnd != nil {
		periodStart = *filter.PeriodStart
		periodEnd = *filter.PeriodEnd
	} else {
		switch filter.Period {
		case interfaces.PeriodDaily:
			periodStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			periodEnd = periodStart.AddDate(0, 0, 1)
		case interfaces.PeriodWeekly:
			weekday := int(now.Weekday())
			periodStart = now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
			periodEnd = periodStart.AddDate(0, 0, 7)
		case interfaces.PeriodMonthly:
			periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			periodEnd = periodStart.AddDate(0, 1, 0)
		default:
			// Default to current day
			periodStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			periodEnd = periodStart.AddDate(0, 0, 1)
		}
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{})

	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if len(filter.UsageTypes) > 0 {
		query = query.Where("usage_type IN ?", filter.UsageTypes)
	}
	query = query.Where("timestamp >= ? AND timestamp < ?", periodStart, periodEnd)

	// Get aggregated data by usage type
	var results []struct {
		UsageType     string    `gorm:"column:usage_type"`
		TotalAmount   int64     `gorm:"column:total_amount"`
		AverageAmount float64   `gorm:"column:average_amount"`
		MinAmount     int64     `gorm:"column:min_amount"`
		MaxAmount     int64     `gorm:"column:max_amount"`
		RecordCount   int64     `gorm:"column:record_count"`
		FirstUsage    time.Time `gorm:"column:first_usage"`
		LastUsage     time.Time `gorm:"column:last_usage"`
		Unit          string    `gorm:"column:unit"`
	}

	err := query.Select(`
		usage_type,
		SUM(amount) as total_amount,
		AVG(amount) as average_amount,
		MIN(amount) as min_amount,
		MAX(amount) as max_amount,
		COUNT(*) as record_count,
		MIN(timestamp) as first_usage,
		MAX(timestamp) as last_usage,
		unit
	`).Group("usage_type, unit").Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Build usage by type map
	usageByType := make(map[string]*entities.UsageTypeStats)
	var totalUsage, totalRecords int64

	for _, result := range results {
		usageByType[result.UsageType] = &entities.UsageTypeStats{
			UsageType:     result.UsageType,
			TotalAmount:   result.TotalAmount,
			AverageAmount: result.AverageAmount,
			MinAmount:     result.MinAmount,
			MaxAmount:     result.MaxAmount,
			RecordCount:   result.RecordCount,
			FirstUsage:    result.FirstUsage,
			LastUsage:     result.LastUsage,
			Unit:          result.Unit,
		}
		totalUsage += result.TotalAmount
		totalRecords += result.RecordCount
	}

	summary := &entities.UsageSummary{
		UserSubscriptionID: filter.UserSubscriptionID,
		Period:             filter.Period,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		UsageByType:        usageByType,
		TotalUsage:         totalUsage,
		TotalRecords:       totalRecords,
		GeneratedAt:        time.Now(),
	}

	return summary, nil
}

// Bulk Operations

func (r *usageRepository) CreateUsageRecordsBatch(ctx context.Context, records []*entities.UsageRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Use batch size to avoid overwhelming the database
	batchSize := 1000
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		if err := r.db.WithContext(ctx).CreateInBatches(batch, len(batch)).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *usageRepository) DeleteOldUsageRecords(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", olderThan).Delete(&entities.UsageRecord{})
	return result.RowsAffected, result.Error
}

// Current Usage Queries

func (r *usageRepository) GetCurrentUsage(ctx context.Context, subscriptionID uint, usageType string) (int64, error) {
	var totalUsage int64

	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{}).
		Where("user_subscription_id = ?", subscriptionID)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	err := query.Select("COALESCE(SUM(amount), 0)").Scan(&totalUsage).Error
	return totalUsage, err
}

func (r *usageRepository) GetCurrentUsageByPeriod(ctx context.Context, subscriptionID uint, usageType string, periodStart, periodEnd time.Time) (int64, error) {
	var totalUsage int64

	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{}).
		Where("user_subscription_id = ? AND timestamp >= ? AND timestamp < ?", subscriptionID, periodStart, periodEnd)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	err := query.Select("COALESCE(SUM(amount), 0)").Scan(&totalUsage).Error
	return totalUsage, err
}

func (r *usageRepository) GetUsageByTimeRange(ctx context.Context, subscriptionID uint, usageType string, startTime, endTime time.Time) ([]*entities.UsageRecord, error) {
	var records []*entities.UsageRecord

	query := r.db.WithContext(ctx).Where("user_subscription_id = ? AND timestamp >= ? AND timestamp <= ?", subscriptionID, startTime, endTime)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	err := query.Order("timestamp ASC").Find(&records).Error
	return records, err
}

// Statistics

func (r *usageRepository) GetUsageStatistics(ctx context.Context, filter interfaces.UsageStatsFilter) (*interfaces.UsageStatistics, error) {
	// This is a complex aggregation query that would need to be implemented
	// based on specific requirements. For now, return a basic implementation.

	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{})

	if filter.UserSubscriptionID != 0 {
		query = query.Where("user_subscription_id = ?", filter.UserSubscriptionID)
	}
	if filter.UsageType != "" {
		query = query.Where("usage_type = ?", filter.UsageType)
	}
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	var result struct {
		TotalUsage    int64     `gorm:"column:total_usage"`
		TotalRecords  int64     `gorm:"column:total_records"`
		AverageUsage  float64   `gorm:"column:average_usage"`
		PeakUsage     int64     `gorm:"column:peak_usage"`
		PeakUsageTime time.Time `gorm:"column:peak_usage_time"`
		StartTime     time.Time `gorm:"column:start_time"`
		EndTime       time.Time `gorm:"column:end_time"`
	}

	err := query.Select(`
		COALESCE(SUM(amount), 0) as total_usage,
		COUNT(*) as total_records,
		COALESCE(AVG(amount), 0) as average_usage,
		COALESCE(MAX(amount), 0) as peak_usage,
		(SELECT timestamp FROM usage_records ur2 WHERE ur2.amount = MAX(usage_records.amount) LIMIT 1) as peak_usage_time,
		MIN(timestamp) as start_time,
		MAX(timestamp) as end_time
	`).Scan(&result).Error

	if err != nil {
		return nil, err
	}

	stats := &interfaces.UsageStatistics{
		UserSubscriptionID: filter.UserSubscriptionID,
		Period:             filter.Period,
		TotalUsage:         result.TotalUsage,
		TotalRecords:       result.TotalRecords,
		AverageUsage:       result.AverageUsage,
		PeakUsage:          result.PeakUsage,
		PeakUsageTime:      result.PeakUsageTime,
		UsageByType:        make(map[string]*entities.UsageTypeStats),
		UsageByHour:        make(map[int]int64),
		UsageByDayOfWeek:   make(map[int]int64),
		UsageBySource:      make(map[string]int64),
		StartTime:          result.StartTime,
		EndTime:            result.EndTime,
		GeneratedAt:        time.Now(),
	}

	return stats, nil
}

func (r *usageRepository) GetTopUsageSubscriptions(ctx context.Context, usageType string, limit int, period time.Duration) ([]*interfaces.TopUsageSubscription, error) {
	startTime := time.Now().Add(-period)

	query := r.db.WithContext(ctx).Model(&entities.UsageRecord{}).
		Where("timestamp >= ?", startTime)

	if usageType != "" {
		query = query.Where("usage_type = ?", usageType)
	}

	var results []struct {
		UserSubscriptionID uint    `gorm:"column:user_subscription_id"`
		TotalUsage         int64   `gorm:"column:total_usage"`
		RecordCount        int64   `gorm:"column:record_count"`
		AverageUsage       float64 `gorm:"column:average_usage"`
		PeakUsage          int64   `gorm:"column:peak_usage"`
	}

	err := query.Select(`
		user_subscription_id,
		SUM(amount) as total_usage,
		COUNT(*) as record_count,
		AVG(amount) as average_usage,
		MAX(amount) as peak_usage
	`).Group("user_subscription_id").
		Order("total_usage DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	topUsage := make([]*interfaces.TopUsageSubscription, len(results))
	for i, result := range results {
		topUsage[i] = &interfaces.TopUsageSubscription{
			UserSubscriptionID: result.UserSubscriptionID,
			TotalUsage:         result.TotalUsage,
			UsageType:          usageType,
			Period:             fmt.Sprintf("%.0fh", period.Hours()),
			PeriodStart:        startTime,
			PeriodEnd:          time.Now(),
			RecordCount:        result.RecordCount,
			AverageUsage:       result.AverageUsage,
			PeakUsage:          result.PeakUsage,
		}
	}

	return topUsage, nil
}
