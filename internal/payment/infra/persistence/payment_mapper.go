package persistence

import (
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/entity"
	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// PaymentMapper handles conversion between Payment aggregate and PaymentPO
type PaymentMapper struct{}

// NewPaymentMapper creates a new PaymentMapper
func NewPaymentMapper() *PaymentMapper {
	return &PaymentMapper{}
}

// ToAggregate converts PaymentPO to Payment aggregate
func (m *PaymentMapper) ToAggregate(po *PaymentPO) (*aggregate.Payment, error) {
	// Convert IDs
	paymentID, err := valueobject.NewPaymentID(po.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	paymentNumber, err := valueobject.NewPaymentNumber(po.PaymentNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid payment number: %w", err)
	}

	// Create shared value objects directly
	invoiceID, err := sharedvo.NewInvoiceID(po.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	userID, err := sharedvo.NewUserIDFromUint(po.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Convert currency and amounts
	currency, err := valueobject.NewCurrency(po.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}
	
	// Convert to shared currency for amounts
	sharedCurrency, err := valueobject.ConvertToSharedCurrency(currency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}

	amount, err := sharedvo.NewMoney(po.Amount, sharedCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	gatewayFee, err := sharedvo.NewMoney(po.GatewayFee, sharedCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway fee: %w", err)
	}

	refundAmount, err := sharedvo.NewMoney(po.RefundAmount, sharedCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid refund amount: %w", err)
	}

	// Convert status
	status, err := valueobject.NewPaymentStatus(po.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid payment status: %w", err)
	}

	// Convert payment method
	paymentMethod, err := valueobject.NewPaymentMethod(po.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}

	// Convert payment gateway
	paymentGateway, err := valueobject.NewPaymentGateway(po.PaymentGateway)
	if err != nil {
		return nil, fmt.Errorf("invalid payment gateway: %w", err)
	}

	// Convert refund reference if exists
	var refundReference valueobject.PaymentNumber
	if po.RefundReference != "" {
		refundReference, err = valueobject.NewPaymentNumber(po.RefundReference)
		if err != nil {
			return nil, fmt.Errorf("invalid refund reference: %w", err)
		}
	}

	// Load aggregate using factory method
	payment := aggregate.LoadPayment(
		paymentID,
		paymentNumber,
		invoiceID,
		userID,
		amount,
		status,
		paymentMethod,
		paymentGateway,
		po.PaymentIntentID,
		po.GatewayTransactionID,
		gatewayFee,
		po.PaymentURL,
		po.QRCodeURL,
		po.RedirectURL,
		po.ExpiresAt,
		po.ProcessedAt,
		po.CompletedAt,
		refundAmount,
		po.RefundedAt,
		po.RefundReason,
		refundReference,
		po.WebhookData,
		po.NotificationCount,
		po.LastNotificationAt,
		po.Notes,
		po.Metadata,
		po.CreatedAt,
		po.UpdatedAt,
		func() *time.Time {
			if po.DeletedAt.Valid {
				return &po.DeletedAt.Time
			}
			return nil
		}(),
	)

	return payment, nil
}

// ToPersistentObject converts Payment aggregate to PaymentPO
func (m *PaymentMapper) ToPersistentObject(payment *aggregate.Payment) (*PaymentPO, error) {
	po := &PaymentPO{
		ID:                   payment.ID().Value(),
		InvoiceID:            payment.InvoiceID().Value(),
		UserID:               payment.UserID().ToUint(),
		PaymentNumber:        payment.PaymentNumber().Value(),
		PaymentIntentID:      payment.PaymentIntentID(),
		Status:               payment.Status().Value(),
		Amount:               payment.Amount().Amount(),
		Currency:             payment.Amount().Currency().Code(),
		PaymentMethod:        payment.PaymentMethod().Value(),
		PaymentGateway:       payment.PaymentGateway().Value(),
		GatewayTransactionID: payment.GatewayTransactionID(),
		GatewayFee:          payment.GatewayFee().Amount(),
		PaymentURL:          payment.PaymentURL(),
		QRCodeURL:           payment.QRCodeURL(),
		RedirectURL:         payment.RedirectURL(),
		ExpiresAt:           payment.ExpiresAt(),
		ProcessedAt:         payment.ProcessedAt(),
		CompletedAt:         payment.CompletedAt(),
		RefundAmount:        payment.RefundAmount().Amount(),
		RefundedAt:          payment.RefundedAt(),
		RefundReason:        payment.RefundReason(),
		RefundReference:     payment.RefundReference().Value(),
		WebhookData:         payment.WebhookData(),
		NotificationCount:   payment.NotificationCount(),
		LastNotificationAt:  payment.LastNotificationAt(),
		Notes:               payment.Notes(),
		Metadata:            payment.Metadata(),
		CreatedAt:           payment.CreatedAt(),
		UpdatedAt:           payment.UpdatedAt(),
	}

	// Handle soft delete
	if payment.DeletedAt() != nil {
		po.DeletedAt.Time = *payment.DeletedAt()
		po.DeletedAt.Valid = true
	}

	return po, nil
}

// ToAggregateList converts slice of PaymentPO to slice of Payment aggregates
func (m *PaymentMapper) ToAggregateList(pos []*PaymentPO) ([]*aggregate.Payment, error) {
	payments := make([]*aggregate.Payment, 0, len(pos))
	
	for _, po := range pos {
		payment, err := m.ToAggregate(po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert payment PO to aggregate: %w", err)
		}
		payments = append(payments, payment)
	}
	
	return payments, nil
}

// ToPersistentObjectList converts slice of Payment aggregates to slice of PaymentPO
func (m *PaymentMapper) ToPersistentObjectList(payments []*aggregate.Payment) ([]*PaymentPO, error) {
	pos := make([]*PaymentPO, 0, len(payments))
	
	for _, payment := range payments {
		po, err := m.ToPersistentObject(payment)
		if err != nil {
			return nil, fmt.Errorf("failed to convert payment aggregate to PO: %w", err)
		}
		pos = append(pos, po)
	}
	
	return pos, nil
}

// PaymentConfigMapper handles conversion between PaymentConfig aggregate and PaymentConfigPO
type PaymentConfigMapper struct{}

// NewPaymentConfigMapper creates a new PaymentConfigMapper
func NewPaymentConfigMapper() *PaymentConfigMapper {
	return &PaymentConfigMapper{}
}

// ToAggregate converts PaymentConfigPO to PaymentConfig aggregate
func (m *PaymentConfigMapper) ToAggregate(po *PaymentConfigPO) (*aggregate.PaymentConfig, error) {
	// Convert ID
	configID, err := valueobject.NewPaymentConfigID(po.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment config ID: %w", err)
	}

	// Convert gateway
	gateway, err := valueobject.NewPaymentGateway(po.Gateway)
	if err != nil {
		return nil, fmt.Errorf("invalid payment gateway: %w", err)
	}

	// Parse supported currencies
	var currencyStrings []string
	if err := json.Unmarshal([]byte(po.SupportedCurrencies), &currencyStrings); err != nil {
		return nil, fmt.Errorf("invalid supported currencies JSON: %w", err)
	}

	supportedCurrencies := make([]valueobject.Currency, 0, len(currencyStrings))
	for _, currencyStr := range currencyStrings {
		currency, err := valueobject.NewCurrency(currencyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid currency %s: %w", currencyStr, err)
		}
		supportedCurrencies = append(supportedCurrencies, currency)
	}

	// Parse supported methods
	var methodData []map[string]interface{}
	if po.SupportedMethods != "" {
		if err := json.Unmarshal([]byte(po.SupportedMethods), &methodData); err != nil {
			return nil, fmt.Errorf("invalid supported methods JSON: %w", err)
		}
	}

	supportedMethods := make([]entity.PaymentMethodConfig, 0, len(methodData))
	for _, methodMap := range methodData {
		methodConfig, err := m.parsePaymentMethodConfig(methodMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse payment method config: %w", err)
		}
		supportedMethods = append(supportedMethods, methodConfig)
	}

	// Assume the first currency is the base currency for amounts
	baseCurrency := supportedCurrencies[0]
	
	minAmount, err := valueobject.NewMoney(po.MinAmount, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid min amount: %w", err)
	}

	maxAmount, err := valueobject.NewMoney(po.MaxAmount, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid max amount: %w", err)
	}

	fixedFee, err := valueobject.NewMoney(po.FixedFee, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid fixed fee: %w", err)
	}

	// Load aggregate using factory method
	config := aggregate.LoadPaymentConfig(
		configID,
		gateway,
		po.Name,
		po.Config,
		po.IsEnabled,
		po.SortOrder,
		supportedCurrencies,
		supportedMethods,
		minAmount,
		maxAmount,
		fixedFee,
		po.PercentageFee,
		po.CreatedAt,
		po.UpdatedAt,
		func() *time.Time {
			if po.DeletedAt.Valid {
				return &po.DeletedAt.Time
			}
			return nil
		}(),
	)

	return config, nil
}

// parsePaymentMethodConfig parses payment method config from map
func (m *PaymentConfigMapper) parsePaymentMethodConfig(data map[string]interface{}) (entity.PaymentMethodConfig, error) {
	methodStr, ok := data["method"].(string)
	if !ok {
		return entity.PaymentMethodConfig{}, fmt.Errorf("method field is required")
	}

	method, err := valueobject.NewPaymentMethod(methodStr)
	if err != nil {
		return entity.PaymentMethodConfig{}, fmt.Errorf("invalid payment method: %w", err)
	}

	name, _ := data["name"].(string)
	description, _ := data["description"].(string)
	icon, _ := data["icon"].(string)
	isEnabled, _ := data["is_enabled"].(bool)
	sortOrder, _ := data["sort_order"].(float64)
	feeType, _ := data["fee_type"].(string)
	feeValue, _ := data["fee_value"].(float64)
	feeMin, _ := data["fee_min"].(float64)
	feeMax, _ := data["fee_max"].(float64)
	environment, _ := data["environment"].(string)

	return entity.NewPaymentMethodConfig(
		method,
		name,
		description,
		icon,
		isEnabled,
		int(sortOrder),
		entity.FeeType(feeType),
		feeValue,
		feeMin,
		feeMax,
		entity.Environment(environment),
	), nil
}

// ToPersistentObject converts PaymentConfig aggregate to PaymentConfigPO
func (m *PaymentConfigMapper) ToPersistentObject(config *aggregate.PaymentConfig) (*PaymentConfigPO, error) {
	// Convert supported currencies to JSON
	currencyStrings := make([]string, 0, len(config.SupportedCurrencies()))
	for _, currency := range config.SupportedCurrencies() {
		currencyStrings = append(currencyStrings, currency.Code())
	}

	supportedCurrenciesJSON, err := json.Marshal(currencyStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal supported currencies: %w", err)
	}

	// Convert supported methods to JSON
	methodsData := make([]map[string]interface{}, 0, len(config.SupportedMethods()))
	for _, methodConfig := range config.SupportedMethods() {
		methodData := map[string]interface{}{
			"method":      methodConfig.Method().Value(),
			"name":        methodConfig.Name(),
			"description": methodConfig.Description(),
			"icon":        methodConfig.Icon(),
			"is_enabled":  methodConfig.IsEnabled(),
			"sort_order":  methodConfig.SortOrder(),
			"fee_type":    string(methodConfig.FeeType()),
			"fee_value":   methodConfig.FeeValue(),
			"fee_min":     methodConfig.FeeMin(),
			"fee_max":     methodConfig.FeeMax(),
			"environment": string(methodConfig.Environment()),
		}
		methodsData = append(methodsData, methodData)
	}

	supportedMethodsJSON, err := json.Marshal(methodsData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal supported methods: %w", err)
	}

	po := &PaymentConfigPO{
		ID:                  config.ID().Value(),
		Gateway:             config.Gateway().Value(),
		Name:                config.Name(),
		Config:              config.Config(),
		IsEnabled:           config.IsEnabled(),
		SortOrder:           config.SortOrder(),
		SupportedCurrencies: string(supportedCurrenciesJSON),
		SupportedMethods:    string(supportedMethodsJSON),
		MinAmount:           config.MinAmount().Amount(),
		MaxAmount:           config.MaxAmount().Amount(),
		FixedFee:           config.FixedFee().Amount(),
		PercentageFee:      config.PercentageFee(),
		CreatedAt:          config.CreatedAt(),
		UpdatedAt:          config.UpdatedAt(),
	}

	// Handle soft delete
	if config.DeletedAt() != nil {
		po.DeletedAt.Time = *config.DeletedAt()
		po.DeletedAt.Valid = true
	}

	return po, nil
}

// ToAggregateList converts slice of PaymentConfigPO to slice of PaymentConfig aggregates
func (m *PaymentConfigMapper) ToAggregateList(pos []*PaymentConfigPO) ([]*aggregate.PaymentConfig, error) {
	configs := make([]*aggregate.PaymentConfig, 0, len(pos))
	
	for _, po := range pos {
		config, err := m.ToAggregate(po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert payment config PO to aggregate: %w", err)
		}
		configs = append(configs, config)
	}
	
	return configs, nil
}

// ToPersistentObjectList converts slice of PaymentConfig aggregates to slice of PaymentConfigPO
func (m *PaymentConfigMapper) ToPersistentObjectList(configs []*aggregate.PaymentConfig) ([]*PaymentConfigPO, error) {
	pos := make([]*PaymentConfigPO, 0, len(configs))
	
	for _, config := range configs {
		po, err := m.ToPersistentObject(config)
		if err != nil {
			return nil, fmt.Errorf("failed to convert payment config aggregate to PO: %w", err)
		}
		pos = append(pos, po)
	}
	
	return pos, nil
}