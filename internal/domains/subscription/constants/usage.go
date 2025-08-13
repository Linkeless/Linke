package constants

// Usage Type Constants
const (
	UsageTypeTraffic     = "traffic"
	UsageTypeAPICall     = "api_call"
	UsageTypeStorage     = "storage"
	UsageTypeBandwidth   = "bandwidth"
	UsageTypeConnections = "connections"
)

// Unit Constants
const (
	UnitBytes = "bytes"
	UnitCount = "count"
	UnitMB    = "mb"
	UnitGB    = "gb"
	UnitTB    = "tb"
)

// Source Type Constants
const (
	SourceTypeServer = "server"
	SourceTypeAPI    = "api"
	SourceTypeAdmin  = "admin"
	SourceTypeSystem = "system"
	SourceTypeUser   = "user"
)

// Prediction Type Constants
const (
	PredictionTypeDaily     = "daily"
	PredictionTypeWeekly    = "weekly"
	PredictionTypeMonthly   = "monthly"
	PredictionTypePeriodEnd = "period_end"
)

// Trend Constants
const (
	TrendIncreasing = "increasing"
	TrendDecreasing = "decreasing"
	TrendStable     = "stable"
)

// Trend Direction Constants
const (
	TrendDirectionUp     = "up"
	TrendDirectionDown   = "down"
	TrendDirectionStable = "stable"
)

// Anomaly Type Constants
const (
	AnomalyTypeSpike         = "spike"
	AnomalyTypeDrop          = "drop"
	AnomalyTypePatternChange = "pattern_change"
)

// Export Format Constants
const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXLSX = "xlsx"
)

// Default Usage Values
const (
	DefaultPageSize       = 50
	MaxPageSize           = 1000
	DefaultPredictionDays = 30
	MaxPredictionDays     = 90
)
