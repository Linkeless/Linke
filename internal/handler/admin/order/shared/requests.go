package shared

// GetOrdersRequest represents the request to get orders with filtering
type GetOrdersRequest struct {
	UserID         *uint   `form:"user_id" example:"1"`
	Status         string  `form:"status" example:"paid"`
	OrderType      string  `form:"order_type" example:"new"`
	PaymentMethod  string  `form:"payment_method" example:"alipay"`
	PaymentGateway string  `form:"payment_gateway" example:"epay"`
	MinAmount      float64 `form:"min_amount" example:"0"`
	MaxAmount      float64 `form:"max_amount" example:"1000"`
	StartDate      string  `form:"start_date" example:"2024-01-01"`
	EndDate        string  `form:"end_date" example:"2024-12-31"`
	CouponCode     string  `form:"coupon_code" example:"SAVE20"`
	Search         string  `form:"search" example:"ORD2024001"`
	SortBy         string  `form:"sort_by" example:"created_at"`
	SortOrder      string  `form:"sort_order" example:"desc"`
	Limit          int     `form:"limit" example:"10"`
	Offset         int     `form:"offset" example:"0"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status         string                  `json:"status" binding:"required,oneof=pending paid failed cancelled refunded" example:"paid"`
	Notes          string                  `json:"notes,omitempty" example:"Manual status update by admin"`
	Reason         string                  `json:"reason,omitempty" example:"Payment verification completed"`
	NotifyUser     bool                    `json:"notify_user,omitempty" example:"true"`
	// Payment evidence required when marking as paid
	PaymentEvidence *PaymentEvidenceRequest `json:"payment_evidence,omitempty"`
	// Admin confirmation for critical operations
	AdminConfirmed bool `json:"admin_confirmed,omitempty" example:"true"`
}

// ProcessRefundRequest represents the request to process a refund
type ProcessRefundRequest struct {
	Amount         float64 `json:"amount" binding:"required,min=0.01" example:"29.99"`
	Reason         string  `json:"reason" binding:"required" example:"Customer request"`
	RefundMethod   string  `json:"refund_method,omitempty" example:"original"`
	Notes          string  `json:"notes,omitempty" example:"Refund processed by admin"`
	NotifyUser     bool    `json:"notify_user,omitempty" example:"true"`
	AdminConfirmed bool    `json:"admin_confirmed" binding:"required" example:"true"`
}

// GetOrderStatsResponse represents order statistics
type GetOrderStatsResponse struct {
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	PaidOrders      int64   `json:"paid_orders"`
	FailedOrders    int64   `json:"failed_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	RefundedOrders  int64   `json:"refunded_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalRefunded   float64 `json:"total_refunded"`
	AvgOrderValue   float64 `json:"avg_order_value"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// BulkUpdateRequest represents bulk operation request
type BulkUpdateRequest struct {
	OrderIDs       []uint `json:"order_ids" binding:"required,min=1" example:"1,2,3"`
	Operation      string `json:"operation" binding:"required,oneof=cancel refund export" example:"cancel"`
	Reason         string `json:"reason,omitempty" example:"Bulk cancellation"`
	Notes          string `json:"notes,omitempty" example:"Bulk operation by admin"`
	AdminConfirmed bool   `json:"admin_confirmed" binding:"required" example:"true"`
}