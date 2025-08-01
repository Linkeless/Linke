package valueobject

import "fmt"

// PaymentMethod represents a payment method
type PaymentMethod struct {
	value string
}

// Payment method constants
const (
	PaymentMethodCreditCard = "credit_card"
	PaymentMethodDebitCard  = "debit_card"
	PaymentMethodAlipay     = "alipay"
	PaymentMethodWechat     = "wechat"
	PaymentMethodQQPay      = "qqpay"
	PaymentMethodUnionPay   = "unionpay"
	PaymentMethodUSDT       = "usdt"
	PaymentMethodBTC        = "btc"
	PaymentMethodETH        = "eth"
	PaymentMethodPayPal     = "paypal"
	PaymentMethodBankWire   = "bank_wire"
	PaymentMethodApplePay   = "apple_pay"
	PaymentMethodGooglePay  = "google_pay"
)

var validPaymentMethods = map[string]bool{
	PaymentMethodCreditCard: true,
	PaymentMethodDebitCard:  true,
	PaymentMethodAlipay:     true,
	PaymentMethodWechat:     true,
	PaymentMethodQQPay:      true,
	PaymentMethodUnionPay:   true,
	PaymentMethodUSDT:       true,
	PaymentMethodBTC:        true,
	PaymentMethodETH:        true,
	PaymentMethodPayPal:     true,
	PaymentMethodBankWire:   true,
	PaymentMethodApplePay:   true,
	PaymentMethodGooglePay:  true,
}

// Payment method categories
var paymentMethodCategories = map[string]string{
	PaymentMethodCreditCard: "card",
	PaymentMethodDebitCard:  "card",
	PaymentMethodAlipay:     "digital_wallet",
	PaymentMethodWechat:     "digital_wallet",
	PaymentMethodQQPay:      "digital_wallet",
	PaymentMethodUnionPay:   "card",
	PaymentMethodUSDT:       "crypto",
	PaymentMethodBTC:        "crypto",
	PaymentMethodETH:        "crypto",
	PaymentMethodPayPal:     "digital_wallet",
	PaymentMethodBankWire:   "bank_transfer",
	PaymentMethodApplePay:   "mobile_payment",
	PaymentMethodGooglePay:  "mobile_payment",
}

// NewPaymentMethod creates a new PaymentMethod with validation
func NewPaymentMethod(value string) (PaymentMethod, error) {
	if value == "" {
		return PaymentMethod{}, fmt.Errorf("payment method cannot be empty")
	}
	
	if !validPaymentMethods[value] {
		return PaymentMethod{}, fmt.Errorf("invalid payment method: %s", value)
	}
	
	return PaymentMethod{value: value}, nil
}

// NewCreditCardPaymentMethod creates a credit card payment method
func NewCreditCardPaymentMethod() PaymentMethod {
	method, _ := NewPaymentMethod(PaymentMethodCreditCard)
	return method
}

// NewAlipayPaymentMethod creates an Alipay payment method
func NewAlipayPaymentMethod() PaymentMethod {
	method, _ := NewPaymentMethod(PaymentMethodAlipay)
	return method
}

// NewWechatPaymentMethod creates a WeChat payment method
func NewWechatPaymentMethod() PaymentMethod {
	method, _ := NewPaymentMethod(PaymentMethodWechat)
	return method
}

// NewUSDTPaymentMethod creates a USDT payment method
func NewUSDTPaymentMethod() PaymentMethod {
	method, _ := NewPaymentMethod(PaymentMethodUSDT)
	return method
}

// Value returns the underlying string value
func (pm PaymentMethod) Value() string {
	return pm.value
}

// String returns string representation
func (pm PaymentMethod) String() string {
	return pm.value
}

// Equals checks if two PaymentMethods are equal
func (pm PaymentMethod) Equals(other PaymentMethod) bool {
	return pm.value == other.value
}

// IsEmpty checks if the payment method is empty
func (pm PaymentMethod) IsEmpty() bool {
	return pm.value == ""
}

// GetCategory returns the category of the payment method
func (pm PaymentMethod) GetCategory() string {
	if category, exists := paymentMethodCategories[pm.value]; exists {
		return category
	}
	return "unknown"
}

// IsCard checks if this is a card payment method
func (pm PaymentMethod) IsCard() bool {
	return pm.GetCategory() == "card"
}

// IsDigitalWallet checks if this is a digital wallet payment method
func (pm PaymentMethod) IsDigitalWallet() bool {
	return pm.GetCategory() == "digital_wallet"
}

// IsCrypto checks if this is a cryptocurrency payment method
func (pm PaymentMethod) IsCrypto() bool {
	return pm.GetCategory() == "crypto"
}

// IsMobilePayment checks if this is a mobile payment method
func (pm PaymentMethod) IsMobilePayment() bool {
	return pm.GetCategory() == "mobile_payment"
}

// IsBankTransfer checks if this is a bank transfer payment method
func (pm PaymentMethod) IsBankTransfer() bool {
	return pm.GetCategory() == "bank_transfer"
}

// GetDisplayName returns a human-readable display name
func (pm PaymentMethod) GetDisplayName() string {
	displayNames := map[string]string{
		PaymentMethodCreditCard: "Credit Card",
		PaymentMethodDebitCard:  "Debit Card",
		PaymentMethodAlipay:     "Alipay",
		PaymentMethodWechat:     "WeChat Pay",
		PaymentMethodQQPay:      "QQ Pay",
		PaymentMethodUnionPay:   "UnionPay",
		PaymentMethodUSDT:       "USDT",
		PaymentMethodBTC:        "Bitcoin",
		PaymentMethodETH:        "Ethereum",
		PaymentMethodPayPal:     "PayPal",
		PaymentMethodBankWire:   "Bank Wire",
		PaymentMethodApplePay:   "Apple Pay",
		PaymentMethodGooglePay:  "Google Pay",
	}
	
	if displayName, exists := displayNames[pm.value]; exists {
		return displayName
	}
	
	return pm.value
}

// RequiresKYC checks if this payment method requires KYC verification
func (pm PaymentMethod) RequiresKYC() bool {
	kycRequiredMethods := map[string]bool{
		PaymentMethodBankWire: true,
		PaymentMethodBTC:      true,
		PaymentMethodETH:      true,
		PaymentMethodUSDT:     true,
	}
	
	return kycRequiredMethods[pm.value]
}

// GetProcessingTime returns typical processing time in minutes
func (pm PaymentMethod) GetProcessingTime() int {
	processingTimes := map[string]int{
		PaymentMethodCreditCard: 1,  // 1 minute
		PaymentMethodDebitCard:  1,  // 1 minute
		PaymentMethodAlipay:     1,  // 1 minute
		PaymentMethodWechat:     1,  // 1 minute
		PaymentMethodQQPay:      1,  // 1 minute
		PaymentMethodUnionPay:   5,  // 5 minutes
		PaymentMethodUSDT:       10, // 10 minutes
		PaymentMethodBTC:        60, // 60 minutes
		PaymentMethodETH:        15, // 15 minutes
		PaymentMethodPayPal:     5,  // 5 minutes
		PaymentMethodBankWire:   1440, // 24 hours
		PaymentMethodApplePay:   1,  // 1 minute
		PaymentMethodGooglePay:  1,  // 1 minute
	}
	
	if time, exists := processingTimes[pm.value]; exists {
		return time
	}
	
	return 5 // Default to 5 minutes
}