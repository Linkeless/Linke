package handlers

import (
	"fmt"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionBulkHandler handles bulk operations on subscriptions
type AdminSubscriptionBulkHandler struct {
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionBulkHandler creates a new admin subscription bulk handler
func NewAdminSubscriptionBulkHandler(base *AdminSubscriptionHandlerBase) *AdminSubscriptionBulkHandler {
	return &AdminSubscriptionBulkHandler{
		AdminSubscriptionHandlerBase: base,
	}
}

// BulkSubscriptionAction godoc
// @Summary Bulk subscription actions
// @Description Perform bulk actions on multiple subscriptions (Admin only)
// @Tags Admin-Subscription-Bulk
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkSubscriptionActionRequest true "Bulk action data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/bulk/action [post]
func (h *AdminSubscriptionBulkHandler) BulkSubscriptionAction(c *gin.Context) {
	var bulkReq BulkSubscriptionActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedIDs := make([]uint, 0)
	errors := make([]string, 0)

	for _, subID := range bulkReq.SubscriptionIDs {
		var err error

		switch bulkReq.Action {
		case "pause":
			reason := "Bulk admin pause"
			if bulkReq.Reason != nil {
				reason = *bulkReq.Reason
			}
			pauseReq := &interfaces.PauseSubscriptionRequest{
				Reason: reason,
			}
			_, err = h.userSubscriptionService.PauseUserSubscription(c.Request.Context(), subID, pauseReq, 0)

		case "resume":
			resumeReq := &interfaces.ResumeSubscriptionRequest{
				AdjustBillingDate: true,
			}
			_, err = h.userSubscriptionService.ResumeUserSubscription(c.Request.Context(), subID, resumeReq, 0)

		case "cancel":
			reason := "Bulk admin cancellation"
			if bulkReq.Reason != nil {
				reason = *bulkReq.Reason
			}
			err = h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), subID, reason, false)

		case "extend":
			if bulkReq.ExtendByDays == nil {
				err = fmt.Errorf("extend_by_days is required for extend action")
			} else {
				reason := "Bulk admin extension"
				if bulkReq.Reason != nil {
					reason = *bulkReq.Reason
				}
				err = h.userSubscriptionService.ExtendSubscription(c.Request.Context(), subID, *bulkReq.ExtendByDays, reason)
			}

		case "reset_traffic":
			_, err = h.userSubscriptionService.ResetTrafficUsage(c.Request.Context(), subID, 0)

		default:
			err = fmt.Errorf("unknown action: %s", bulkReq.Action)
		}

		if err != nil {
			failedIDs = append(failedIDs, subID)
			errors = append(errors, fmt.Sprintf("ID %d: %s", subID, err.Error()))
		} else {
			successCount++
		}
	}

	logger.Info("Admin performed bulk subscription action",
		logger.String("action", bulkReq.Action),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", fmt.Sprintf("bulk_%s", bulkReq.Action)),
	)

	result := gin.H{
		"action":        bulkReq.Action,
		"total_count":   len(bulkReq.SubscriptionIDs),
		"success_count": successCount,
		"failed_count":  len(failedIDs),
		"failed_ids":    failedIDs,
		"errors":        errors,
	}

	if len(failedIDs) > 0 {
		response.OK(c, gin.H{"message": fmt.Sprintf("Bulk action completed with %d successes and %d failures", successCount, len(failedIDs)), "data": result})
	} else {
		response.OK(c, gin.H{"message": fmt.Sprintf("Bulk action completed successfully for all %d subscriptions", successCount), "data": result})
	}
}