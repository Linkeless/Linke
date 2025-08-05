package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	paymentEntities "linke/internal/domains/payment/entities"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionOrderService mocks the subscription order service
type MockSubscriptionOrderService struct {
	mock.Mock
}

func (m *MockSubscriptionOrderService) CreateSubscriptionOrder(ctx context.Context, req *interfaces.CreateSubscriptionOrderRequest) (*interfaces.CreateSubscriptionOrderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.CreateSubscriptionOrderResponse), args.Error(1)
}

func (m *MockSubscriptionOrderService) GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*entities.SubscriptionOrder), args.Error(1)
}

func (m *MockSubscriptionOrderService) GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	args := m.Called(ctx, orderNumber)
	return args.Get(0).(*entities.SubscriptionOrder), args.Error(1)
}

func (m *MockSubscriptionOrderService) ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockSubscriptionOrderService) CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error {
	args := m.Called(ctx, orderID, reason)
	return args.Error(0)
}

func (m *MockSubscriptionOrderService) GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.SubscriptionOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockSubscriptionOrderService) GetSubscriptionOrders(ctx context.Context, req *interfaces.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*entities.SubscriptionOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockSubscriptionOrderService) GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error) {
	args := m.Called(ctx, fromDate, toDate)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockSubscriptionOrderService) QuickPurchase(ctx context.Context, req *interfaces.QuickPurchaseRequest) (*interfaces.QuickPurchaseResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.QuickPurchaseResponse), args.Error(1)
}

func TestQuickPurchaseHandler_QuickPurchase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful quick purchase", func(t *testing.T) {
		// Setup
		mockService := new(MockSubscriptionOrderService)
		handler := NewQuickPurchaseHandler(mockService)

		// Create test user
		user := &userEntities.User{
			ID:    1,
			Email: "test@example.com",
		}

		// Mock service response
		expectedResponse := &interfaces.QuickPurchaseResponse{
			PaymentRecord: &paymentEntities.PaymentRecordResponse{
				PaymentNo: "PAY123456789",
			},
			PaymentURL: "https://payment.example.com/pay/123",
			ExpiredAt:  time.Now().Add(30 * time.Minute),
		}

		mockService.On("QuickPurchase", mock.Anything, mock.AnythingOfType("*interfaces.QuickPurchaseRequest")).
			Return(expectedResponse, nil)

		// Create request
		requestBody := interfaces.QuickPurchaseRequest{
			UserID:         1,
			PlanID:         1,
			PaymentGateway: "epay",
			PaymentMethod:  "alipay",
		}

		jsonBody, _ := json.Marshal(requestBody)

		// Setup Gin context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/subscription/quick-purchase", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.AuthContextKey, user)

		// Execute
		handler.QuickPurchase(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized request", func(t *testing.T) {
		// Setup
		mockService := new(MockSubscriptionOrderService)
		handler := NewQuickPurchaseHandler(mockService)

		requestBody := interfaces.QuickPurchaseRequest{
			UserID:         1,
			PlanID:         1,
			PaymentGateway: "epay",
			PaymentMethod:  "alipay",
		}

		jsonBody, _ := json.Marshal(requestBody)

		// Setup Gin context without user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/subscription/quick-purchase", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Execute
		handler.QuickPurchase(c)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		mockService := new(MockSubscriptionOrderService)
		handler := NewQuickPurchaseHandler(mockService)

		user := &userEntities.User{
			ID:    1,
			Email: "test@example.com",
		}

		// Invalid JSON
		invalidJSON := []byte(`{"invalid": json}`)

		// Setup Gin context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/subscription/quick-purchase", bytes.NewBuffer(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.AuthContextKey, user)

		// Execute
		handler.QuickPurchase(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
