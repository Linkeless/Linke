package implementations

import (
	"testing"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"

	"github.com/stretchr/testify/assert"
)

func TestUserSubscription_PauseResume_EntityMethods(t *testing.T) {
	t.Run("CanBePaused returns true for active subscription", func(t *testing.T) {
		subscription := &entities.UserSubscription{
			Status: constants.UserSubscriptionStatusActive,
		}
		assert.True(t, subscription.CanBePaused())
	})

	t.Run("CanBePaused returns false for non-active subscription", func(t *testing.T) {
		subscription := &entities.UserSubscription{
			Status: constants.UserSubscriptionStatusCancelled,
		}
		assert.False(t, subscription.CanBePaused())
	})

	t.Run("CanBeResumed returns true for paused subscription", func(t *testing.T) {
		subscription := &entities.UserSubscription{
			Status: constants.UserSubscriptionStatusPaused,
		}
		assert.True(t, subscription.CanBeResumed())
	})

	t.Run("CanBeResumed returns false for non-paused subscription", func(t *testing.T) {
		subscription := &entities.UserSubscription{
			Status: constants.UserSubscriptionStatusActive,
		}
		assert.False(t, subscription.CanBeResumed())
	})

	t.Run("IsPaused returns true for paused subscription", func(t *testing.T) {
		subscription := &entities.UserSubscription{
			Status: constants.UserSubscriptionStatusPaused,
		}
		assert.True(t, subscription.IsPaused())
	})

	t.Run("GetPauseDuration calculates correct pause duration", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -5) // 5 days ago
		subscription := &entities.UserSubscription{
			PausedAt: &pausedAt,
		}
		duration := subscription.GetPauseDuration()
		assert.Equal(t, 5, duration)
	})

	t.Run("IsMaxPauseDurationExceeded returns true when exceeded", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -100) // 100 days ago
		subscription := &entities.UserSubscription{
			PausedAt:         &pausedAt,
			MaxPauseDuration: 90,
		}
		assert.True(t, subscription.IsMaxPauseDurationExceeded())
	})

	t.Run("ShouldAutoResume returns true for paused subscription exceeding max duration", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -100) // 100 days ago
		subscription := &entities.UserSubscription{
			Status:           constants.UserSubscriptionStatusPaused,
			PausedAt:         &pausedAt,
			MaxPauseDuration: 90,
		}
		assert.True(t, subscription.ShouldAutoResume())
	})

	t.Run("GetRemainingPauseDays calculates correct remaining days", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -30) // 30 days ago
		subscription := &entities.UserSubscription{
			PausedAt:         &pausedAt,
			MaxPauseDuration: 90,
		}
		remaining := subscription.GetRemainingPauseDays()
		assert.Equal(t, 60, remaining)
	})
}

func TestUserSubscription_PauseResumeRequests(t *testing.T) {
	t.Run("PauseSubscriptionRequest validation", func(t *testing.T) {
		req := &interfaces.PauseSubscriptionRequest{
			Reason:           "User requested pause",
			MaxPauseDuration: intPtr(60),
		}

		assert.Equal(t, "User requested pause", req.Reason)
		assert.Equal(t, 60, *req.MaxPauseDuration)
	})

	t.Run("ResumeSubscriptionRequest validation", func(t *testing.T) {
		req := &interfaces.ResumeSubscriptionRequest{
			AdjustBillingDate: true,
		}

		assert.True(t, req.AdjustBillingDate)
	})
}

func TestUserSubscription_ResponseFields(t *testing.T) {
	t.Run("ToResponse includes pause fields for paused subscription", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -30) // 30 days ago
		adminID := uint(1)

		subscription := &entities.UserSubscription{
			ID:               1,
			Status:           constants.UserSubscriptionStatusPaused,
			PausedAt:         &pausedAt,
			PauseReason:      "User request",
			PausedByAdminID:  &adminID,
			MaxPauseDuration: 90,
		}

		response := subscription.ToResponse()

		assert.Equal(t, constants.UserSubscriptionStatusPaused, response.Status)
		assert.True(t, response.IsPaused)
		assert.Equal(t, &pausedAt, response.PausedAt)
		assert.Equal(t, "User request", response.PauseReason)
		assert.Equal(t, &adminID, response.PausedByAdminID)
		assert.Equal(t, 90, response.MaxPauseDuration)
		assert.Equal(t, 30, response.PauseDurationDays)  // 30 days paused
		assert.Equal(t, 60, response.RemainingPauseDays) // 60 days remaining
		assert.False(t, response.IsMaxPauseDurationExceeded)
	})

	t.Run("ToResponse includes resume fields for resumed subscription", func(t *testing.T) {
		pausedAt := time.Now().AddDate(0, 0, -30) // 30 days ago
		resumedAt := time.Now()                   // Just resumed
		adminID := uint(1)

		subscription := &entities.UserSubscription{
			ID:               1,
			Status:           constants.UserSubscriptionStatusActive,
			PausedAt:         &pausedAt,
			PauseReason:      "User request",
			PausedByAdminID:  &adminID,
			MaxPauseDuration: 90,
			ResumedAt:        &resumedAt,
			ResumedByAdminID: &adminID,
		}

		response := subscription.ToResponse()

		assert.Equal(t, constants.UserSubscriptionStatusActive, response.Status)
		assert.False(t, response.IsPaused)
		assert.Equal(t, &pausedAt, response.PausedAt)
		assert.Equal(t, "User request", response.PauseReason)
		assert.Equal(t, &adminID, response.PausedByAdminID)
		assert.Equal(t, 90, response.MaxPauseDuration)
		assert.Equal(t, &resumedAt, response.ResumedAt)
		assert.Equal(t, &adminID, response.ResumedByAdminID)
		assert.Equal(t, 30, response.PauseDurationDays)  // Shows historical pause duration
		assert.Equal(t, 60, response.RemainingPauseDays) // Remaining days from when it was paused
		assert.False(t, response.IsMaxPauseDurationExceeded)
	})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
