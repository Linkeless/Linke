package valueobject

import "fmt"

// PaymentGateway represents a payment gateway
type PaymentGateway struct {
	value string
}

// Payment gateway constants
const (
	PaymentGatewayStripe  = "stripe"
	PaymentGatewayEpay    = "epay"
	PaymentGatewayEPUSDT  = "epusdt"
	PaymentGatewayPayPal  = "paypal"
	PaymentGatewayAlipay  = "alipay"
	PaymentGatewayWechat  = "wechat"
	PaymentGatewayCoinbase = "coinbase"
	PaymentGatewayBinance = "binance"
)

var validPaymentGateways = map[string]bool{
	PaymentGatewayStripe:   true,
	PaymentGatewayEpay:     true,
	PaymentGatewayEPUSDT:   true,
	PaymentGatewayPayPal:   true,
	PaymentGatewayAlipay:   true,
	PaymentGatewayWechat:   true,
	PaymentGatewayCoinbase: true,
	PaymentGatewayBinance:  true,
}

// Gateway types
var gatewayTypes = map[string]string{
	PaymentGatewayStripe:   "international",
	PaymentGatewayEpay:     "china",
	PaymentGatewayEPUSDT:   "crypto",
	PaymentGatewayPayPal:   "international",
	PaymentGatewayAlipay:   "china",
	PaymentGatewayWechat:   "china",
	PaymentGatewayCoinbase: "crypto",
	PaymentGatewayBinance:  "crypto",
}

// NewPaymentGateway creates a new PaymentGateway with validation
func NewPaymentGateway(value string) (PaymentGateway, error) {
	if value == "" {
		return PaymentGateway{}, fmt.Errorf("payment gateway cannot be empty")
	}
	
	if !validPaymentGateways[value] {
		return PaymentGateway{}, fmt.Errorf("invalid payment gateway: %s", value)
	}
	
	return PaymentGateway{value: value}, nil
}

// NewStripePaymentGateway creates a Stripe payment gateway
func NewStripePaymentGateway() PaymentGateway {
	gateway, _ := NewPaymentGateway(PaymentGatewayStripe)
	return gateway
}

// NewEpayPaymentGateway creates an Epay payment gateway
func NewEpayPaymentGateway() PaymentGateway {
	gateway, _ := NewPaymentGateway(PaymentGatewayEpay)
	return gateway
}

// NewEPUSDTPaymentGateway creates an EPUSDT payment gateway
func NewEPUSDTPaymentGateway() PaymentGateway {
	gateway, _ := NewPaymentGateway(PaymentGatewayEPUSDT)
	return gateway
}

// Value returns the underlying string value
func (pg PaymentGateway) Value() string {
	return pg.value
}

// String returns string representation
func (pg PaymentGateway) String() string {
	return pg.value
}

// Equals checks if two PaymentGateways are equal
func (pg PaymentGateway) Equals(other PaymentGateway) bool {
	return pg.value == other.value
}

// IsEmpty checks if the payment gateway is empty
func (pg PaymentGateway) IsEmpty() bool {
	return pg.value == ""
}

// GetType returns the type of the payment gateway
func (pg PaymentGateway) GetType() string {
	if gatewayType, exists := gatewayTypes[pg.value]; exists {
		return gatewayType
	}
	return "unknown"
}

// IsInternational checks if this is an international payment gateway
func (pg PaymentGateway) IsInternational() bool {
	return pg.GetType() == "international"
}

// IsChina checks if this is a China-specific payment gateway
func (pg PaymentGateway) IsChina() bool {
	return pg.GetType() == "china"
}

// IsCrypto checks if this is a cryptocurrency payment gateway
func (pg PaymentGateway) IsCrypto() bool {
	return pg.GetType() == "crypto"
}

// GetDisplayName returns a human-readable display name
func (pg PaymentGateway) GetDisplayName() string {
	displayNames := map[string]string{
		PaymentGatewayStripe:   "Stripe",
		PaymentGatewayEpay:     "易支付",
		PaymentGatewayEPUSDT:   "EPUSDT",
		PaymentGatewayPayPal:   "PayPal",
		PaymentGatewayAlipay:   "支付宝",
		PaymentGatewayWechat:   "微信支付",
		PaymentGatewayCoinbase: "Coinbase",
		PaymentGatewayBinance:  "Binance",
	}
	
	if displayName, exists := displayNames[pg.value]; exists {
		return displayName
	}
	
	return pg.value
}

// GetSupportedMethods returns payment methods supported by this gateway
func (pg PaymentGateway) GetSupportedMethods() []PaymentMethod {
	supportedMethods := map[string][]string{
		PaymentGatewayStripe: {
			PaymentMethodCreditCard,
			PaymentMethodDebitCard,
			PaymentMethodApplePay,
			PaymentMethodGooglePay,
		},
		PaymentGatewayEpay: {
			PaymentMethodAlipay,
			PaymentMethodWechat,
			PaymentMethodQQPay,
			PaymentMethodUnionPay,
		},
		PaymentGatewayEPUSDT: {
			PaymentMethodUSDT,
		},
		PaymentGatewayPayPal: {
			PaymentMethodPayPal,
		},
		PaymentGatewayAlipay: {
			PaymentMethodAlipay,
		},
		PaymentGatewayWechat: {
			PaymentMethodWechat,
		},
		PaymentGatewayCoinbase: {
			PaymentMethodBTC,
			PaymentMethodETH,
			PaymentMethodUSDT,
		},
		PaymentGatewayBinance: {
			PaymentMethodBTC,
			PaymentMethodETH,
			PaymentMethodUSDT,
		},
	}
	
	methodStrings, exists := supportedMethods[pg.value]
	if !exists {
		return []PaymentMethod{}
	}
	
	var methods []PaymentMethod
	for _, methodString := range methodStrings {
		if method, err := NewPaymentMethod(methodString); err == nil {
			methods = append(methods, method)
		}
	}
	
	return methods
}

// SupportsCurrency checks if the gateway supports a specific currency
func (pg PaymentGateway) SupportsCurrency(currency Currency) bool {
	supportedCurrencies := map[string][]string{
		PaymentGatewayStripe: {
			CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyJPY,
		},
		PaymentGatewayEpay: {
			CurrencyCNY,
		},
		PaymentGatewayEPUSDT: {
			CurrencyUSDT,
		},
		PaymentGatewayPayPal: {
			CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyCNY,
		},
		PaymentGatewayAlipay: {
			CurrencyCNY, CurrencyUSD,
		},
		PaymentGatewayWechat: {
			CurrencyCNY,
		},
		PaymentGatewayCoinbase: {
			CurrencyBTC, CurrencyETH, CurrencyUSDT, CurrencyUSD,
		},
		PaymentGatewayBinance: {
			CurrencyBTC, CurrencyETH, CurrencyUSDT, CurrencyUSD,
		},
	}
	
	currencies, exists := supportedCurrencies[pg.value]
	if !exists {
		return false
	}
	
	for _, supportedCurrency := range currencies {
		if currency.Code() == supportedCurrency {
			return true
		}
	}
	
	return false
}

// RequiresWebhook checks if the gateway requires webhook configuration
func (pg PaymentGateway) RequiresWebhook() bool {
	webhookRequired := map[string]bool{
		PaymentGatewayStripe:   true,
		PaymentGatewayEpay:     true,
		PaymentGatewayEPUSDT:   true,
		PaymentGatewayPayPal:   true,
		PaymentGatewayAlipay:   true,
		PaymentGatewayWechat:   true,
		PaymentGatewayCoinbase: true,
		PaymentGatewayBinance:  true,
	}
	
	return webhookRequired[pg.value]
}