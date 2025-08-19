package dto

// BulkOperationResponse represents the result of a bulk operation
type BulkOperationResponse struct {
	SuccessCount int      `json:"success_count" example:"5"`
	FailedCount  int      `json:"failed_count" example:"2"`
	FailedIDs    []uint   `json:"failed_ids,omitempty" example:"[3,7]"`
	Errors       []string `json:"errors,omitempty" example:"[\"Failed to process item 3\", \"Invalid data for item 7\"]"`
	Message      string   `json:"message" example:"Bulk operation completed"`
}

// ToggleStatusResponse represents the result of a status toggle operation
type ToggleStatusResponse struct {
	ID        uint   `json:"id" example:"1"`
	NewStatus string `json:"new_status" example:"active"`
	Message   string `json:"message" example:"Status updated successfully"`
}

// ExtendExpiryResponse represents the result of extending expiry
type ExtendExpiryResponse struct {
	ID        uint   `json:"id" example:"1"`
	NewExpiry string `json:"new_expiry" example:"2024-12-31T23:59:59Z"`
	Message   string `json:"message" example:"Expiry date extended successfully"`
}

// BatchOperationResponse represents the result of a batch operation
type BatchOperationResponse struct {
	ProcessedCount int      `json:"processed_count" example:"10"`
	SuccessCount   int      `json:"success_count" example:"8"`
	FailedCount    int      `json:"failed_count" example:"2"`
	FailedItems    []string `json:"failed_items,omitempty" example:"[\"user123\", \"user456\"]"`
	Message        string   `json:"message" example:"Batch operation completed"`
}

// SearchResultResponse represents search results with metadata
type SearchResultResponse struct {
	Query         string `json:"query" example:"test"`
	TotalResults  int64  `json:"total_results" example:"25"`
	ResultsCount  int    `json:"results_count" example:"10"`
	Page          int    `json:"page" example:"1"`
	PageSize      int    `json:"page_size" example:"10"`
	TotalPages    int    `json:"total_pages" example:"3"`
	ExecutionTime string `json:"execution_time" example:"15ms"`
}

// UpdateStatusResponse represents the result of a status update
type UpdateStatusResponse struct {
	ID        uint   `json:"id" example:"1"`
	OldStatus string `json:"old_status" example:"pending"`
	NewStatus string `json:"new_status" example:"active"`
	Message   string `json:"message" example:"Status updated successfully"`
}

// UpdateRoleResponse represents the result of a role update
type UpdateRoleResponse struct {
	ID      uint   `json:"id" example:"1"`
	OldRole string `json:"old_role" example:"user"`
	NewRole string `json:"new_role" example:"admin"`
	Message string `json:"message" example:"Role updated successfully"`
}

// PasswordResetResponse represents the result of a password reset
type PasswordResetResponse struct {
	UserID    uint   `json:"user_id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	ResetSent bool   `json:"reset_sent" example:"true"`
	Message   string `json:"message" example:"Password reset email sent"`
}

// StatisticsResponse represents general statistics data
type StatisticsResponse struct {
	TotalCount   int64                  `json:"total_count" example:"1000"`
	Period       string                 `json:"period" example:"30d"`
	Breakdown    map[string]int64       `json:"breakdown,omitempty"`
	Trends       []TrendDataPoint       `json:"trends,omitempty"`
	LastUpdated  string                 `json:"last_updated" example:"2024-01-15T10:30:00Z"`
	GeneratedAt  string                 `json:"generated_at" example:"2024-01-15T10:35:00Z"`
}

// TrendDataPoint represents a single point in trend data
type TrendDataPoint struct {
	Date  string `json:"date" example:"2024-01-15"`
	Value int64  `json:"value" example:"150"`
	Label string `json:"label,omitempty" example:"New Users"`
}

// UserStatisticsResponse represents user-specific statistics
type UserStatisticsResponse struct {
	UserID       uint                   `json:"user_id" example:"1"`
	TotalUsers   int64                  `json:"total_users" example:"5000"`
	ActiveUsers  int64                  `json:"active_users" example:"4500"`
	NewUsers     int64                  `json:"new_users" example:"150"`
	Period       string                 `json:"period" example:"30d"`
	StatusBreakdown map[string]int64    `json:"status_breakdown"`
	RoleBreakdown   map[string]int64    `json:"role_breakdown"`
	DailyTrends     []TrendDataPoint    `json:"daily_trends,omitempty"`
	GeneratedAt     string              `json:"generated_at" example:"2024-01-15T10:35:00Z"`
}