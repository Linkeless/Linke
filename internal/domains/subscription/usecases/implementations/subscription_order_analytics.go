package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
)

// SubscriptionOrderAnalyticsService handles statistics, analytics, and bulk operations
type SubscriptionOrderAnalyticsService struct {
	*SubscriptionOrderServiceBase
}

// NewSubscriptionOrderAnalyticsService creates a new order analytics service
func NewSubscriptionOrderAnalyticsService(base *SubscriptionOrderServiceBase) *SubscriptionOrderAnalyticsService {
	return &SubscriptionOrderAnalyticsService{
		SubscriptionOrderServiceBase: base,
	}
}

// GetOrderStatistics gets comprehensive order statistics
func (soas *SubscriptionOrderAnalyticsService) GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error) {
	query := soas.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

	// Apply date range
	if !fromDate.IsZero() {
		query = query.Where("created_at >= ?", fromDate)
	}

	if !toDate.IsZero() {
		query = query.Where("created_at <= ?", toDate)
	}

	stats := make(map[string]any)

	// Get total orders
	var totalOrders int64
	query.Count(&totalOrders)
	stats["total_orders"] = totalOrders

	// Get orders by status
	var statusCounts []struct {
		Status string
		Count  int64
	}

	if err := query.Select("status, count(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	var pendingOrders, paidOrders, failedOrders, cancelledOrders, refundedOrders int64
	for _, sc := range statusCounts {
		switch sc.Status {
		case constants.OrderStatusPending:
			pendingOrders = sc.Count
		case constants.OrderStatusPaid:
			paidOrders = sc.Count
		case constants.OrderStatusFailed:
			failedOrders = sc.Count
		case constants.OrderStatusCancelled:
			cancelledOrders = sc.Count
		case constants.OrderStatusRefunded:
			refundedOrders = sc.Count
		}
	}

	stats["pending_orders"] = pendingOrders
	stats["paid_orders"] = paidOrders
	stats["failed_orders"] = failedOrders
	stats["cancelled_orders"] = cancelledOrders
	stats["refunded_orders"] = refundedOrders

	// Get revenue statistics
	var revenueStats struct {
		TotalRevenue  float64
		TotalRefunded float64
		AvgOrderValue float64
	}

	if err := query.Select(
		"COALESCE(SUM(CASE WHEN status = 'paid' THEN total_amount ELSE 0 END), 0) as total_revenue, " +
			"COALESCE(SUM(refund_amount), 0) as total_refunded, " +
			"COALESCE(AVG(CASE WHEN status = 'paid' THEN total_amount ELSE NULL END), 0) as avg_order_value",
	).Scan(&revenueStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	stats["total_revenue"] = revenueStats.TotalRevenue
	stats["total_refunded"] = revenueStats.TotalRefunded
	stats["avg_order_value"] = revenueStats.AvgOrderValue

	// Calculate conversion rate (paid orders / total orders)
	if totalOrders > 0 {
		stats["conversion_rate"] = float64(paidOrders) / float64(totalOrders) * 100
	} else {
		stats["conversion_rate"] = 0.0
	}

	return stats, nil
}

// BulkUpdateOrders performs bulk operations on orders
func (soas *SubscriptionOrderAnalyticsService) BulkUpdateOrders(ctx context.Context, req *BulkUpdateOrdersRequest) (*BulkUpdateResult, error) {
	result := &BulkUpdateResult{
		Successful: make([]uint, 0),
		Failed:     make([]BulkUpdateFailure, 0),
	}

	// We need access to the operations service for bulk operations
	// This is a limitation of the current split - we'll need to inject the operations service
	// For now, we'll implement the logic directly but this should be refactored

	for _, orderID := range req.OrderIDs {
		var err error

		switch req.Operation {
		case "cancel":
			// Get order first
			var order entities.SubscriptionOrder
			if getErr := soas.db.WithContext(ctx).First(&order, orderID).Error; getErr != nil {
				err = getErr
			} else if order.Status == constants.OrderStatusPending {
				// Update order status directly for cancel operations
				updateData := map[string]any{
					"status":              constants.OrderStatusCancelled,
					"cancellation_reason": req.Reason,
					"cancelled_at":        time.Now(),
					"updated_at":          time.Now(),
				}
				if updateErr := soas.db.WithContext(ctx).Model(&order).Updates(updateData).Error; updateErr != nil {
					err = updateErr
				}
			} else {
				err = fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
			}
		case "refund":
			// For bulk refund, we'll refund the full amount
			var order entities.SubscriptionOrder
			if getErr := soas.db.WithContext(ctx).First(&order, orderID).Error; getErr != nil {
				err = getErr
			} else if order.CanBeRefunded() {
				refundAmount := order.GetRefundableAmount()
				updateData := map[string]any{
					"refund_amount": order.RefundAmount + refundAmount,
					"refund_reason": req.Reason,
					"refunded_at":   time.Now(),
					"updated_at":    time.Now(),
				}
				// If full refund, update status
				if order.RefundAmount+refundAmount >= order.TotalAmount {
					updateData["status"] = constants.OrderStatusRefunded
				}
				// Add admin notes
				if req.Notes != "" {
					currentNotes := order.Notes
					adminNote := fmt.Sprintf("[%s] Bulk refund processed by Admin (ID:%d): Amount=%.2f, Reason=%s, Notes=%s",
						time.Now().Format("2006-01-02 15:04:05"), req.ProcessedBy, refundAmount, req.Reason, req.Notes)
					if currentNotes != "" {
						updateData["notes"] = currentNotes + "\n" + adminNote
					} else {
						updateData["notes"] = adminNote
					}
				}
				if updateErr := soas.db.WithContext(ctx).Model(&order).Updates(updateData).Error; updateErr != nil {
					err = updateErr
				}
			} else {
				err = fmt.Errorf("order cannot be refunded")
			}
		default:
			err = fmt.Errorf("unsupported operation: %s", req.Operation)
		}

		if err != nil {
			result.Failed = append(result.Failed, BulkUpdateFailure{
				OrderID: orderID,
				Error:   err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, orderID)
		}
	}

	// Generate summary
	result.Summary = map[string]any{
		"total_processed": len(req.OrderIDs),
		"successful":      len(result.Successful),
		"failed":          len(result.Failed),
		"operation":       req.Operation,
	}

	return result, nil
}
