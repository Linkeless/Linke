package aggregate

import (
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/payment/domain/entity"
	"linke/internal/payment/domain/event"
	"linke/internal/payment/domain/valueobject"
	"linke/internal/shared/domain"
)

// PaymentConfig represents the payment configuration aggregate root
type PaymentConfig struct {
	// Identity
	id valueobject.PaymentConfigID
	
	// Gateway information
	gateway valueobject.PaymentGateway
	name    string
	
	// Configuration
	config string // JSON configuration
	
	// Status and settings
	isEnabled bool
	sortOrder int
	
	// Currency and methods
	supportedCurrencies []valueobject.Currency
	supportedMethods    []entity.PaymentMethodConfig
	
	// Limits
	minAmount valueobject.Money
	maxAmount valueobject.Money
	
	// Fee information
	fixedFee      valueobject.Money
	percentageFee float64 // Percentage as decimal (e.g., 0.025 for 2.5%)
	
	// Audit fields
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	
	// Domain events
	domainEvents []domain.DomainEvent
}

// NewPaymentConfig creates a new payment configuration aggregate
func NewPaymentConfig(
	gateway valueobject.PaymentGateway,
	name string,
	config string,
	supportedCurrencies []valueobject.Currency,
	baseCurrency valueobject.Currency,
) (*PaymentConfig, error) {
	
	if gateway.IsEmpty() {
		return nil, fmt.Errorf("payment gateway cannot be empty")
	}
	
	if name == "" {
		return nil, fmt.Errorf("payment config name cannot be empty")
	}
	
	if config == "" {
		return nil, fmt.Errorf("payment config cannot be empty")
	}
	
	if len(supportedCurrencies) == 0 {
		return nil, fmt.Errorf("supported currencies cannot be empty")
	}
	
	// Validate config is valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return nil, fmt.Errorf("invalid JSON configuration: %w", err)
	}
	
	// Set default limits
	minAmount, _ := valueobject.NewMoney(0.01, baseCurrency)
	maxAmount, _ := valueobject.NewMoney(99999.99, baseCurrency)
	fixedFee := valueobject.NewZeroMoney(baseCurrency)
	
	now := time.Now()
	paymentConfig := &PaymentConfig{
		gateway:             gateway,
		name:                name,
		config:              config,
		isEnabled:           true,
		sortOrder:           0,
		supportedCurrencies: supportedCurrencies,
		supportedMethods:    make([]entity.PaymentMethodConfig, 0),
		minAmount:           minAmount,
		maxAmount:           maxAmount,
		fixedFee:           fixedFee,
		percentageFee:      0.0,
		createdAt:          now,
		updatedAt:          now,
		domainEvents:       make([]domain.DomainEvent, 0),
	}
	
	// Add default supported methods based on gateway
	defaultMethods := gateway.GetSupportedMethods()
	for i, method := range defaultMethods {
		methodConfig := entity.NewPaymentMethodConfig(
			method,
			method.GetDisplayName(),
			"",
			"",
			true,
			i,
			entity.FeeTypeNone,
			0.0,
			0.0,
			0.0,
			entity.EnvironmentProduction,
		)
		paymentConfig.supportedMethods = append(paymentConfig.supportedMethods, methodConfig)
	}
	
	// Add domain event
	event := event.NewPaymentConfigCreatedEvent(
		gateway,
		name,
		supportedCurrencies,
		now,
	)
	paymentConfig.AddDomainEvent(event)
	
	return paymentConfig, nil
}

// Factory method for loading from persistence
func LoadPaymentConfig(
	id valueobject.PaymentConfigID,
	gateway valueobject.PaymentGateway,
	name string,
	config string,
	isEnabled bool,
	sortOrder int,
	supportedCurrencies []valueobject.Currency,
	supportedMethods []entity.PaymentMethodConfig,
	minAmount valueobject.Money,
	maxAmount valueobject.Money,
	fixedFee valueobject.Money,
	percentageFee float64,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *PaymentConfig {
	return &PaymentConfig{
		id:                  id,
		gateway:             gateway,
		name:                name,
		config:              config,
		isEnabled:           isEnabled,
		sortOrder:           sortOrder,
		supportedCurrencies: supportedCurrencies,
		supportedMethods:    supportedMethods,
		minAmount:           minAmount,
		maxAmount:           maxAmount,
		fixedFee:           fixedFee,
		percentageFee:      percentageFee,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		deletedAt:          deletedAt,
		domainEvents:       make([]domain.DomainEvent, 0),
	}
}

// Aggregate root interface implementation
func (pc *PaymentConfig) ID() valueobject.PaymentConfigID {
	return pc.id
}

func (pc *PaymentConfig) DomainEvents() []domain.DomainEvent {
	return pc.domainEvents
}

func (pc *PaymentConfig) ClearDomainEvents() {
	pc.domainEvents = make([]domain.DomainEvent, 0)
}

func (pc *PaymentConfig) AddDomainEvent(event domain.DomainEvent) {
	pc.domainEvents = append(pc.domainEvents, event)
}

func (pc *PaymentConfig) IsDeleted() bool {
	return pc.deletedAt != nil
}

// Getters
func (pc *PaymentConfig) Gateway() valueobject.PaymentGateway {
	return pc.gateway
}

func (pc *PaymentConfig) Name() string {
	return pc.name
}

func (pc *PaymentConfig) Config() string {
	return pc.config
}

func (pc *PaymentConfig) IsEnabled() bool {
	return pc.isEnabled
}

func (pc *PaymentConfig) SortOrder() int {
	return pc.sortOrder
}

func (pc *PaymentConfig) SupportedCurrencies() []valueobject.Currency {
	return pc.supportedCurrencies
}

func (pc *PaymentConfig) SupportedMethods() []entity.PaymentMethodConfig {
	return pc.supportedMethods
}

func (pc *PaymentConfig) MinAmount() valueobject.Money {
	return pc.minAmount
}

func (pc *PaymentConfig) MaxAmount() valueobject.Money {
	return pc.maxAmount
}

func (pc *PaymentConfig) FixedFee() valueobject.Money {
	return pc.fixedFee
}

func (pc *PaymentConfig) PercentageFee() float64 {
	return pc.percentageFee
}

func (pc *PaymentConfig) CreatedAt() time.Time {
	return pc.createdAt
}

func (pc *PaymentConfig) UpdatedAt() time.Time {
	return pc.updatedAt
}

func (pc *PaymentConfig) DeletedAt() *time.Time {
	return pc.deletedAt
}

// Business methods

// UpdateName updates the configuration name
func (pc *PaymentConfig) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("payment config name cannot be empty")
	}
	
	if pc.name != name {
		pc.name = name
		pc.updatedAt = time.Now()
		
		// Add domain event
		event := event.NewPaymentConfigUpdatedEvent(
			pc.gateway,
			pc.name,
			pc.isEnabled,
			pc.updatedAt,
		)
		pc.AddDomainEvent(event)
	}
	
	return nil
}

// UpdateConfig updates the gateway configuration
func (pc *PaymentConfig) UpdateConfig(config string) error {
	if config == "" {
		return fmt.Errorf("payment config cannot be empty")
	}
	
	// Validate config is valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return fmt.Errorf("invalid JSON configuration: %w", err)
	}
	
	if pc.config != config {
		pc.config = config
		pc.updatedAt = time.Now()
		
		// Add domain event
		event := event.NewPaymentConfigUpdatedEvent(
			pc.gateway,
			pc.name,
			pc.isEnabled,
			pc.updatedAt,
		)
		pc.AddDomainEvent(event)
	}
	
	return nil
}

// Enable enables the payment configuration
func (pc *PaymentConfig) Enable() error {
	if !pc.isEnabled {
		pc.isEnabled = true
		pc.updatedAt = time.Now()
		
		// Add domain event
		event := event.NewPaymentConfigEnabledEvent(
			pc.gateway,
			pc.name,
			pc.updatedAt,
		)
		pc.AddDomainEvent(event)
	}
	
	return nil
}

// Disable disables the payment configuration
func (pc *PaymentConfig) Disable() error {
	if pc.isEnabled {
		pc.isEnabled = false
		pc.updatedAt = time.Now()
		
		// Add domain event
		event := event.NewPaymentConfigDisabledEvent(
			pc.gateway,
			pc.name,
			pc.updatedAt,
		)
		pc.AddDomainEvent(event)
	}
	
	return nil
}

// UpdateSortOrder updates the sort order
func (pc *PaymentConfig) UpdateSortOrder(sortOrder int) error {
	if sortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}
	
	if pc.sortOrder != sortOrder {
		pc.sortOrder = sortOrder
		pc.updatedAt = time.Now()
	}
	
	return nil
}

// AddSupportedCurrency adds a supported currency
func (pc *PaymentConfig) AddSupportedCurrency(currency valueobject.Currency) error {
	// Check if currency is already supported
	for _, supportedCurrency := range pc.supportedCurrencies {
		if supportedCurrency.Equals(currency) {
			return fmt.Errorf("currency %s is already supported", currency.String())
		}
	}
	
	pc.supportedCurrencies = append(pc.supportedCurrencies, currency)
	pc.updatedAt = time.Now()
	
	return nil
}

// RemoveSupportedCurrency removes a supported currency
func (pc *PaymentConfig) RemoveSupportedCurrency(currency valueobject.Currency) error {
	if len(pc.supportedCurrencies) <= 1 {
		return fmt.Errorf("cannot remove the last supported currency")
	}
	
	for i, supportedCurrency := range pc.supportedCurrencies {
		if supportedCurrency.Equals(currency) {
			// Remove currency from slice
			pc.supportedCurrencies = append(pc.supportedCurrencies[:i], pc.supportedCurrencies[i+1:]...)
			pc.updatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("currency %s is not supported", currency.String())
}

// AddSupportedMethod adds a supported payment method
func (pc *PaymentConfig) AddSupportedMethod(methodConfig entity.PaymentMethodConfig) error {
	// Check if method is already supported
	for _, supportedMethod := range pc.supportedMethods {
		if supportedMethod.Method().Equals(methodConfig.Method()) {
			return fmt.Errorf("method %s is already supported", methodConfig.Method().String())
		}
	}
	
	pc.supportedMethods = append(pc.supportedMethods, methodConfig)
	pc.updatedAt = time.Now()
	
	return nil
}

// UpdateSupportedMethod updates a supported payment method
func (pc *PaymentConfig) UpdateSupportedMethod(method valueobject.PaymentMethod, methodConfig entity.PaymentMethodConfig) error {
	for i, supportedMethod := range pc.supportedMethods {
		if supportedMethod.Method().Equals(method) {
			pc.supportedMethods[i] = methodConfig
			pc.updatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("method %s is not supported", method.String())
}

// RemoveSupportedMethod removes a supported payment method
func (pc *PaymentConfig) RemoveSupportedMethod(method valueobject.PaymentMethod) error {
	for i, supportedMethod := range pc.supportedMethods {
		if supportedMethod.Method().Equals(method) {
			// Remove method from slice
			pc.supportedMethods = append(pc.supportedMethods[:i], pc.supportedMethods[i+1:]...)
			pc.updatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("method %s is not supported", method.String())
}

// UpdateAmountLimits updates the minimum and maximum amount limits
func (pc *PaymentConfig) UpdateAmountLimits(minAmount, maxAmount valueobject.Money) error {
	// Validate currencies match
	if !minAmount.Currency().Equals(maxAmount.Currency()) {
		return fmt.Errorf("min and max amount currencies must match")
	}
	
	// Validate min is less than max
	if greater, err := minAmount.IsGreaterThan(maxAmount); err != nil {
		return fmt.Errorf("failed to compare amounts: %w", err)
	} else if greater {
		return fmt.Errorf("minimum amount cannot be greater than maximum amount")
	}
	
	pc.minAmount = minAmount
	pc.maxAmount = maxAmount
	pc.updatedAt = time.Now()
	
	return nil
}

// UpdateFeeSettings updates the fee settings
func (pc *PaymentConfig) UpdateFeeSettings(fixedFee valueobject.Money, percentageFee float64) error {
	if percentageFee < 0 || percentageFee > 100 {
		return fmt.Errorf("percentage fee must be between 0 and 100")
	}
	
	pc.fixedFee = fixedFee
	pc.percentageFee = percentageFee
	pc.updatedAt = time.Now()
	
	return nil
}

// Business queries

// IsActive checks if the payment config is active
func (pc *PaymentConfig) IsActive() bool {
	return pc.isEnabled && !pc.IsDeleted()
}

// SupportsCurrency checks if the config supports a specific currency
func (pc *PaymentConfig) SupportsCurrency(currency valueobject.Currency) bool {
	for _, supportedCurrency := range pc.supportedCurrencies {
		if supportedCurrency.Equals(currency) {
			return true
		}
	}
	return false
}

// SupportsMethod checks if the config supports a specific payment method
func (pc *PaymentConfig) SupportsMethod(method valueobject.PaymentMethod) bool {
	for _, supportedMethod := range pc.supportedMethods {
		if supportedMethod.Method().Equals(method) && supportedMethod.IsEnabled() {
			return true
		}
	}
	return false
}

// GetMethodConfig returns the configuration for a specific payment method
func (pc *PaymentConfig) GetMethodConfig(method valueobject.PaymentMethod) (entity.PaymentMethodConfig, error) {
	for _, supportedMethod := range pc.supportedMethods {
		if supportedMethod.Method().Equals(method) {
			return supportedMethod, nil
		}
	}
	
	return entity.PaymentMethodConfig{}, fmt.Errorf("method %s is not supported", method.String())
}

// IsAmountValid checks if the amount is within the valid range
func (pc *PaymentConfig) IsAmountValid(amount valueobject.Money) (bool, error) {
	// Check currency is supported
	if !pc.SupportsCurrency(amount.Currency()) {
		return false, fmt.Errorf("currency %s is not supported", amount.Currency().String())
	}
	
	// Compare with limits (assuming same currency)
	minValid, err := amount.IsGreaterThanOrEqual(pc.minAmount)
	if err != nil {
		return false, fmt.Errorf("failed to compare with minimum amount: %w", err)
	}
	
	maxValid, err := amount.IsLessThanOrEqual(pc.maxAmount)
	if err != nil {
		return false, fmt.Errorf("failed to compare with maximum amount: %w", err)
	}
	
	return minValid && maxValid, nil
}

// CalculateTotalFee calculates the total fee for a given amount
func (pc *PaymentConfig) CalculateTotalFee(amount valueobject.Money, method valueobject.PaymentMethod) (valueobject.Money, error) {
	// Get method-specific fee configuration
	methodConfig, err := pc.GetMethodConfig(method)
	if err != nil {
		return valueobject.NewZeroMoney(amount.Currency()), err
	}
	
	totalFee := pc.fixedFee
	
	// Add percentage fee
	if pc.percentageFee > 0 {
		percentageFeeAmount, err := amount.Multiply(pc.percentageFee / 100)
		if err != nil {
			return valueobject.NewZeroMoney(amount.Currency()), fmt.Errorf("failed to calculate percentage fee: %w", err)
		}
		
		totalFee, err = totalFee.Add(percentageFeeAmount)
		if err != nil {
			return valueobject.NewZeroMoney(amount.Currency()), fmt.Errorf("failed to add percentage fee: %w", err)
		}
	}
	
	// Add method-specific fee
	methodFee, err := methodConfig.CalculateFee(amount)
	if err != nil {
		return valueobject.NewZeroMoney(amount.Currency()), fmt.Errorf("failed to calculate method fee: %w", err)
	}
	
	totalFee, err = totalFee.Add(methodFee)
	if err != nil {
		return valueobject.NewZeroMoney(amount.Currency()), fmt.Errorf("failed to add method fee: %w", err)
	}
	
	return totalFee, nil
}

// SoftDelete marks the payment config as deleted
func (pc *PaymentConfig) SoftDelete() error {
	if pc.IsDeleted() {
		return fmt.Errorf("payment config is already deleted")
	}
	
	now := time.Now()
	pc.deletedAt = &now
	pc.updatedAt = now
	
	// Add domain event
	event := event.NewPaymentConfigDeletedEvent(
		pc.gateway,
		pc.name,
		now,
	)
	pc.AddDomainEvent(event)
	
	return nil
}