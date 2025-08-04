# Quick Purchase Feature

## Overview

The Quick Purchase feature allows users to purchase subscriptions with a single API call that handles payment creation directly. This streamlines the subscription purchasing process by reducing the number of steps required compared to the traditional order → invoice → payment flow.

## Implementation Details

### Architecture

The Quick Purchase feature follows the project's clean architecture pattern and integrates with the existing subscription, payment, and coupon domains.

### Key Components

1. **Handler**: `/internal/domains/subscription/handlers/quick_purchase_handler.go`
   - Handles HTTP requests for quick purchase
   - Validates user authentication and input
   - Extracts client IP for payment processing

2. **Service Interface**: `/internal/domains/subscription/usecases/interfaces/subscription_order_service.go`
   - Added `QuickPurchase` method to the service interface
   - Defined `QuickPurchaseRequest` and `QuickPurchaseResponse` structures

3. **Service Implementation**: `/internal/domains/subscription/usecases/implementations/subscription_order.go`
   - Implements the core quick purchase logic
   - Validates subscription plan availability
   - Applies coupon discounts if provided
   - Creates payment directly without creating order/invoice first

4. **Route Registration**: `/cmd/server/main.go`
   - Registered the quick purchase endpoint at `POST /api/v1/subscription/quick-purchase`

## API Usage

### Endpoint
```
POST /api/v1/subscription/quick-purchase
```

### Authentication
Requires Bearer token authentication.

### Request Body
```json
{
  "user_id": 1,
  "plan_id": 1,
  "payment_gateway": "epay",
  "payment_method": "alipay",
  "coupon_code": "SAVE20",
  "client_ip": "192.168.1.1",
  "return_url": "https://example.com/payment/return",
  "metadata": "{\"source\": \"web\"}"
}
```

### Response
```json
{
  "status": "success",
  "message": "Quick purchase created successfully",
  "data": {
    "payment_record": {
      "payment_no": "PAY20240101123456",
      "status": "pending",
      "amount": 29.99,
      "currency": "USD"
    },
    "payment_url": "https://payment.gateway.com/pay/123456",
    "qr_code_url": "https://payment.gateway.com/qr/123456",
    "expired_at": "2024-01-01T12:30:00Z",
    "plan_info": {
      "id": 1,
      "name": "Premium Plan",
      "price": 29.99,
      "currency": "USD",
      "billing_cycle": "monthly"
    },
    "discount_info": {
      "coupon_code": "SAVE20",
      "discount_amount": 5.99,
      "original_amount": 29.99,
      "final_amount": 24.00
    }
  }
}
```

## Flow Description

### Quick Purchase Flow
1. **Validation**: 
   - Checks for duplicate pending orders
   - Validates subscription plan availability
   - Validates coupon if provided

2. **Price Calculation**:
   - Calculates base amount and setup fees
   - Applies coupon discounts
   - Validates final amount is reasonable

3. **Payment Creation**:
   - Creates payment order directly
   - Returns payment URL to user
   - No order or invoice created at this stage

4. **Asynchronous Processing**:
   - After payment success, order and invoice are created automatically
   - Subscription is activated upon successful payment

### Security Features

- **Duplicate Order Prevention**: Checks for existing pending orders
- **Rate Limiting**: Prevents spam with 5-minute cooldown
- **Price Protection**: Validates amounts and prevents excessive discounts
- **Input Validation**: Validates all input parameters
- **Authentication**: Requires valid user authentication

### Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid input data or validation errors
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: User trying to create purchase for another user
- `500 Internal Server Error`: Server-side errors

## Testing

The implementation includes comprehensive unit tests in:
- `/internal/domains/subscription/handlers/quick_purchase_handler_test.go`

Tests cover:
- Successful quick purchase scenarios
- Authentication failures
- Invalid request data
- Service layer errors

## Benefits

1. **Improved User Experience**: Single API call for subscription purchase
2. **Reduced Latency**: Fewer round trips between client and server
3. **Simplified Integration**: Easier for frontend applications to implement
4. **Maintained Security**: All existing security validations are preserved
5. **Backwards Compatibility**: Existing order/invoice flow remains available

## Future Enhancements

1. **Async Order Creation**: Implement proper async processing for order/invoice creation after payment success
2. **Payment Webhooks**: Handle payment status updates to trigger order/subscription creation
3. **Retry Mechanism**: Add retry logic for failed async operations
4. **Metrics**: Add monitoring and metrics for quick purchase success rates

## Configuration

No additional configuration is required. The feature uses existing:
- Payment gateway configurations
- Coupon service settings
- Subscription plan data
- Security settings