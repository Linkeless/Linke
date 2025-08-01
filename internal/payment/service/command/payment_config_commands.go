package command

import (
	"time"

	"linke/internal/payment/domain/entity"
	"linke/internal/payment/domain/valueobject"
)

// CreatePaymentConfigCommand represents the command to create a new payment config
type CreatePaymentConfigCommand struct {
	Gateway             string                        `json:"gateway" validate:"required"`
	Name                string                        `json:"name" validate:"required"`
	Config              string                        `json:"config" validate:"required"`
	SupportedCurrencies []string                      `json:"supported_currencies" validate:"required,min=1"`
	BaseCurrency        string                        `json:"base_currency" validate:"required"`
	SortOrder           int                           `json:"sort_order"`
	Methods             []PaymentMethodConfigRequest  `json:"methods,omitempty"`
	MinAmount           float64                       `json:"min_amount" validate:"min=0.01"`
	MaxAmount           float64                       `json:"max_amount" validate:"min=0.01"`
	FixedFee            float64                       `json:"fixed_fee" validate:"min=0"`
	PercentageFee       float64                       `json:"percentage_fee" validate:"min=0,max=100"`
}

// UpdatePaymentConfigCommand represents the command to update a payment config
type UpdatePaymentConfigCommand struct {
	ConfigID            uint                          `json:"config_id" validate:"required"`
	Name                *string                       `json:"name,omitempty"`
	Config              *string                       `json:"config,omitempty"`
	SupportedCurrencies []string                      `json:"supported_currencies,omitempty"`
	SortOrder           *int                          `json:"sort_order,omitempty"`
	Methods             []PaymentMethodConfigRequest  `json:"methods,omitempty"`
	MinAmount           *float64                      `json:"min_amount,omitempty" validate:"omitempty,min=0.01"`
	MaxAmount           *float64                      `json:"max_amount,omitempty" validate:"omitempty,min=0.01"`
	FixedFee            *float64                      `json:"fixed_fee,omitempty" validate:"omitempty,min=0"`
	PercentageFee       *float64                      `json:"percentage_fee,omitempty" validate:"omitempty,min=0,max=100"`
}

// EnablePaymentConfigCommand represents the command to enable a payment config
type EnablePaymentConfigCommand struct {
	ConfigID uint `json:"config_id" validate:"required"`
}

// DisablePaymentConfigCommand represents the command to disable a payment config
type DisablePaymentConfigCommand struct {
	ConfigID uint `json:"config_id" validate:"required"`
}

// DeletePaymentConfigCommand represents the command to delete a payment config
type DeletePaymentConfigCommand struct {
	ConfigID uint `json:"config_id" validate:"required"`
}

// AddSupportedCurrencyCommand represents the command to add a supported currency
type AddSupportedCurrencyCommand struct {
	ConfigID uint   `json:"config_id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

// RemoveSupportedCurrencyCommand represents the command to remove a supported currency
type RemoveSupportedCurrencyCommand struct {
	ConfigID uint   `json:"config_id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

// AddSupportedMethodCommand represents the command to add a supported payment method
type AddSupportedMethodCommand struct {
	ConfigID     uint                         `json:"config_id" validate:"required"`
	MethodConfig PaymentMethodConfigRequest   `json:"method_config" validate:"required"`
}

// UpdateSupportedMethodCommand represents the command to update a supported payment method
type UpdateSupportedMethodCommand struct {
	ConfigID     uint                         `json:"config_id" validate:"required"`
	Method       string                       `json:"method" validate:"required"`
	MethodConfig PaymentMethodConfigRequest   `json:"method_config" validate:"required"`
}

// RemoveSupportedMethodCommand represents the command to remove a supported payment method
type RemoveSupportedMethodCommand struct {
	ConfigID uint   `json:"config_id" validate:"required"`
	Method   string `json:"method" validate:"required"`
}

// UpdateAmountLimitsCommand represents the command to update amount limits
type UpdateAmountLimitsCommand struct {
	ConfigID  uint    `json:"config_id" validate:"required"`
	MinAmount float64 `json:"min_amount" validate:"required,min=0.01"`
	MaxAmount float64 `json:"max_amount" validate:"required,min=0.01"`
	Currency  string  `json:"currency" validate:"required"`
}

// UpdateFeeSettingsCommand represents the command to update fee settings
type UpdateFeeSettingsCommand struct {
	ConfigID      uint    `json:"config_id" validate:"required"`
	FixedFee      float64 `json:"fixed_fee" validate:"min=0"`
	PercentageFee float64 `json:"percentage_fee" validate:"min=0,max=100"`
	Currency      string  `json:"currency" validate:"required"`
}

// PaymentMethodConfigRequest represents a payment method configuration request
type PaymentMethodConfigRequest struct {
	Method      string  `json:"method" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description,omitempty"`
	Icon        string  `json:"icon,omitempty"`
	IsEnabled   bool    `json:"is_enabled"`
	SortOrder   int     `json:"sort_order"`
	FeeType     string  `json:"fee_type" validate:"required,oneof=none fixed percentage"`
	FeeValue    float64 `json:"fee_value" validate:"min=0"`
	FeeMin      float64 `json:"fee_min" validate:"min=0"`
	FeeMax      float64 `json:"fee_max" validate:"min=0"`
	Environment string  `json:"environment" validate:"required,oneof=production sandbox test"`
}

// ToEntity converts PaymentMethodConfigRequest to entity
func (req PaymentMethodConfigRequest) ToEntity() (entity.PaymentMethodConfig, error) {
	method, err := valueobject.NewPaymentMethod(req.Method)
	if err != nil {
		return entity.PaymentMethodConfig{}, err
	}
	
	return entity.NewPaymentMethodConfig(
		method,
		req.Name,
		req.Description,
		req.Icon,
		req.IsEnabled,
		req.SortOrder,
		entity.FeeType(req.FeeType),
		req.FeeValue,
		req.FeeMin,
		req.FeeMax,
		entity.Environment(req.Environment),
	), nil
}

// CreatePaymentConfigResult represents the result of creating a payment config
type CreatePaymentConfigResult struct {
	ConfigID            valueobject.PaymentConfigID `json:"config_id"`
	Gateway             valueobject.PaymentGateway  `json:"gateway"`
	Name                string                      `json:"name"`
	SupportedCurrencies []valueobject.Currency      `json:"supported_currencies"`
	IsEnabled           bool                        `json:"is_enabled"`
	CreatedAt           time.Time                   `json:"created_at"`
}

// PaymentConfigCommandResult represents a generic payment config command result
type PaymentConfigCommandResult struct {
	ConfigID  valueobject.PaymentConfigID `json:"config_id"`
	Gateway   valueobject.PaymentGateway  `json:"gateway"`
	Name      string                      `json:"name"`
	IsEnabled bool                        `json:"is_enabled"`
	UpdatedAt time.Time                   `json:"updated_at"`
}