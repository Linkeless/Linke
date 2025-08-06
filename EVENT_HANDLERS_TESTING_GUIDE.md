# Event Handlers Testing Guide

This guide provides comprehensive testing instructions for the newly implemented cross-domain event handling system.

## Overview

The event handling system has been completely refactored to support full cross-domain business flows through event-driven architecture. Key features include:

- **Complete Business Flow Integration**: Payment → Order Processing → Subscription Activation → Invoice Generation → Usage Monitoring
- **Idempotency Protection**: All event handlers include duplicate processing prevention
- **Error Recovery**: Robust error handling with partial failure tolerance  
- **Real-time Usage Monitoring**: Automatic traffic limit checking and alerts
- **Async Processing**: Email notifications and heavy tasks run asynchronously

## Testing Strategy

### 1. Build and Dependency Verification

First, verify that all dependencies are correctly resolved:

```bash
# Build the application to check for compilation errors
make build

# Check for import issues
go mod tidy
go mod verify
```

### 2. Unit Testing Event Handlers

Test individual event handlers in isolation:

```bash
# Test event system components
go test ./internal/shared/events/... -v

# Test specific handlers if individual tests exist
go test ./internal/shared/events/ -run TestPaymentCompletedHandler -v
go test ./internal/shared/events/ -run TestSubscriptionCreatedHandler -v
```

### 3. Integration Testing

#### 3.1 Complete Payment Flow Test

Test the full payment → subscription activation flow:

```bash
# Start the application
make safe-dev

# In another terminal, test the flow
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "payment.completed",
    "payment_id": "test_payment_123",
    "user_id": 1,
    "amount": 29.99,
    "order_id": 1,
    "currency": "USD"
  }'
```

**Expected Flow:**
1. PaymentCompletedHandler processes payment
2. Order status updated to "paid"  
3. Subscription created/renewed
4. Invoice generated
5. Welcome/confirmation emails queued
6. User status activated

#### 3.2 User Registration Flow Test

Test user onboarding:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "securepassword123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Expected Flow:**
1. User created in database
2. UserRegisteredHandler triggers
3. Welcome email queued
4. User configuration cache initialized
5. Account ready for subscription

#### 3.3 Traffic Monitoring Test

Test usage monitoring and limit enforcement:

```bash
# Simulate traffic usage for a subscription
curl -X POST http://localhost:8080/api/v1/subscriptions/1/usage \
  -H "Content-Type: application/json" \
  -d '{
    "bytes_used": 85000000000,
    "subscription_id": 1
  }'
```

**Expected Flow:**
1. Usage updated in database
2. Warning events triggered at 80%, 90%
3. Limit exceeded event at 100%
4. Account suspended when limit exceeded
5. Alert emails sent to user

### 4. Event System Monitoring

#### 4.1 Event Bus Health Check

```bash
# Check event system status
curl http://localhost:8080/health

# Monitor application logs for event processing
tail -f logs/application.log | grep -i "event"
```

#### 4.2 Idempotency Testing

Test that duplicate events are handled correctly:

```bash
# Send the same payment completion event twice
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "duplicate_test_123",
    "event_type": "payment.completed", 
    "payment_id": "test_payment_456",
    "user_id": 1,
    "amount": 49.99,
    "order_id": 2
  }'

# Send again immediately - should be skipped
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "duplicate_test_123",
    "event_type": "payment.completed",
    "payment_id": "test_payment_456", 
    "user_id": 1,
    "amount": 49.99,
    "order_id": 2
  }'
```

**Expected Result:** Second request should log "already processed, skipping"

### 5. Database State Verification

After running tests, verify the database state:

```sql
-- Check event processing records
SELECT * FROM events ORDER BY created_at DESC LIMIT 10;

-- Check subscription states  
SELECT id, user_id, status, traffic_used, traffic_limit, traffic_suspended 
FROM user_subscriptions WHERE user_id = 1;

-- Check order processing
SELECT id, user_id, status, total_amount, created_at 
FROM subscription_orders ORDER BY created_at DESC LIMIT 5;

-- Check invoice generation
SELECT id, user_id, status, amount, invoice_number 
FROM invoices ORDER BY created_at DESC LIMIT 5;
```

### 6. Async Task Verification

Check that async tasks are being processed:

```bash
# Monitor task queue processing
redis-cli MONITOR | grep -i "task\|queue\|email\|notification"

# Check task queue status
redis-cli
> LLEN queue:email
> LLEN queue:notification  
> LLEN queue:data_processing
```

### 7. Cache System Testing

Verify caching behavior:

```bash
# Check cache keys
redis-cli KEYS "event_processed:*"
redis-cli KEYS "user_config:*" 
redis-cli KEYS "user_billing:*"

# Verify cache TTL
redis-cli TTL "event_processed:some_event_id"
```

## Load Testing

### Stress Test Event Processing

```bash
# Install Apache Bench if not available
brew install httpd  # macOS
# or
apt-get install apache2-utils  # Ubuntu

# Test payment webhook processing
ab -n 100 -c 10 -T application/json -p payment_data.json \
   http://localhost:8080/api/v1/payments/webhook

# payment_data.json content:
{
  "event_type": "payment.completed",
  "payment_id": "load_test_{{#}}",
  "user_id": 1,
  "amount": 29.99,
  "order_id": 1
}
```

## Error Scenarios Testing

### 1. Service Dependency Failures

Test behavior when services are unavailable:

```bash
# Simulate database connection issues
# Stop database temporarily and send events

# Simulate Redis cache unavailability  
# Stop Redis and test event processing

# Test with invalid data
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "payment.completed",
    "payment_id": "",
    "user_id": "invalid",
    "amount": -10,
    "order_id": "not_number"
  }'
```

### 2. Partial Processing Failures

Test graceful degradation:

- Email service down → Events still process, emails queued for retry
- Invoice service error → Payment still processed, invoice creation retried
- Cache unavailable → Events process without caching (slower but functional)

## Performance Monitoring

### Key Metrics to Track

1. **Event Processing Latency**
   - Average time from event publish to completion
   - Target: < 100ms for simple events, < 500ms for complex flows

2. **Throughput**  
   - Events processed per second
   - Target: > 100 events/second

3. **Error Rates**
   - Failed event processing percentage  
   - Target: < 1% failure rate

4. **Queue Depths**
   - Async task queue lengths
   - Target: Queues drain within 30 seconds under normal load

### Monitoring Commands

```bash
# Application metrics
curl http://localhost:8080/metrics

# Event processing stats
grep "event processed" logs/application.log | wc -l

# Error tracking
grep "ERROR" logs/application.log | tail -20
```

## Troubleshooting Common Issues

### Issue: Events not processing
**Diagnosis:**
```bash
# Check event bus initialization
grep "Event system initialized" logs/application.log

# Verify handler registration
grep "event handlers registered" logs/application.log
```

### Issue: High memory usage
**Diagnosis:**
```bash
# Check processed events cache
redis-cli DBSIZE
redis-cli KEYS "event_processed:*" | wc -l
```

### Issue: Duplicate processing
**Diagnosis:**  
```bash
# Check idempotency cache hits
grep "already processed, skipping" logs/application.log
```

## Rollback Plan

If issues arise in production:

1. **Immediate Rollback:** Revert to previous handlers.go version
2. **Partial Disable:** Comment out specific handler registrations  
3. **Emergency Mode:** Disable event system entirely (manual processing)

```go
// Emergency disable in bootstrap/app.go
// Comment out event handler registrations:
/*
if err := crossDomainHandlers.RegisterCrossDomainHandlers(eventBus); err != nil {
    logger.Error("Failed to register cross-domain event handlers", zap.Error(err))
}
*/
```

## Success Criteria

The implementation is successful if:

- ✅ All compilation errors resolved
- ✅ Payment → subscription flow works end-to-end  
- ✅ User registration triggers welcome flow
- ✅ Traffic monitoring generates alerts correctly
- ✅ Idempotency prevents duplicate processing
- ✅ Error handling maintains system stability  
- ✅ Async tasks process within acceptable timeframes
- ✅ Database state remains consistent across failures
- ✅ Performance meets defined targets

## Next Steps

After successful testing:

1. **Production Deployment:** Deploy with feature flags
2. **Monitoring Setup:** Configure alerts for error rates
3. **Documentation:** Update API documentation
4. **Team Training:** Brief team on new event flows
5. **Gradual Rollout:** Enable features incrementally