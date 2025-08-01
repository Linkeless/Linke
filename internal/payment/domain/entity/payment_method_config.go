package entity

import (
	"fmt"

	"linke/internal/payment/domain/valueobject"
)

// FeeType represents the type of fee calculation
type FeeType string

const (
	FeeTypeNone       FeeType = "none"
	FeeTypeFixed      FeeType = "fixed"
	FeeTypePercentage FeeType = "percentage"
)

// Environment represents the payment environment
type Environment string

const (
	EnvironmentProduction Environment = "production"
	EnvironmentSandbox    Environment = "sandbox"
	EnvironmentTest       Environment = "test"
)

// PaymentMethodConfig represents the configuration for a payment method within a gateway
type PaymentMethodConfig struct {
	method      valueobject.PaymentMethod
	name        string
	description string
	icon        string
	isEnabled   bool
	sortOrder   int
	feeType     FeeType
	feeValue    float64
	feeMin      float64
	feeMax      float64
	environment Environment
}

// NewPaymentMethodConfig creates a new PaymentMethodConfig
func NewPaymentMethodConfig(
	method valueobject.PaymentMethod,
	name string,
	description string,
	icon string,
	isEnabled bool,
	sortOrder int,
	feeType FeeType,
	feeValue float64,
	feeMin float64,
	feeMax float64,
	environment Environment,
) PaymentMethodConfig {
	return PaymentMethodConfig{
		method:      method,
		name:        name,
		description: description,
		icon:        icon,
		isEnabled:   isEnabled,
		sortOrder:   sortOrder,
		feeType:     feeType,
		feeValue:    feeValue,
		feeMin:      feeMin,
		feeMax:      feeMax,
		environment: environment,
	}
}

// Getters
func (pmc PaymentMethodConfig) Method() valueobject.PaymentMethod {
	return pmc.method
}

func (pmc PaymentMethodConfig) Name() string {
	return pmc.name
}

func (pmc PaymentMethodConfig) Description() string {
	return pmc.description
}

func (pmc PaymentMethodConfig) Icon() string {
	return pmc.icon
}

func (pmc PaymentMethodConfig) IsEnabled() bool {
	return pmc.isEnabled
}

func (pmc PaymentMethodConfig) SortOrder() int {
	return pmc.sortOrder
}

func (pmc PaymentMethodConfig) FeeType() FeeType {
	return pmc.feeType
}

func (pmc PaymentMethodConfig) FeeValue() float64 {
	return pmc.feeValue
}

func (pmc PaymentMethodConfig) FeeMin() float64 {
	return pmc.feeMin
}

func (pmc PaymentMethodConfig) FeeMax() float64 {
	return pmc.feeMax
}

func (pmc PaymentMethodConfig) Environment() Environment {
	return pmc.environment
}

// Business methods

// UpdateName updates the method name
func (pmc *PaymentMethodConfig) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("payment method name cannot be empty")
	}
	pmc.name = name
	return nil
}

// UpdateDescription updates the method description
func (pmc *PaymentMethodConfig) UpdateDescription(description string) {
	pmc.description = description
}

// UpdateIcon updates the method icon
func (pmc *PaymentMethodConfig) UpdateIcon(icon string) {
	pmc.icon = icon
}

// Enable enables the payment method
func (pmc *PaymentMethodConfig) Enable() {
	pmc.isEnabled = true
}

// Disable disables the payment method
func (pmc *PaymentMethodConfig) Disable() {
	pmc.isEnabled = false
}

// UpdateSortOrder updates the sort order
func (pmc *PaymentMethodConfig) UpdateSortOrder(sortOrder int) error {
	if sortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}
	pmc.sortOrder = sortOrder
	return nil
}

// UpdateFeeSettings updates the fee configuration
func (pmc *PaymentMethodConfig) UpdateFeeSettings(feeType FeeType, feeValue, feeMin, feeMax float64) error {
	if feeType == FeeTypePercentage && (feeValue < 0 || feeValue > 100) {
		return fmt.Errorf("percentage fee must be between 0 and 100")
	}
	
	if feeType == FeeTypeFixed && feeValue < 0 {
		return fmt.Errorf("fixed fee cannot be negative")
	}
	
	if feeMin < 0 || feeMax < 0 {
		return fmt.Errorf("fee min and max cannot be negative")
	}
	
	if feeMin > feeMax && feeMax > 0 {
		return fmt.Errorf("fee min cannot be greater than fee max")
	}
	
	pmc.feeType = feeType
	pmc.feeValue = feeValue
	pmc.feeMin = feeMin
	pmc.feeMax = feeMax
	
	return nil
}

// UpdateEnvironment updates the environment
func (pmc *PaymentMethodConfig) UpdateEnvironment(environment Environment) error {
	validEnvironments := map[Environment]bool{
		EnvironmentProduction: true,
		EnvironmentSandbox:    true,
		EnvironmentTest:       true,
	}
	
	if !validEnvironments[environment] {
		return fmt.Errorf("invalid environment: %s", environment)
	}
	
	pmc.environment = environment
	return nil
}

// Business queries

// CalculateFee calculates the fee for a given amount
func (pmc PaymentMethodConfig) CalculateFee(amount valueobject.Money) (valueobject.Money, error) {
	switch pmc.feeType {
	case FeeTypeNone:
		return valueobject.NewZeroMoney(amount.Currency()), nil
		
	case FeeTypeFixed:
		fee, err := valueobject.NewMoney(pmc.feeValue, amount.Currency())
		if err != nil {
			return valueobject.NewZeroMoney(amount.Currency()), err
		}
		
		// Apply min/max limits if set
		if pmc.feeMin > 0 {
			minFee, err := valueobject.NewMoney(pmc.feeMin, amount.Currency())
			if err == nil {
				if less, err := fee.IsLessThan(minFee); err == nil && less {
					fee = minFee
				}
			}
		}
		
		if pmc.feeMax > 0 {
			maxFee, err := valueobject.NewMoney(pmc.feeMax, amount.Currency())
			if err == nil {
				if greater, err := fee.IsGreaterThan(maxFee); err == nil && greater {
					fee = maxFee
				}
			}
		}
		
		return fee, nil
		
	case FeeTypePercentage:
		fee, err := amount.Multiply(pmc.feeValue / 100)
		if err != nil {
			return valueobject.NewZeroMoney(amount.Currency()), err
		}
		
		// Apply min/max limits if set
		if pmc.feeMin > 0 {
			minFee, err := valueobject.NewMoney(pmc.feeMin, amount.Currency())
			if err == nil {
				if less, err := fee.IsLessThan(minFee); err == nil && less {
					fee = minFee
				}
			}
		}
		
		if pmc.feeMax > 0 {
			maxFee, err := valueobject.NewMoney(pmc.feeMax, amount.Currency())
			if err == nil {
				if greater, err := fee.IsGreaterThan(maxFee); err == nil && greater {
					fee = maxFee
				}
			}
		}
		
		return fee, nil
		
	default:
		return valueobject.NewZeroMoney(amount.Currency()), fmt.Errorf("unsupported fee type: %s", pmc.feeType)
	}
}

// IsProduction checks if the method is in production environment
func (pmc PaymentMethodConfig) IsProduction() bool {
	return pmc.environment == EnvironmentProduction
}

// IsSandbox checks if the method is in sandbox environment
func (pmc PaymentMethodConfig) IsSandbox() bool {
	return pmc.environment == EnvironmentSandbox
}

// IsTest checks if the method is in test environment
func (pmc PaymentMethodConfig) IsTest() bool {
	return pmc.environment == EnvironmentTest
}

// HasFee checks if the method has any fee configured
func (pmc PaymentMethodConfig) HasFee() bool {
	return pmc.feeType != FeeTypeNone && pmc.feeValue > 0
}

// Equals checks if two PaymentMethodConfigs are equal
func (pmc PaymentMethodConfig) Equals(other PaymentMethodConfig) bool {
	return pmc.method.Equals(other.method) &&
		pmc.name == other.name &&
		pmc.description == other.description &&
		pmc.icon == other.icon &&
		pmc.isEnabled == other.isEnabled &&
		pmc.sortOrder == other.sortOrder &&
		pmc.feeType == other.feeType &&
		pmc.feeValue == other.feeValue &&
		pmc.feeMin == other.feeMin &&
		pmc.feeMax == other.feeMax &&
		pmc.environment == other.environment
}