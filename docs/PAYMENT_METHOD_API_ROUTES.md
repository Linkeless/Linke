# Payment Method Management API Routes

This document outlines the API routes for the Self-Service Payment Method Management feature implemented in the Linke platform.

## Integration Points

### Main Server Configuration
The payment method handler should be added to the `NewHTTPServer` function in `/cmd/server/main.go`:

```go
func NewHTTPServer(
    // ... existing parameters
    paymentHandler *paymentHandlers.PaymentHandler,
    paymentMethodHandler *paymentHandlers.PaymentMethodHandler, // Add this
    // ... other parameters
) *HTTPServer {
```

### API Routes
The following routes should be added to the payment group in the main server:

```go
// 支付方法路由 (/api/v1/payment-methods)
paymentMethodGroup := apiV1.Group("/payment-methods")
{
    // List user payment methods
    paymentMethodGroup.GET("", paymentMethodHandler.ListPaymentMethods)
    
    // Get default payment method
    paymentMethodGroup.GET("/default", paymentMethodHandler.GetDefaultPaymentMethod)
    
    // Create new payment method
    paymentMethodGroup.POST("", paymentMethodHandler.CreatePaymentMethod)
    
    // Get specific payment method
    paymentMethodGroup.GET("/:id", paymentMethodHandler.GetPaymentMethod)
    
    // Update payment method
    paymentMethodGroup.PUT("/:id", paymentMethodHandler.UpdatePaymentMethod)
    
    // Delete payment method
    paymentMethodGroup.DELETE("/:id", paymentMethodHandler.DeletePaymentMethod)
    
    // Set payment method as default
    paymentMethodGroup.PUT("/:id/default", paymentMethodHandler.SetDefaultPaymentMethod)
    
    // Validate payment method
    paymentMethodGroup.POST("/:id/validate", paymentMethodHandler.ValidatePaymentMethod)
    
    // Get payment method usage statistics
    paymentMethodGroup.GET("/:id/stats", paymentMethodHandler.GetPaymentMethodUsageStats)
}
```

## API Endpoints Documentation

### 1. List Payment Methods
- **Endpoint**: `GET /api/v1/payment-methods`
- **Query Parameters**:
  - `gateway` (string, optional): Filter by payment gateway
  - `active_only` (boolean, optional): Show only active payment methods
- **Response**: List of user's payment methods with default method information

### 2. Get Default Payment Method
- **Endpoint**: `GET /api/v1/payment-methods/default`
- **Query Parameters**:
  - `gateway` (string, optional): Filter by payment gateway
- **Response**: User's default payment method

### 3. Create Payment Method
- **Endpoint**: `POST /api/v1/payment-methods`
- **Security**: Rate limited for security
- **Request Body**: Payment method creation data (tokenized)
- **Response**: Created payment method information

### 4. Get Payment Method
- **Endpoint**: `GET /api/v1/payment-methods/{id}`
- **Security**: User ownership validation
- **Response**: Payment method details

### 5. Update Payment Method
- **Endpoint**: `PUT /api/v1/payment-methods/{id}`
- **Security**: User ownership validation
- **Request Body**: Payment method update data
- **Response**: Updated payment method information

### 6. Delete Payment Method
- **Endpoint**: `DELETE /api/v1/payment-methods/{id}`
- **Security**: User ownership validation
- **Response**: Success confirmation

### 7. Set Default Payment Method
- **Endpoint**: `PUT /api/v1/payment-methods/{id}/default`
- **Security**: User ownership validation
- **Response**: Updated payment method with default status

### 8. Validate Payment Method
- **Endpoint**: `POST /api/v1/payment-methods/{id}/validate`
- **Security**: Rate limited, user ownership validation
- **Response**: Validation result and updated payment method

### 9. Get Usage Statistics
- **Endpoint**: `GET /api/v1/payment-methods/{id}/stats`
- **Security**: User ownership validation
- **Response**: Usage statistics for the payment method

## Security Features Implemented

1. **PCI Compliance**: No raw payment data stored, only tokenized data
2. **User Ownership Validation**: All operations validate user ownership
3. **Rate Limiting**: Sensitive operations (create, validate) are rate limited
4. **Audit Logging**: All operations are logged for security tracking
5. **Error Masking**: Sensitive errors are masked in API responses

## Integration with Subscription Orders

The subscription order creation has been enhanced to support:

1. **Default Payment Method**: Use `use_default_payment: true`
2. **Specific Payment Method**: Use `payment_method_id: <id>`
3. **Traditional Method**: Specify `payment_gateway` and `payment_method`

Example subscription order request with saved payment method:
```json
{
  "subscription_plan_id": 1,
  "order_type": "new",
  "use_default_payment": true,
  "coupon_code": "SAVE20"
}
```

## Database Migration

Run the migration to create the payment_methods table:
```bash
make migrate-up
```

The migration file `000018_create_payment_methods_table.up.sql` includes:
- Table structure with security-focused design
- Proper indexes for performance
- Constraints to ensure data integrity
- Triggers for automatic field updates
- Functions to enforce business rules

## Testing

Basic tests have been implemented for:
- Payment method entity validation
- Service constants and configuration
- Request/response structures

For comprehensive testing, run:
```bash
make test
```

## Future Enhancements

1. **Advanced Analytics**: Payment method performance analytics
2. **Smart Recommendations**: AI-powered payment method suggestions
3. **Fraud Detection**: Advanced fraud detection for payment methods
4. **International Support**: Multi-currency and regional payment methods
5. **Subscription Optimization**: Auto-retry with different payment methods