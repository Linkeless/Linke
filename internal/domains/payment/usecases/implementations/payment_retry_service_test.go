package implementations

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/queue"
)

// MockPaymentRetryRepository is a mock implementation of PaymentRetryRepository
type MockPaymentRetryRepository struct {
	mock.Mock
}

func (m *MockPaymentRetryRepository) Create(ctx context.Context, retry *entities.PaymentRetry) error {
	args := m.Called(ctx, retry)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentRetry, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.PaymentRetry), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error) {
	args := m.Called(ctx, paymentRecordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PaymentRetry), args.Error(1)
}

func (m *MockPaymentRetryRepository) Update(ctx context.Context, retry *entities.PaymentRetry) error {
	args := m.Called(ctx, retry)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) GetPendingRetries(ctx context.Context, limit int) ([]*entities.PaymentRetry, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entities.PaymentRetry), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetRetriesDueForProcessing(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRetry, error) {
	args := m.Called(ctx, beforeTime, limit)
	return args.Get(0).([]*entities.PaymentRetry), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetActiveRetriesForGateway(ctx context.Context, gateway string) ([]*entities.PaymentRetry, error) {
	args := m.Called(ctx, gateway)
	return args.Get(0).([]*entities.PaymentRetry), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetRetriesForUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) GetRetryStatsByGateway(ctx context.Context, gateway string, fromDate, toDate time.Time) (*interfaces.RetryStatistics, error) {
	args := m.Called(ctx, gateway, fromDate, toDate)
	return args.Get(0).(*interfaces.RetryStatistics), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetRetryStatsByDate(ctx context.Context, fromDate, toDate time.Time) ([]*interfaces.DailyRetryStats, error) {
	args := m.Called(ctx, fromDate, toDate)
	return args.Get(0).([]*interfaces.DailyRetryStats), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetRetrySuccessRate(ctx context.Context, gateway string, days int) (float64, error) {
	args := m.Called(ctx, gateway, days)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockPaymentRetryRepository) GetAllRetries(ctx context.Context, filters *interfaces.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, filters, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) CancelRetry(ctx context.Context, id uint, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) ResetRetry(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) MarkRetriesAsInProgress(ctx context.Context, ids []uint) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) UpdateRetryStatus(ctx context.Context, id uint, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Methods from Repository interface
func (m *MockPaymentRetryRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockPaymentRetryRepository) BeginTransaction() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockPaymentRetryRepository) CommitTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) RollbackTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

// Methods from GenericRepository interface
func (m *MockPaymentRetryRepository) SoftDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) HardDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) List(ctx context.Context, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPaymentRetryRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryRepository) CountDeleted(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryRepository) BatchRestore(ctx context.Context, ids []uint) (int, []uint, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	args := m.Called(ctx, ids, status)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockPaymentRetryRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, filters, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

// Methods from TimeBasedRepository interface
func (m *MockPaymentRetryRepository) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, field, start, end, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, after, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryRepository) ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	args := m.Called(ctx, after, limit, offset)
	return args.Get(0).([]*entities.PaymentRetry), args.Get(1).(int64), args.Error(2)
}

// MockPaymentRetryHistoryRepository is a mock implementation of PaymentRetryHistoryRepository
type MockPaymentRetryHistoryRepository struct {
	mock.Mock
}

func (m *MockPaymentRetryHistoryRepository) Create(ctx context.Context, history *entities.PaymentRetryHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentRetryHistory, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.PaymentRetryHistory), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) GetByRetryID(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error) {
	args := m.Called(ctx, retryID)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	args := m.Called(ctx, paymentRecordID)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) Update(ctx context.Context, history *entities.PaymentRetryHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) GetRecentAttempts(ctx context.Context, retryID uint, limit int) ([]*entities.PaymentRetryHistory, error) {
	args := m.Called(ctx, retryID, limit)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) GetAttemptsForPayment(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	args := m.Called(ctx, paymentRecordID)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) GetAttemptStatistics(ctx context.Context, retryID uint) (*interfaces.AttemptStatistics, error) {
	args := m.Called(ctx, retryID)
	return args.Get(0).(*interfaces.AttemptStatistics), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) GetFailurePatterns(ctx context.Context, gateway string, days int) ([]*interfaces.FailurePattern, error) {
	args := m.Called(ctx, gateway, days)
	return args.Get(0).([]*interfaces.FailurePattern), args.Error(1)
}

// Methods from Repository interface
func (m *MockPaymentRetryHistoryRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockPaymentRetryHistoryRepository) BeginTransaction() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockPaymentRetryHistoryRepository) CommitTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) RollbackTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

// Methods from GenericRepository interface
func (m *MockPaymentRetryHistoryRepository) SoftDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) HardDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) List(ctx context.Context, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPaymentRetryHistoryRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) CountDeleted(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) BatchRestore(ctx context.Context, ids []uint) (int, []uint, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	args := m.Called(ctx, ids, status)
	return args.Get(0).(int), args.Get(1).([]uint), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockPaymentRetryHistoryRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, filters, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

// Methods from TimeBasedRepository interface
func (m *MockPaymentRetryHistoryRepository) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, field, start, end, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, after, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRetryHistoryRepository) ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.PaymentRetryHistory, int64, error) {
	args := m.Called(ctx, after, limit, offset)
	return args.Get(0).([]*entities.PaymentRetryHistory), args.Get(1).(int64), args.Error(2)
}

// MockPaymentService is a mock implementation of PaymentService
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) RegisterGateway(name string, gateway interfaces.PaymentGateway) error {
	args := m.Called(name, gateway)
	return args.Error(0)
}

func (m *MockPaymentService) GetGateway(name string) (interfaces.PaymentGateway, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PaymentGateway), args.Error(1)
}

func (m *MockPaymentService) CreatePaymentOrder(ctx context.Context, req *interfaces.CreatePaymentOrderRequest) (*entities.PaymentRecord, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*entities.PaymentRecord), args.Error(1)
}

func (m *MockPaymentService) GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	args := m.Called(ctx, paymentNo)
	return args.Get(0).(*entities.PaymentRecord), args.Error(1)
}

func (m *MockPaymentService) GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error) {
	args := m.Called(ctx, outTradeNo)
	return args.Get(0).(*entities.PaymentRecord), args.Error(1)
}

func (m *MockPaymentService) UpdatePaymentStatus(ctx context.Context, paymentNo string, status string, transactionID string, paidAt *time.Time) error {
	args := m.Called(ctx, paymentNo, status, transactionID, paidAt)
	return args.Error(0)
}

func (m *MockPaymentService) ProcessNotification(ctx context.Context, gateway string, data map[string]interface{}) error {
	args := m.Called(ctx, gateway, data)
	return args.Error(0)
}

func (m *MockPaymentService) GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.PaymentRecord), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentService) GetAvailablePaymentMethods(ctx context.Context) (map[string][]string, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (m *MockPaymentService) SetSubscriptionOrderService(subscriptionOrderService interfaces.SubscriptionOrderServiceInterface) {
	m.Called(subscriptionOrderService)
}

func (m *MockPaymentService) GeneratePaymentNo() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

// MockTaskQueue is a mock implementation of DelayedTaskQueue interface
type MockTaskQueue struct {
	mock.Mock
}

func (m *MockTaskQueue) Enqueue(ctx context.Context, queueName string, task *queue.Task) error {
	args := m.Called(ctx, queueName, task)
	return args.Error(0)
}

func (m *MockTaskQueue) EnqueueDelayed(ctx context.Context, queueName string, task *queue.Task, delay time.Duration) error {
	args := m.Called(ctx, queueName, task, delay)
	return args.Error(0)
}

func (m *MockTaskQueue) EnqueueAt(ctx context.Context, queueName string, task *queue.Task, processAt time.Time) error {
	args := m.Called(ctx, queueName, task, processAt)
	return args.Error(0)
}

// Test cases for PaymentRetryService

func TestPaymentRetryService_InitiateRetry(t *testing.T) {
	// Setup mocks
	mockRetryRepo := new(MockPaymentRetryRepository)
	mockHistoryRepo := new(MockPaymentRetryHistoryRepository)
	mockPaymentService := new(MockPaymentService)
	mockTaskQueue := new(MockTaskQueue)

	// Create service
	service := NewPaymentRetryService(
		mockRetryRepo,
		mockHistoryRepo,
		mockPaymentService,
		mockTaskQueue,
	)

	ctx := context.Background()
	paymentRecord := &entities.PaymentRecord{
		ID:            1,
		Gateway:       entities.PaymentGatewayEpay,
		PaymentMethod: entities.PaymentMethodAlipay,
		Amount:        100.0,
		Currency:      entities.CurrencyCNY,
	}

	// Test case: New retry
	mockRetryRepo.On("GetByPaymentRecordID", ctx, uint(1)).Return(nil, fmt.Errorf("not found"))
	mockRetryRepo.On("Create", ctx, mock.AnythingOfType("*entities.PaymentRetry")).Return(nil)
	mockTaskQueue.On("EnqueueDelayed", ctx, "payment_retries", mock.AnythingOfType("*queue.Task"), mock.AnythingOfType("time.Duration")).Return(nil)

	retry, err := service.InitiateRetry(ctx, paymentRecord, entities.FailureTypeTemporary, "NETWORK_ERROR", "Network timeout")

	assert.NoError(t, err)
	assert.NotNil(t, retry)
	assert.Equal(t, uint(1), retry.PaymentRecordID)
	assert.Equal(t, entities.FailureTypeTemporary, retry.FailureType)
	assert.Equal(t, "NETWORK_ERROR", retry.LastFailureCode)
	assert.Equal(t, entities.PaymentRetryStatusPending, retry.Status)

	mockRetryRepo.AssertExpectations(t)
	mockTaskQueue.AssertExpectations(t)
}

func TestPaymentRetryService_ClassifyFailure(t *testing.T) {
	// Setup mocks
	mockRetryRepo := new(MockPaymentRetryRepository)
	mockHistoryRepo := new(MockPaymentRetryHistoryRepository)
	mockPaymentService := new(MockPaymentService)
	mockTaskQueue := new(MockTaskQueue)

	// Create service
	service := NewPaymentRetryService(
		mockRetryRepo,
		mockHistoryRepo,
		mockPaymentService,
		mockTaskQueue,
	)

	ctx := context.Background()

	// Test cases
	testCases := []struct {
		name          string
		gateway       string
		paymentMethod string
		errorCode     string
		errorMessage  string
		expected      string
	}{
		{
			name:          "Permanent failure - invalid card",
			gateway:       entities.PaymentGatewayEpay,
			paymentMethod: entities.PaymentMethodAlipay,
			errorCode:     "INVALID_CARD",
			errorMessage:  "Invalid card number",
			expected:      entities.FailureTypePermanent,
		},
		{
			name:          "Network failure",
			gateway:       entities.PaymentGatewayEpay,
			paymentMethod: entities.PaymentMethodAlipay,
			errorCode:     "NETWORK_ERROR",
			errorMessage:  "Connection timeout",
			expected:      entities.FailureTypeNetwork,
		},
		{
			name:          "Gateway failure",
			gateway:       entities.PaymentGatewayEpay,
			paymentMethod: entities.PaymentMethodAlipay,
			errorCode:     "GATEWAY_ERROR",
			errorMessage:  "Gateway temporarily unavailable",
			expected:      entities.FailureTypeGateway,
		},
		{
			name:          "Temporary failure",
			gateway:       entities.PaymentGatewayEpay,
			paymentMethod: entities.PaymentMethodAlipay,
			errorCode:     "UNKNOWN_ERROR",
			errorMessage:  "Unknown error occurred",
			expected:      entities.FailureTypeTemporary,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.ClassifyFailure(ctx, tc.gateway, tc.paymentMethod, tc.errorCode, tc.errorMessage)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPaymentRetryService_GetRetryStrategy(t *testing.T) {
	// Setup mocks
	mockRetryRepo := new(MockPaymentRetryRepository)
	mockHistoryRepo := new(MockPaymentRetryHistoryRepository)
	mockPaymentService := new(MockPaymentService)
	mockTaskQueue := new(MockTaskQueue)

	// Create service
	service := NewPaymentRetryService(
		mockRetryRepo,
		mockHistoryRepo,
		mockPaymentService,
		mockTaskQueue,
	)

	ctx := context.Background()

	// Test getting default strategy for known gateway
	strategy, err := service.GetRetryStrategy(ctx, entities.PaymentGatewayEpay, entities.PaymentMethodAlipay)

	assert.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.Equal(t, entities.PaymentGatewayEpay, strategy.Gateway)
	assert.Equal(t, 3, strategy.MaxAttempts)
	assert.Equal(t, 3600, strategy.InitialDelay) // 1 hour
	assert.Equal(t, entities.RetryStrategyExponential, strategy.Strategy)
}

func TestPaymentRetry_CalculateNextRetryTime(t *testing.T) {
	now := time.Now()

	// Test exponential backoff
	retry := &entities.PaymentRetry{
		AttemptNumber: 0,
		InitialDelay:  3600,  // 1 hour
		MaxDelay:      86400, // 24 hours
		BackoffFactor: 2.0,
		RetryStrategy: entities.RetryStrategyExponential,
	}

	// First attempt - should be initial delay
	nextTime := retry.CalculateNextRetryTime()
	expectedFirstDelay := now.Add(time.Duration(3600) * time.Second)
	assert.WithinDuration(t, expectedFirstDelay, nextTime, time.Second)

	// Second attempt - should be doubled
	retry.AttemptNumber = 1
	nextTime = retry.CalculateNextRetryTime()
	expectedSecondDelay := now.Add(time.Duration(7200) * time.Second) // 2 hours
	assert.WithinDuration(t, expectedSecondDelay, nextTime, time.Second*2)
}

func TestPaymentRetry_ShouldRetry(t *testing.T) {
	testCases := []struct {
		name     string
		retry    *entities.PaymentRetry
		expected bool
	}{
		{
			name: "Should retry - temporary failure, attempts left",
			retry: &entities.PaymentRetry{
				AttemptNumber: 1,
				MaxAttempts:   3,
				FailureType:   entities.FailureTypeTemporary,
			},
			expected: true,
		},
		{
			name: "Should not retry - permanent failure",
			retry: &entities.PaymentRetry{
				AttemptNumber: 1,
				MaxAttempts:   3,
				FailureType:   entities.FailureTypePermanent,
			},
			expected: false,
		},
		{
			name: "Should not retry - max attempts reached",
			retry: &entities.PaymentRetry{
				AttemptNumber: 3,
				MaxAttempts:   3,
				FailureType:   entities.FailureTypeTemporary,
			},
			expected: false,
		},
		{
			name: "Should retry - network failure, attempts left",
			retry: &entities.PaymentRetry{
				AttemptNumber: 1,
				MaxAttempts:   3,
				FailureType:   entities.FailureTypeNetwork,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.retry.ShouldRetry()
			assert.Equal(t, tc.expected, result)
		})
	}
}
