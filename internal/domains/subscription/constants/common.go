package constants

// Granularity Constants
const (
	GranularityHourly  = "hourly"
	GranularityDaily   = "daily"
	GranularityWeekly  = "weekly"
	GranularityMonthly = "monthly"
)

// Period Constants
const (
	PeriodHourly  = "hourly"
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodCustom  = "custom"
)

// Order By Constants for Usage Records
const (
	OrderByTimestamp = "timestamp"
	OrderByAmount    = "amount"
	OrderByCreatedAt = "created_at"
)

// Order By Constants for Alerts
const (
	OrderByFiredAt      = "fired_at"
	OrderByResolvedAt   = "resolved_at"
	OrderBySeverity     = "severity"
	OrderByUsagePercent = "usage_percent"
)

// Order Direction Constants
const (
	OrderDirectionAsc  = "asc"
	OrderDirectionDesc = "desc"
)

// Group By Constants
const (
	GroupByUsageType  = "usage_type"
	GroupBySourceType = "source_type"
	GroupByHour       = "hour"
	GroupByDay        = "day"
	GroupByWeek       = "week"
	GroupByMonth      = "month"
)
