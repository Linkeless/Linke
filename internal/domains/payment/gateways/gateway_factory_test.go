package gateways

import (
	"context"
	"testing"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentConfigService is a mock implementation of PaymentConfigService for testing
type MockPaymentConfigService struct {
	mock.Mock
}

func (m *MockPaymentConfigService) CreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	args := m.Called(ctx, configID)
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) GetPaymentConfigByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error) {
	args := m.Called(ctx, method)
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) UpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	args := m.Called(ctx, configID, req)
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) DeletePaymentConfig(ctx context.Context, configID uint) error {
	args := m.Called(ctx, configID)
	return args.Error(0)
}

func (m *MockPaymentConfigService) GetPaymentConfigs(ctx context.Context, req *dto.GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*entities.PaymentConfig), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentConfigService) GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	args := m.Called(ctx, currency)
	return args.Get(0).([]*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) GetPaymentConfigsByMethod(ctx context.Context, method string) ([]*entities.PaymentConfig, error) {
	args := m.Called(ctx, method)
	return args.Get(0).([]*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	args := m.Called(ctx, configID)
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) GetEnabledConfigs() ([]*entities.PaymentConfig, error) {
	args := m.Called()
	return args.Get(0).([]*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) GetConfigByMethod(method string) (*entities.PaymentConfig, error) {
	args := m.Called(method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PaymentConfig), args.Error(1)
}

func (m *MockPaymentConfigService) ValidatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) []string {
	args := m.Called(ctx, req)
	return args.Get(0).([]string)
}

func (m *MockPaymentConfigService) ValidatePaymentConfigUpdate(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) []string {
	args := m.Called(ctx, configID, req)
	return args.Get(0).([]string)
}

func (m *MockPaymentConfigService) ValidateCreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) []string {
	args := m.Called(ctx, req)
	return args.Get(0).([]string)
}

func (m *MockPaymentConfigService) ValidateUpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) []string {
	args := m.Called(ctx, configID, req)
	return args.Get(0).([]string)
}

func (m *MockPaymentConfigService) PrepareCreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) (*dto.CreatePaymentConfigRequest, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dto.CreatePaymentConfigRequest), args.Error(1)
}

func (m *MockPaymentConfigService) PrepareUpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) (*dto.UpdatePaymentConfigRequest, error) {
	args := m.Called(ctx, configID, req)
	return args.Get(0).(*dto.UpdatePaymentConfigRequest), args.Error(1)
}

// MockPaymentGateway is a mock implementation of PaymentGateway for testing
type MockPaymentGateway struct {
	mock.Mock
}

func (m *MockPaymentGateway) CreatePaymentOrder(req *dto.CreatePaymentOrderRequest) (*dto.CreatePaymentOrderResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*dto.CreatePaymentOrderResponse), args.Error(1)
}

func (m *MockPaymentGateway) QueryPaymentOrder(outTradeNo string) (*dto.QueryPaymentOrderResponse, error) {
	args := m.Called(outTradeNo)
	return args.Get(0).(*dto.QueryPaymentOrderResponse), args.Error(1)
}

func (m *MockPaymentGateway) VerifyPaymentNotify(data map[string]any) (bool, *dto.NotifyData) {
	args := m.Called(data)
	return args.Bool(0), args.Get(1).(*dto.NotifyData)
}

func (m *MockPaymentGateway) IsPaymentCompleted(status string) bool {
	args := m.Called(status)
	return args.Bool(0)
}

func (m *MockPaymentGateway) GetSupportedPaymentMethods() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockPaymentGateway) GetPaymentMethodName(method string) string {
	args := m.Called(method)
	return args.String(0)
}

func (m *MockPaymentGateway) ValidateConfig() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPaymentGateway) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}

// TestGatewayFactory_NewGatewayFactory tests the creation of a new gateway factory
func TestGatewayFactory_NewGatewayFactory(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	assert.NotNil(t, factory)
	assert.NotNil(t, factory.gateways)
	assert.NotNil(t, factory.configs)
	assert.Equal(t, mockConfigService, factory.paymentConfigService)
}

// TestGatewayFactory_RegisterGateway tests gateway registration
func TestGatewayFactory_RegisterGateway(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Create a mock gateway that validates successfully
	mockGateway := &MockPaymentGateway{}
	mockGateway.On("ValidateConfig").Return(nil)
	mockGateway.On("TestConnection").Return(nil)

	err := factory.RegisterGateway("test_gateway", mockGateway)
	assert.NoError(t, err)

	// Verify gateway is registered
	assert.True(t, factory.IsGatewayRegistered("test_gateway"))
	assert.Equal(t, 1, factory.GetGatewayCount())

	mockGateway.AssertExpectations(t)
}

// TestGatewayFactory_RegisterGateway_ValidationFailure tests registration with invalid config
func TestGatewayFactory_RegisterGateway_ValidationFailure(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Create a mock gateway that fails validation
	mockGateway := &MockPaymentGateway{}
	mockGateway.On("ValidateConfig").Return(assert.AnError)

	err := factory.RegisterGateway("test_gateway", mockGateway)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway validation failed")

	// Verify gateway is not registered
	assert.False(t, factory.IsGatewayRegistered("test_gateway"))
	assert.Equal(t, 0, factory.GetGatewayCount())

	mockGateway.AssertExpectations(t)
}

// TestGatewayFactory_GetGateway tests gateway retrieval
func TestGatewayFactory_GetGateway(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Register a mock gateway
	mockGateway := &MockPaymentGateway{}
	mockGateway.On("ValidateConfig").Return(nil)
	mockGateway.On("TestConnection").Return(nil)

	err := factory.RegisterGateway("test_gateway", mockGateway)
	assert.NoError(t, err)

	// Test getting existing gateway
	gateway, err := factory.GetGateway("test_gateway")
	assert.NoError(t, err)
	assert.Equal(t, mockGateway, gateway)

	// Test getting non-existent gateway
	_, err = factory.GetGateway("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway 'non_existent' not found")

	mockGateway.AssertExpectations(t)
}

// TestGatewayFactory_GetAvailableGateways tests listing available gateways
func TestGatewayFactory_GetAvailableGateways(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Initially no gateways
	gateways := factory.GetAvailableGateways()
	assert.Empty(t, gateways)

	// Register mock gateways
	mockGateway1 := &MockPaymentGateway{}
	mockGateway1.On("ValidateConfig").Return(nil)
	mockGateway1.On("TestConnection").Return(nil)

	mockGateway2 := &MockPaymentGateway{}
	mockGateway2.On("ValidateConfig").Return(nil)
	mockGateway2.On("TestConnection").Return(nil)

	err := factory.RegisterGateway("gateway1", mockGateway1)
	assert.NoError(t, err)

	err = factory.RegisterGateway("gateway2", mockGateway2)
	assert.NoError(t, err)

	// Check available gateways
	gateways = factory.GetAvailableGateways()
	assert.Len(t, gateways, 2)
	assert.Contains(t, gateways, "gateway1")
	assert.Contains(t, gateways, "gateway2")

	mockGateway1.AssertExpectations(t)
	mockGateway2.AssertExpectations(t)
}

// TestGatewayFactory_CreateGatewayFromConfig tests gateway creation from config
func TestGatewayFactory_CreateGatewayFromConfig(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	tests := []struct {
		name    string
		config  *entities.PaymentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid epay config",
			config: &entities.PaymentConfig{
				Method: constants.PaymentGatewayEpay,
				URL:    "https://api.epay.test.com/submit.php",
				PID:    "test_pid",
				Key:    "test_key",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "payment config is nil",
		},
		{
			name: "unsupported gateway type",
			config: &entities.PaymentConfig{
				Method: "unsupported_gateway",
				URL:    "https://api.example.com",
				PID:    "test_pid",
				Key:    "test_key",
			},
			wantErr: true,
			errMsg:  "unsupported gateway type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway, err := factory.CreateGatewayFromConfig(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gateway)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gateway)
				assert.IsType(t, &EpayGateway{}, gateway)
			}
		})
	}
}

// TestGatewayFactory_LoadGatewaysFromConfig tests loading gateways from configuration
func TestGatewayFactory_LoadGatewaysFromConfig(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Mock enabled configs
	configs := []*entities.PaymentConfig{
		{
			ID:        1,
			Method:    constants.PaymentGatewayEpay,
			URL:       "https://api.epay.test.com/submit.php",
			PID:       "test_pid",
			Key:       "test_key",
			IsEnabled: true,
			MinAmount: 0.01,
			MaxAmount: 99999.99,
		},
	}

	mockConfigService.On("GetEnabledConfigs").Return(configs, nil)

	err := factory.LoadGatewaysFromConfig()
	assert.NoError(t, err)

	// Verify gateway was loaded
	assert.True(t, factory.IsGatewayRegistered(constants.PaymentGatewayEpay))
	assert.Equal(t, 1, factory.GetGatewayCount())

	mockConfigService.AssertExpectations(t)
}

// TestGatewayFactory_LoadGatewaysFromConfig_NoConfigs tests loading with no configurations
func TestGatewayFactory_LoadGatewaysFromConfig_NoConfigs(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Mock no enabled configs
	configs := []*entities.PaymentConfig{}
	mockConfigService.On("GetEnabledConfigs").Return(configs, nil)

	err := factory.LoadGatewaysFromConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no gateways could be loaded successfully")

	// Verify no gateways were loaded
	assert.Equal(t, 0, factory.GetGatewayCount())

	mockConfigService.AssertExpectations(t)
}

// TestGatewayFactory_ReloadGateway tests gateway reloading
func TestGatewayFactory_ReloadGateway(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Test reloading with updated config
	updatedConfig := &entities.PaymentConfig{
		ID:        1,
		Method:    constants.PaymentGatewayEpay,
		URL:       "https://api.epay.test.com/submit.php",
		PID:       "updated_pid",
		Key:       "updated_key",
		IsEnabled: true,
		MinAmount: 0.01,
		MaxAmount: 99999.99,
	}

	mockConfigService.On("GetConfigByMethod", constants.PaymentGatewayEpay).Return(updatedConfig, nil)

	err := factory.ReloadGateway(constants.PaymentGatewayEpay)
	assert.NoError(t, err)

	// Verify gateway was reloaded
	assert.True(t, factory.IsGatewayRegistered(constants.PaymentGatewayEpay))

	mockConfigService.AssertExpectations(t)
}

// TestGatewayFactory_GetSupportedPaymentMethods tests getting supported payment methods
func TestGatewayFactory_GetSupportedPaymentMethods(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Register mock gateway with supported methods
	mockGateway := &MockPaymentGateway{}
	mockGateway.On("ValidateConfig").Return(nil)
	mockGateway.On("TestConnection").Return(nil)
	mockGateway.On("GetSupportedPaymentMethods").Return([]string{"alipay", "wechat"})

	err := factory.RegisterGateway("test_gateway", mockGateway)
	assert.NoError(t, err)

	methods := factory.GetSupportedPaymentMethods()
	assert.Contains(t, methods, "test_gateway")
	assert.Equal(t, []string{"alipay", "wechat"}, methods["test_gateway"])

	mockGateway.AssertExpectations(t)
}

// TestGatewayFactory_ValidateAllGateways tests validation of all registered gateways
func TestGatewayFactory_ValidateAllGateways(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Register gateways with different validation results
	validGateway := &MockPaymentGateway{}
	validGateway.On("ValidateConfig").Return(nil).Times(2) // Called during registration and validation
	validGateway.On("TestConnection").Return(nil)

	invalidGateway := &MockPaymentGateway{}
	invalidGateway.On("ValidateConfig").Return(nil).Once() // Called during registration
	invalidGateway.On("TestConnection").Return(nil)
	invalidGateway.On("ValidateConfig").Return(assert.AnError).Once() // Called during validation

	err := factory.RegisterGateway("valid_gateway", validGateway)
	assert.NoError(t, err)

	err = factory.RegisterGateway("invalid_gateway", invalidGateway)
	assert.NoError(t, err)

	// Validate all gateways
	err = factory.ValidateAllGateways()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway validation errors")

	validGateway.AssertExpectations(t)
	invalidGateway.AssertExpectations(t)
}

// TestGatewayFactory_TestAllGatewayConnections tests connection testing for all gateways
func TestGatewayFactory_TestAllGatewayConnections(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Register gateways with different connection test results
	workingGateway := &MockPaymentGateway{}
	workingGateway.On("ValidateConfig").Return(nil)
	workingGateway.On("TestConnection").Return(nil).Times(2) // Called during registration and testing

	brokenGateway := &MockPaymentGateway{}
	brokenGateway.On("ValidateConfig").Return(nil)
	brokenGateway.On("TestConnection").Return(nil).Once() // Called during registration
	brokenGateway.On("TestConnection").Return(assert.AnError).Once() // Called during testing

	err := factory.RegisterGateway("working_gateway", workingGateway)
	assert.NoError(t, err)

	err = factory.RegisterGateway("broken_gateway", brokenGateway)
	assert.NoError(t, err)

	// Test all connections
	results := factory.TestAllGatewayConnections()
	assert.Len(t, results, 2)
	assert.NoError(t, results["working_gateway"])
	assert.Error(t, results["broken_gateway"])

	workingGateway.AssertExpectations(t)
	brokenGateway.AssertExpectations(t)
}

// TestGatewayFactory_CreateDefaultEpayGateway tests creation of default epay gateway
func TestGatewayFactory_CreateDefaultEpayGateway(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	tests := []struct {
		name    string
		url     string
		pid     string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid parameters",
			url:     "https://api.epay.test.com/submit.php",
			pid:     "test_pid",
			key:     "test_key",
			wantErr: false,
		},
		{
			name:    "empty url",
			url:     "",
			pid:     "test_pid",
			key:     "test_key",
			wantErr: true,
			errMsg:  "url, pid, and key are required",
		},
		{
			name:    "empty pid",
			url:     "https://api.epay.test.com/submit.php",
			pid:     "",
			key:     "test_key",
			wantErr: true,
			errMsg:  "url, pid, and key are required",
		},
		{
			name:    "empty key",
			url:     "https://api.epay.test.com/submit.php",
			pid:     "test_pid",
			key:     "",
			wantErr: true,
			errMsg:  "url, pid, and key are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset factory for each test
			factory = NewGatewayFactory(mockConfigService)
			
			err := factory.CreateDefaultEpayGateway(tt.url, tt.pid, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, factory.IsGatewayRegistered(constants.PaymentGatewayEpay))
			}
		})
	}
}

// TestGatewayFactory_UnregisterGateway tests gateway unregistration
func TestGatewayFactory_UnregisterGateway(t *testing.T) {
	mockConfigService := &MockPaymentConfigService{}
	factory := NewGatewayFactory(mockConfigService)

	// Register a gateway
	mockGateway := &MockPaymentGateway{}
	mockGateway.On("ValidateConfig").Return(nil)
	mockGateway.On("TestConnection").Return(nil)

	err := factory.RegisterGateway("test_gateway", mockGateway)
	assert.NoError(t, err)
	assert.True(t, factory.IsGatewayRegistered("test_gateway"))

	// Unregister the gateway
	factory.UnregisterGateway("test_gateway")
	assert.False(t, factory.IsGatewayRegistered("test_gateway"))
	assert.Equal(t, 0, factory.GetGatewayCount())

	mockGateway.AssertExpectations(t)
}