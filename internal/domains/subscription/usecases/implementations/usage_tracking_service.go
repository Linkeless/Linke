package implementations

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
)

// usageTrackingService implements the UsageTrackingService interface
type usageTrackingService struct {
	usageRepo       interfaces.UsageRepository
	alertRepo       interfaces.AlertRepository
	subscriptionSvc subscriptionInterfaces.UserSubscriptionService
	alertSvc        interfaces.UsageAlertService
}

// NewUsageTrackingService creates a new usage tracking service instance
func NewUsageTrackingService(
	usageRepo interfaces.UsageRepository,
	alertRepo interfaces.AlertRepository,
	subscriptionSvc subscriptionInterfaces.UserSubscriptionService,
	alertSvc interfaces.UsageAlertService,
) interfaces.UsageTrackingService {
	return &usageTrackingService{
		usageRepo:       usageRepo,
		alertRepo:       alertRepo,
		subscriptionSvc: subscriptionSvc,
		alertSvc:        alertSvc,
	}
}

// Real-time Usage Tracking

func (s *usageTrackingService) RecordUsage(ctx context.Context, req *interfaces.RecordUsageRequest) error {
	// Validate request
	if req.UserSubscriptionID == 0 {
		return fmt.Errorf("user_subscription_id is required")
	}
	if req.UsageType == "" {
		return fmt.Errorf("usage_type is required")
	}
	if req.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	// Set defaults
	if req.Unit == "" {
		req.Unit = entities.UnitBytes
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	if req.SourceType == "" {
		req.SourceType = entities.SourceTypeSystem
	}

	// Create usage record
	record := &entities.UsageRecord{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		Amount:             req.Amount,
		Unit:               req.Unit,
		Timestamp:          req.Timestamp,
		SourceType:         req.SourceType,
		SourceID:           req.SourceID,
		Metadata:           req.Metadata,
		UserAgent:          req.UserAgent,
		IPAddress:          req.IPAddress,
	}

	// Save usage record
	if err := s.usageRepo.CreateUsageRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	// Update subscription usage and check alerts asynchronously to avoid blocking
	go func() {
		// Use a new context with timeout for async operations
		asyncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Update subscription usage totals
		if err := s.UpdateSubscriptionUsage(asyncCtx, req.UserSubscriptionID); err != nil {
			// Log error but don't fail the main operation
			// In production, this should use proper logging
			fmt.Printf("Failed to update subscription usage: %v\n", err)
		}

		// Process usage alerts
		if err := s.alertSvc.ProcessUsageUpdate(asyncCtx, req.UserSubscriptionID, req.UsageType, req.Amount); err != nil {
			// Log error but don't fail the main operation
			fmt.Printf("Failed to process usage alerts: %v\n", err)
		}
	}()

	return nil
}

func (s *usageTrackingService) RecordUsageBatch(ctx context.Context, requests []*interfaces.RecordUsageRequest) error {
	if len(requests) == 0 {
		return nil
	}

	// Validate all requests first
	for i, req := range requests {
		if req.UserSubscriptionID == 0 {
			return fmt.Errorf("request %d: user_subscription_id is required", i)
		}
		if req.UsageType == "" {
			return fmt.Errorf("request %d: usage_type is required", i)
		}
		if req.Amount < 0 {
			return fmt.Errorf("request %d: amount cannot be negative", i)
		}
	}

	// Convert requests to usage records
	records := make([]*entities.UsageRecord, len(requests))
	subscriptionUpdates := make(map[uint]bool) // Track which subscriptions need updates

	for i, req := range requests {
		// Set defaults
		if req.Unit == "" {
			req.Unit = entities.UnitBytes
		}
		if req.Timestamp.IsZero() {
			req.Timestamp = time.Now()
		}
		if req.SourceType == "" {
			req.SourceType = entities.SourceTypeSystem
		}

		records[i] = &entities.UsageRecord{
			UserSubscriptionID: req.UserSubscriptionID,
			UsageType:          req.UsageType,
			Amount:             req.Amount,
			Unit:               req.Unit,
			Timestamp:          req.Timestamp,
			SourceType:         req.SourceType,
			SourceID:           req.SourceID,
			Metadata:           req.Metadata,
			UserAgent:          req.UserAgent,
			IPAddress:          req.IPAddress,
		}

		subscriptionUpdates[req.UserSubscriptionID] = true
	}

	// Save all records in batch
	if err := s.usageRepo.CreateUsageRecordsBatch(ctx, records); err != nil {
		return fmt.Errorf("failed to create usage records batch: %w", err)
	}

	// Update subscriptions and process alerts asynchronously
	go func() {
		asyncCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		for subscriptionID := range subscriptionUpdates {
			// Update subscription usage
			if err := s.UpdateSubscriptionUsage(asyncCtx, subscriptionID); err != nil {
				fmt.Printf("Failed to update subscription %d usage: %v\n", subscriptionID, err)
			}

			// Get unique usage types for this subscription in the batch
			usageTypes := make(map[string]bool)
			for _, req := range requests {
				if req.UserSubscriptionID == subscriptionID {
					usageTypes[req.UsageType] = true
				}
			}

			// Process alerts for each usage type
			for usageType := range usageTypes {
				currentUsage, err := s.usageRepo.GetCurrentUsage(asyncCtx, subscriptionID, usageType)
				if err != nil {
					fmt.Printf("Failed to get current usage for subscription %d, type %s: %v\n", subscriptionID, usageType, err)
					continue
				}

				if err := s.alertSvc.ProcessUsageUpdate(asyncCtx, subscriptionID, usageType, currentUsage); err != nil {
					fmt.Printf("Failed to process alerts for subscription %d, type %s: %v\n", subscriptionID, usageType, err)
				}
			}
		}
	}()

	return nil
}

func (s *usageTrackingService) GetCurrentUsage(ctx context.Context, subscriptionID uint, usageType string) (*interfaces.CurrentUsageResponse, error) {
	// Get subscription to determine limits and period
	subscription, err := s.subscriptionSvc.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Determine current period based on subscription's traffic reset cycle
	now := time.Now()
	var periodStart, periodEnd time.Time

	if subscription.TrafficResetDate != nil {
		// Use the subscription's specific reset date as period boundary
		resetDate := *subscription.TrafficResetDate
		if resetDate.After(now) {
			// If reset date is in the future, find the previous reset
			switch subscription.TrafficResetCycle {
			case entities.TrafficResetCycleMonthly:
				periodStart = resetDate.AddDate(0, -1, 0)
			default:
				periodStart = resetDate.AddDate(0, -1, 0) // Default to monthly
			}
		} else {
			periodStart = resetDate
		}

		switch subscription.TrafficResetCycle {
		case entities.TrafficResetCycleMonthly:
			periodEnd = periodStart.AddDate(0, 1, 0)
		default:
			periodEnd = periodStart.AddDate(0, 1, 0) // Default to monthly
		}
	} else {
		// Default to monthly periods starting from subscription start
		startDate := subscription.StartDate
		if subscription.CurrentPeriodStart != nil {
			startDate = *subscription.CurrentPeriodStart
		}

		// Find current period
		monthsSinceStart := int(now.Sub(startDate).Hours() / (24 * 30)) // Approximate
		periodStart = startDate.AddDate(0, monthsSinceStart, 0)
		periodEnd = periodStart.AddDate(0, 1, 0)
	}

	// Get usage data
	usageByType := make(map[string]*interfaces.CurrentUsageDetails)

	if usageType != "" {
		// Get usage for specific type
		currentUsage, err := s.usageRepo.GetCurrentUsageByPeriod(ctx, subscriptionID, usageType, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get current usage: %w", err)
		}

		// Get usage limit based on type
		var usageLimit int64
		var unit string = entities.UnitBytes

		switch usageType {
		case entities.UsageTypeTraffic:
			usageLimit = subscription.TrafficLimit
			unit = entities.UnitBytes
		default:
			// For other usage types, we'd need to get limits from subscription plan
			usageLimit = 0 // Unlimited for now
		}

		// Calculate percentage and remaining
		var usagePercent float64
		var remainingUsage int64
		var isUnlimited, isExceeded bool

		if usageLimit == 0 {
			isUnlimited = true
			usagePercent = 0
			remainingUsage = -1 // Indicates unlimited
		} else {
			usagePercent = float64(currentUsage) / float64(usageLimit) * 100
			if usagePercent > 100 {
				usagePercent = 100
				isExceeded = true
			}
			remainingUsage = usageLimit - currentUsage
			if remainingUsage < 0 {
				remainingUsage = 0
			}
		}

		// Get record count and last recorded time
		records, err := s.usageRepo.GetUsageByTimeRange(ctx, subscriptionID, usageType, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get usage records: %w", err)
		}

		var lastRecorded time.Time
		if len(records) > 0 {
			// Records should be ordered by timestamp ASC
			lastRecorded = records[len(records)-1].Timestamp
		}

		usageByType[usageType] = &interfaces.CurrentUsageDetails{
			UsageType:      usageType,
			CurrentUsage:   currentUsage,
			UsageLimit:     usageLimit,
			UsagePercent:   usagePercent,
			RemainingUsage: remainingUsage,
			Unit:           unit,
			LastRecorded:   lastRecorded,
			RecordCount:    int64(len(records)),
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			IsUnlimited:    isUnlimited,
			IsExceeded:     isExceeded,
		}
	} else {
		// Get usage for all types
		allUsageTypes := []string{entities.UsageTypeTraffic, entities.UsageTypeAPICall, entities.UsageTypeStorage}

		for _, uType := range allUsageTypes {
			currentUsage, err := s.usageRepo.GetCurrentUsageByPeriod(ctx, subscriptionID, uType, periodStart, periodEnd)
			if err != nil {
				continue // Skip if error getting usage for this type
			}

			// Skip if no usage recorded
			if currentUsage == 0 {
				continue
			}

			// Get usage limit and unit based on type
			var usageLimit int64
			var unit string = entities.UnitBytes

			switch uType {
			case entities.UsageTypeTraffic:
				usageLimit = subscription.TrafficLimit
				unit = entities.UnitBytes
			default:
				usageLimit = 0 // Unlimited for now
			}

			// Calculate metrics
			var usagePercent float64
			var remainingUsage int64
			var isUnlimited, isExceeded bool

			if usageLimit == 0 {
				isUnlimited = true
				usagePercent = 0
				remainingUsage = -1
			} else {
				usagePercent = float64(currentUsage) / float64(usageLimit) * 100
				if usagePercent > 100 {
					usagePercent = 100
					isExceeded = true
				}
				remainingUsage = usageLimit - currentUsage
				if remainingUsage < 0 {
					remainingUsage = 0
				}
			}

			// Get record count and last recorded time
			records, err := s.usageRepo.GetUsageByTimeRange(ctx, subscriptionID, uType, periodStart, periodEnd)
			if err != nil {
				continue
			}

			var lastRecorded time.Time
			if len(records) > 0 {
				lastRecorded = records[len(records)-1].Timestamp
			}

			usageByType[uType] = &interfaces.CurrentUsageDetails{
				UsageType:      uType,
				CurrentUsage:   currentUsage,
				UsageLimit:     usageLimit,
				UsagePercent:   usagePercent,
				RemainingUsage: remainingUsage,
				Unit:           unit,
				LastRecorded:   lastRecorded,
				RecordCount:    int64(len(records)),
				PeriodStart:    periodStart,
				PeriodEnd:      periodEnd,
				IsUnlimited:    isUnlimited,
				IsExceeded:     isExceeded,
			}
		}
	}

	// Calculate total usage
	var totalUsage int64
	var lastUpdated time.Time
	for _, details := range usageByType {
		totalUsage += details.CurrentUsage
		if details.LastRecorded.After(lastUpdated) {
			lastUpdated = details.LastRecorded
		}
	}

	return &interfaces.CurrentUsageResponse{
		UserSubscriptionID: subscriptionID,
		UsageByType:        usageByType,
		TotalUsage:         totalUsage,
		LastUpdated:        lastUpdated,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
	}, nil
}

func (s *usageTrackingService) GetUsageHistory(ctx context.Context, req *interfaces.UsageHistoryRequest) (*interfaces.UsageHistoryResponse, error) {
	// Set defaults
	if req.Granularity == "" {
		req.Granularity = interfaces.GranularityDaily
	}
	if req.Limit == 0 {
		req.Limit = interfaces.DefaultPageSize
	}
	if req.Limit > interfaces.MaxPageSize {
		req.Limit = interfaces.MaxPageSize
	}

	// Determine time range
	var startTime, endTime time.Time
	now := time.Now()

	if req.StartTime != nil {
		startTime = *req.StartTime
	} else {
		// Default to last 30 days
		startTime = now.AddDate(0, 0, -30)
	}

	if req.EndTime != nil {
		endTime = *req.EndTime
	} else {
		endTime = now
	}

	// Get aggregated usage data
	filter := interfaces.UsageAggregationFilter{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		Period:             req.Granularity,
		StartTime:          &startTime,
		EndTime:            &endTime,
		OrderBy:            "period_start",
		OrderDirection:     "ASC",
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	aggregations, err := s.usageRepo.GetUsageAggregations(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage aggregations: %w", err)
	}

	// Convert to history entries
	historyEntries := make([]*interfaces.UsageHistoryEntry, 0)

	// Group aggregations by period
	periodMap := make(map[string]*interfaces.UsageHistoryEntry)

	for _, agg := range aggregations {
		periodKey := agg.PeriodStart.Format("2006-01-02")
		if req.Granularity == interfaces.GranularityHourly {
			periodKey = agg.PeriodStart.Format("2006-01-02 15:00")
		}

		entry, exists := periodMap[periodKey]
		if !exists {
			entry = &interfaces.UsageHistoryEntry{
				Period:      periodKey,
				PeriodStart: agg.PeriodStart,
				PeriodEnd:   agg.PeriodEnd,
				UsageByType: make(map[string]*entities.UsageTypeStats),
				TotalUsage:  0,
				RecordCount: 0,
			}
			periodMap[periodKey] = entry
		}

		// Add usage type data
		entry.UsageByType[agg.UsageType] = &entities.UsageTypeStats{
			UsageType:     agg.UsageType,
			TotalAmount:   agg.TotalAmount,
			AverageAmount: agg.AverageAmount,
			MinAmount:     agg.MinAmount,
			MaxAmount:     agg.MaxAmount,
			RecordCount:   agg.RecordCount,
			FirstUsage:    agg.FirstUsage,
			LastUsage:     agg.LastUsage,
			Unit:          entities.UnitBytes, // Default unit
		}

		entry.TotalUsage += agg.TotalAmount
		entry.RecordCount += agg.RecordCount
	}

	// Convert map to slice and sort
	for _, entry := range periodMap {
		historyEntries = append(historyEntries, entry)
	}

	sort.Slice(historyEntries, func(i, j int) bool {
		return historyEntries[i].PeriodStart.Before(historyEntries[j].PeriodStart)
	})

	// Get summary if requested
	var summary *entities.UsageSummary
	if req.IncludeDetails {
		summaryFilter := interfaces.UsageSummaryFilter{
			UserSubscriptionID: req.UserSubscriptionID,
			Period:             interfaces.PeriodCustom,
			PeriodStart:        &startTime,
			PeriodEnd:          &endTime,
			IncludeBreakdown:   true,
		}
		summary, _ = s.usageRepo.GetUsageSummary(ctx, summaryFilter)
	}

	// Calculate pagination info
	totalRecords := int64(len(historyEntries))
	totalPages := int(math.Ceil(float64(totalRecords) / float64(req.Limit)))
	currentPage := (req.Offset / req.Limit) + 1

	pagination := &interfaces.PaginationInfo{
		CurrentPage: currentPage,
		PageSize:    req.Limit,
		TotalPages:  totalPages,
		TotalCount:  totalRecords,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}

	return &interfaces.UsageHistoryResponse{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageHistory:       historyEntries,
		Summary:            summary,
		Pagination:         pagination,
		Granularity:        req.Granularity,
		TotalRecords:       totalRecords,
	}, nil
}

// Usage Aggregation and Reporting

func (s *usageTrackingService) GetUsageSummary(ctx context.Context, req *interfaces.UsageSummaryRequest) (*entities.UsageSummary, error) {
	filter := interfaces.UsageSummaryFilter{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageTypes:         req.UsageTypes,
		Period:             req.Period,
		PeriodStart:        req.PeriodStart,
		PeriodEnd:          req.PeriodEnd,
		IncludeBreakdown:   req.IncludeBreakdown,
		IncludePredictions: req.IncludePredictions,
	}

	return s.usageRepo.GetUsageSummary(ctx, filter)
}

func (s *usageTrackingService) GetUsageStatistics(ctx context.Context, req *interfaces.UsageStatsRequest) (*interfaces.UsageStatistics, error) {
	filter := interfaces.UsageStatsFilter{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		Period:             req.Period,
		StartTime:          req.StartTime,
		EndTime:            req.EndTime,
		GroupBy:            req.GroupBy,
	}

	return s.usageRepo.GetUsageStatistics(ctx, filter)
}

func (s *usageTrackingService) GetUsageTrends(ctx context.Context, req *interfaces.UsageTrendsRequest) (*interfaces.UsageTrendsResponse, error) {
	// Parse period to duration
	var period time.Duration
	switch req.Period {
	case "7d":
		period = 7 * 24 * time.Hour
	case "30d":
		period = 30 * 24 * time.Hour
	case "90d":
		period = 90 * 24 * time.Hour
	case "365d":
		period = 365 * 24 * time.Hour
	default:
		period = 30 * 24 * time.Hour // Default to 30 days
	}

	startTime := time.Now().Add(-period)
	endTime := time.Now()

	// Get usage history for trend analysis
	historyReq := &interfaces.UsageHistoryRequest{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		StartTime:          &startTime,
		EndTime:            &endTime,
		Granularity:        req.Granularity,
		Limit:              1000, // Get enough data for trend analysis
		IncludeDetails:     false,
	}

	historyResp, err := s.GetUsageHistory(ctx, historyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage history for trends: %w", err)
	}

	// Convert history to trend entries
	trends := make([]*interfaces.UsageTrendEntry, 0, len(historyResp.UsageHistory))

	for _, entry := range historyResp.UsageHistory {
		// Calculate moving average (simple 3-period moving average)
		movingAvg := float64(entry.TotalUsage)
		// TODO: Implement proper moving average calculation

		// Determine trend direction (simplified)
		trendDirection := interfaces.TrendDirectionStable
		// TODO: Implement trend direction calculation

		trends = append(trends, &interfaces.UsageTrendEntry{
			Timestamp:      entry.PeriodStart,
			UsageAmount:    entry.TotalUsage,
			UsageType:      req.UsageType,
			RecordCount:    entry.RecordCount,
			MovingAvg:      movingAvg,
			TrendDirection: trendDirection,
		})
	}

	// Generate predictions if requested
	var predictions []*entities.UsagePrediction
	if req.IncludePredictions {
		predictionReq := &interfaces.UsagePredictionRequest{
			UserSubscriptionID: req.UserSubscriptionID,
			UsageType:          req.UsageType,
			PredictionType:     interfaces.PredictionTypeDaily,
			BasedOnDays:        30,
			IncludeConfidence:  true,
			IncludeTrends:      true,
		}

		if prediction, err := s.PredictUsage(ctx, predictionReq); err == nil {
			predictions = []*entities.UsagePrediction{prediction}
		}
	}

	// Detect anomalies if requested
	var anomalies []*interfaces.UsageAnomaly
	if req.IncludeAnomalies {
		// TODO: Implement anomaly detection
		anomalies = []*interfaces.UsageAnomaly{}
	}

	// Calculate trend summary
	summary := s.calculateTrendSummary(trends)

	return &interfaces.UsageTrendsResponse{
		UserSubscriptionID: req.UserSubscriptionID,
		Trends:             trends,
		Predictions:        predictions,
		Anomalies:          anomalies,
		Summary:            summary,
		Period:             req.Period,
		Granularity:        req.Granularity,
	}, nil
}

// Usage Predictions

func (s *usageTrackingService) PredictUsage(ctx context.Context, req *interfaces.UsagePredictionRequest) (*entities.UsagePrediction, error) {
	// Get historical data for prediction
	basedOnDays := req.BasedOnDays
	if basedOnDays == 0 {
		basedOnDays = interfaces.DefaultPredictionDays
	}
	if basedOnDays > interfaces.MaxPredictionDays {
		basedOnDays = interfaces.MaxPredictionDays
	}

	startTime := time.Now().AddDate(0, 0, -basedOnDays)
	endTime := time.Now()

	// Get usage records for analysis
	records, err := s.usageRepo.GetUsageByTimeRange(ctx, req.UserSubscriptionID, req.UsageType, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records for prediction: %w", err)
	}

	if len(records) == 0 {
		return &entities.UsagePrediction{
			UserSubscriptionID: req.UserSubscriptionID,
			UsageType:          req.UsageType,
			PredictionType:     req.PredictionType,
			CurrentUsage:       0,
			PredictedUsage:     0,
			Confidence:         0,
			PredictionDate:     time.Now(),
			BasedOnDays:        basedOnDays,
			Trend:              entities.TrendStable,
			EstimatedDaysLeft:  -1,
		}, nil
	}

	// Calculate current usage
	currentUsage, err := s.usageRepo.GetCurrentUsage(ctx, req.UserSubscriptionID, req.UsageType)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// Perform simple linear prediction based on daily averages
	// Group records by day and calculate daily totals
	dailyUsage := make(map[string]int64)
	for _, record := range records {
		day := record.Timestamp.Format("2006-01-02")
		dailyUsage[day] += record.Amount
	}

	// Calculate average daily usage
	totalDays := len(dailyUsage)
	if totalDays == 0 {
		totalDays = 1
	}

	var totalUsage int64
	for _, usage := range dailyUsage {
		totalUsage += usage
	}

	averageDailyUsage := float64(totalUsage) / float64(totalDays)

	// Predict based on type
	var predictedUsage int64
	var confidence float64 = 0.7 // Base confidence

	switch req.PredictionType {
	case entities.PredictionTypeDaily:
		predictedUsage = int64(averageDailyUsage)
	case entities.PredictionTypeWeekly:
		predictedUsage = int64(averageDailyUsage * 7)
	case entities.PredictionTypeMonthly:
		predictedUsage = int64(averageDailyUsage * 30)
	case entities.PredictionTypePeriodEnd:
		// Predict usage until period end (monthly reset)
		// TODO: Get actual period end from subscription
		daysUntilPeriodEnd := 30 // Simplified
		predictedUsage = int64(averageDailyUsage * float64(daysUntilPeriodEnd))
	default:
		predictedUsage = int64(averageDailyUsage)
	}

	// Adjust confidence based on data consistency
	if totalDays >= 7 {
		confidence += 0.1
	}
	if totalDays >= 30 {
		confidence += 0.1
	}

	// Calculate trend
	trend := entities.TrendStable
	if len(records) >= 2 {
		firstHalf := records[:len(records)/2]
		secondHalf := records[len(records)/2:]

		var firstHalfTotal, secondHalfTotal int64
		for _, record := range firstHalf {
			firstHalfTotal += record.Amount
		}
		for _, record := range secondHalf {
			secondHalfTotal += record.Amount
		}

		if secondHalfTotal > firstHalfTotal*110/100 { // 10% increase
			trend = entities.TrendIncreasing
		} else if secondHalfTotal < firstHalfTotal*90/100 { // 10% decrease
			trend = entities.TrendDecreasing
		}
	}

	// Estimate days left until limit
	estimatedDaysLeft := -1 // Unlimited

	// Get subscription to check limits
	subscription, err := s.subscriptionSvc.GetUserSubscription(ctx, req.UserSubscriptionID)
	if err == nil && subscription.TrafficLimit > 0 && req.UsageType == entities.UsageTypeTraffic {
		remainingUsage := subscription.TrafficLimit - currentUsage
		if remainingUsage > 0 && averageDailyUsage > 0 {
			estimatedDaysLeft = int(float64(remainingUsage) / averageDailyUsage)
		}
	}

	return &entities.UsagePrediction{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		PredictionType:     req.PredictionType,
		CurrentUsage:       currentUsage,
		PredictedUsage:     predictedUsage,
		Confidence:         confidence,
		PredictionDate:     time.Now(),
		BasedOnDays:        basedOnDays,
		Trend:              trend,
		EstimatedDaysLeft:  estimatedDaysLeft,
	}, nil
}

func (s *usageTrackingService) GetUsagePredictions(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.UsagePrediction, error) {
	predictions := make([]*entities.UsagePrediction, 0)

	// Generate predictions for different time horizons
	predictionTypes := []string{
		entities.PredictionTypeDaily,
		entities.PredictionTypeWeekly,
		entities.PredictionTypeMonthly,
		entities.PredictionTypePeriodEnd,
	}

	for _, predType := range predictionTypes {
		req := &interfaces.UsagePredictionRequest{
			UserSubscriptionID: subscriptionID,
			UsageType:          usageType,
			PredictionType:     predType,
			BasedOnDays:        30,
			IncludeConfidence:  true,
			IncludeTrends:      true,
		}

		prediction, err := s.PredictUsage(ctx, req)
		if err != nil {
			continue // Skip failed predictions
		}

		predictions = append(predictions, prediction)
	}

	return predictions, nil
}

// Subscription Integration

func (s *usageTrackingService) UpdateSubscriptionUsage(ctx context.Context, subscriptionID uint) error {
	// Get subscription
	subscription, err := s.subscriptionSvc.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Calculate current traffic usage for the period
	now := time.Now()
	var periodStart time.Time

	if subscription.TrafficResetDate != nil {
		resetDate := *subscription.TrafficResetDate
		if resetDate.After(now) {
			// Find previous reset
			switch subscription.TrafficResetCycle {
			case entities.TrafficResetCycleMonthly:
				periodStart = resetDate.AddDate(0, -1, 0)
			default:
				periodStart = resetDate.AddDate(0, -1, 0)
			}
		} else {
			periodStart = resetDate
		}
	} else {
		// Default to monthly from subscription start
		periodStart = subscription.StartDate
		monthsSince := int(now.Sub(periodStart).Hours() / (24 * 30))
		periodStart = periodStart.AddDate(0, monthsSince, 0)
	}

	// Get current traffic usage
	currentTrafficUsage, err := s.usageRepo.GetCurrentUsageByPeriod(ctx, subscriptionID, entities.UsageTypeTraffic, periodStart, now)
	if err != nil {
		return fmt.Errorf("failed to get current traffic usage: %w", err)
	}

	// Update subscription with current usage
	notes := fmt.Sprintf("Traffic usage updated: %d bytes", currentTrafficUsage)
	updateReq := subscriptionInterfaces.UpdateSubscriptionRequest{
		Notes: &notes,
	}

	// Check if traffic limit is exceeded and set status accordingly
	if subscription.TrafficLimit > 0 && currentTrafficUsage >= subscription.TrafficLimit {
		pausedStatus := "paused"
		updateReq.Status = &pausedStatus
	}

	// Use the traffic update method specifically for traffic usage
	if err := s.subscriptionSvc.UpdateTrafficUsage(ctx, subscriptionID, currentTrafficUsage); err != nil {
		return fmt.Errorf("failed to update traffic usage: %w", err)
	}

	// Update other fields if needed
	_, err = s.subscriptionSvc.UpdateUserSubscription(ctx, subscriptionID, &updateReq)
	return err
}

func (s *usageTrackingService) ResetUsageForSubscription(ctx context.Context, subscriptionID uint, usageType string) error {
	// This would typically be called during traffic reset periods
	// For now, we'll just reset the traffic usage to 0
	if usageType == entities.UsageTypeTraffic {
		// Use the reset traffic usage method
		_, err := s.subscriptionSvc.ResetTrafficUsage(ctx, subscriptionID, 0) // 0 for system reset
		return err
	}

	return nil
}

func (s *usageTrackingService) SyncSubscriptionLimits(ctx context.Context, subscriptionID uint) error {
	// Get subscription and its plan to sync limits
	_, err := s.subscriptionSvc.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// TODO: Get subscription plan to sync traffic limits
	// This would require access to subscription plan service
	// For now, just update the current usage
	return s.UpdateSubscriptionUsage(ctx, subscriptionID)
}

// Admin Operations

func (s *usageTrackingService) GetTopUsageSubscriptions(ctx context.Context, req *interfaces.TopUsageRequest) (*interfaces.TopUsageResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	period := req.Period
	if period == 0 {
		period = 24 * time.Hour // Default to last 24 hours
	}

	topUsage, err := s.usageRepo.GetTopUsageSubscriptions(ctx, req.UsageType, limit, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get top usage subscriptions: %w", err)
	}

	return &interfaces.TopUsageResponse{
		TopSubscriptions: topUsage,
		Period:           fmt.Sprintf("%.0fh", period.Hours()),
		GeneratedAt:      time.Now(),
		TotalCount:       int64(len(topUsage)),
	}, nil
}

func (s *usageTrackingService) ExportUsageData(ctx context.Context, req *interfaces.ExportUsageRequest) (*interfaces.ExportUsageResponse, error) {
	// TODO: Implement usage data export
	// This would generate CSV/JSON/XLSX files with usage data
	return &interfaces.ExportUsageResponse{
		DownloadURL: "/api/v1/usage/export/placeholder",
		FileName:    fmt.Sprintf("usage_export_%d.csv", time.Now().Unix()),
		FileSize:    0,
		Format:      req.Format,
		RecordCount: 0,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *usageTrackingService) CleanupOldUsageData(ctx context.Context, olderThan time.Time) (*interfaces.CleanupResult, error) {
	startTime := time.Now()

	// Delete old usage records
	recordsDeleted, err := s.usageRepo.DeleteOldUsageRecords(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old usage records: %w", err)
	}

	// Delete old alerts
	alertsDeleted, err := s.alertRepo.CleanupOldAlerts(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old alerts: %w", err)
	}

	operationTime := time.Since(startTime)

	return &interfaces.CleanupResult{
		RecordsDeleted: recordsDeleted,
		AlertsDeleted:  alertsDeleted,
		SpaceFreed:     recordsDeleted * 100, // Rough estimate of bytes freed
		OperationTime:  operationTime,
		CompletedAt:    time.Now(),
	}, nil
}

// Real-time Monitoring

func (s *usageTrackingService) GetRealTimeUsage(ctx context.Context, subscriptionID uint) (*interfaces.RealTimeUsageResponse, error) {
	// Get current usage for all types
	currentUsageResp, err := s.GetCurrentUsage(ctx, subscriptionID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// Convert to real-time usage data format
	usageByType := make(map[string]*interfaces.RealTimeUsageData)

	for usageType, details := range currentUsageResp.UsageByType {
		// Get recent usage (last hour)
		recentStartTime := time.Now().Add(-time.Hour)
		recentUsage, _ := s.usageRepo.GetCurrentUsageByPeriod(ctx, subscriptionID, usageType, recentStartTime, time.Now())

		// Calculate usage rate (per hour)
		usageRate := float64(recentUsage) // Usage in last hour

		// Estimate time left until limit
		var estimatedTimeLeft time.Duration = -1 // Unlimited
		if details.UsageLimit > 0 && usageRate > 0 {
			remainingUsage := details.RemainingUsage
			if remainingUsage > 0 {
				hoursLeft := float64(remainingUsage) / usageRate
				estimatedTimeLeft = time.Duration(hoursLeft * float64(time.Hour))
			}
		}

		// Determine trend direction (simplified)
		trendDirection := interfaces.TrendDirectionStable
		if usageRate > 0 {
			trendDirection = interfaces.TrendDirectionUp
		}

		// Check if near limit (>80%)
		isNearLimit := details.UsagePercent > 80

		usageByType[usageType] = &interfaces.RealTimeUsageData{
			UsageType:         usageType,
			CurrentUsage:      details.CurrentUsage,
			UsageLimit:        details.UsageLimit,
			UsagePercent:      details.UsagePercent,
			RemainingUsage:    details.RemainingUsage,
			RecentUsage:       recentUsage,
			UsageRate:         usageRate,
			EstimatedTimeLeft: estimatedTimeLeft,
			LastUpdated:       details.LastRecorded,
			TrendDirection:    trendDirection,
			IsNearLimit:       isNearLimit,
			IsExceeded:        details.IsExceeded,
		}
	}

	// Get active alerts count
	activeAlerts, _ := s.alertRepo.GetActiveAlerts(ctx, subscriptionID)
	alertCount := len(activeAlerts)

	// Get predictions
	predictions, _ := s.GetUsagePredictions(ctx, subscriptionID, entities.UsageTypeTraffic)

	return &interfaces.RealTimeUsageResponse{
		UserSubscriptionID: subscriptionID,
		UsageByType:        usageByType,
		LastUpdated:        currentUsageResp.LastUpdated,
		UpdateFrequency:    5 * time.Minute, // Recommend 5-minute updates
		AlertCount:         alertCount,
		Predictions:        predictions,
	}, nil
}

func (s *usageTrackingService) GetUsageAlerts(ctx context.Context, subscriptionID uint) ([]*entities.UsageAlert, error) {
	return s.alertRepo.GetActiveAlerts(ctx, subscriptionID)
}

// Helper methods

func (s *usageTrackingService) calculateTrendSummary(trends []*interfaces.UsageTrendEntry) *interfaces.TrendSummary {
	if len(trends) == 0 {
		return &interfaces.TrendSummary{
			OverallTrend:        interfaces.TrendDirectionStable,
			AverageUsage:        0,
			PeakUsage:           0,
			LowestUsage:         0,
			TrendStrength:       0,
			VariabilityScore:    0,
			AnomalyCount:        0,
			PredictedGrowthRate: 0,
		}
	}

	var totalUsage int64
	var peakUsage, lowestUsage int64
	var peakTime, lowestTime time.Time

	peakUsage = trends[0].UsageAmount
	lowestUsage = trends[0].UsageAmount
	peakTime = trends[0].Timestamp
	lowestTime = trends[0].Timestamp

	for _, trend := range trends {
		totalUsage += trend.UsageAmount

		if trend.UsageAmount > peakUsage {
			peakUsage = trend.UsageAmount
			peakTime = trend.Timestamp
		}

		if trend.UsageAmount < lowestUsage {
			lowestUsage = trend.UsageAmount
			lowestTime = trend.Timestamp
		}
	}

	averageUsage := float64(totalUsage) / float64(len(trends))

	// Simple trend detection
	overallTrend := interfaces.TrendDirectionStable
	if len(trends) >= 2 {
		firstHalf := trends[:len(trends)/2]
		secondHalf := trends[len(trends)/2:]

		var firstHalfAvg, secondHalfAvg float64
		for _, trend := range firstHalf {
			firstHalfAvg += float64(trend.UsageAmount)
		}
		for _, trend := range secondHalf {
			secondHalfAvg += float64(trend.UsageAmount)
		}

		firstHalfAvg /= float64(len(firstHalf))
		secondHalfAvg /= float64(len(secondHalf))

		if secondHalfAvg > firstHalfAvg*1.1 {
			overallTrend = interfaces.TrendDirectionUp
		} else if secondHalfAvg < firstHalfAvg*0.9 {
			overallTrend = interfaces.TrendDirectionDown
		}
	}

	return &interfaces.TrendSummary{
		OverallTrend:        overallTrend,
		AverageUsage:        averageUsage,
		PeakUsage:           peakUsage,
		PeakUsageTime:       peakTime,
		LowestUsage:         lowestUsage,
		LowestUsageTime:     lowestTime,
		TrendStrength:       0.5, // TODO: Calculate proper trend strength
		VariabilityScore:    0.3, // TODO: Calculate proper variability score
		AnomalyCount:        0,   // TODO: Implement anomaly detection
		PredictedGrowthRate: 0,   // TODO: Calculate growth rate
	}
}
