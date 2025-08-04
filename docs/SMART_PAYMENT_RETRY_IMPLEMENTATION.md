# Smart Payment Retry Strategy Implementation

This document details the comprehensive implementation of the Smart Payment Retry Strategy feature for the Linke payment system, based on the requirements outlined in `SUBSCRIPTION_OPTIMIZATION.md`.

## Overview

The Smart Payment Retry Strategy intelligently retries failed payments using configurable retry strategies, exponential backoff algorithms, and failure type classification. This system significantly improves payment success rates while reducing manual intervention.

## Key Features

### 1. Intelligent Retry Logic
- **Exponential Backoff**: Default strategy with 1 hour, 6 hours, and 24 hours delays
- **Linear Backoff**: Alternative strategy for specific gateways
- **Custom Strategies**: Gateway-specific retry configurations
- **Failure Classification**: Automatic classification of temporary vs permanent failures

### 2. Configurable Retry Strategies
- **Per-Gateway Configuration**: Different strategies for different payment gateways
- **Per-Payment Method**: Fine-tuned strategies for specific payment methods
- **Dynamic Configuration**: Runtime updates without system restart
- **Default Fallbacks**: Comprehensive default strategies for all supported gateways

### 3. Comprehensive Monitoring
- **Real-time Health Metrics**: System-wide retry health monitoring
- **Gateway-specific Statistics**: Success rates and performance metrics per gateway
- **Failure Pattern Analysis**: Automatic detection of common failure patterns
- **Alert System**: Configurable alerts for system health issues

### 4. Admin Management Interface
- **Retry Queue Management**: View and manage all active retries
- **Bulk Operations**: Cancel or reset multiple retries at once
- **Detailed Analytics**: Comprehensive statistics and reporting
- **Manual Intervention**: Admin override capabilities

## Architecture

### Database Schema

#### Payment Retries Table
```sql
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    attempt_number INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP NOT NULL,
    last_attempt_at TIMESTAMP NOT NULL,
    retry_strategy VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    failure_type VARCHAR(30) NULL,
    -- ... additional fields
);
```

#### Payment Retry History Table
```sql
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    attempt_number INT NOT NULL,
    attempted_at TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL,
    -- ... additional fields
);
```

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Payment Module                         │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryService                                        │
│  ├─── InitiateRetry()                                       │
│  ├─── ProcessPendingRetries()                               │
│  ├─── ClassifyFailure()                                     │
│  └─── GetRetryHealthMetrics()                               │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryWorker                                         │
│  ├─── handleProcessPaymentRetry()                           │
│  ├─── handleProcessPendingRetries()                         │
│  └─── handleRetryHealthCheck()                              │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryRepository                                     │
│  ├─── Create(), Update(), Delete()                          │
│  ├─── GetRetriesDueForProcessing()                          │
│  └─── GetRetryStatsByGateway()                              │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Details

### 1. Retry Entities

#### PaymentRetry Entity
```go
type PaymentRetry struct {
    ID              uint      `json:"id"`
    PaymentRecordID uint      `json:"payment_record_id"`
    AttemptNumber   int       `json:"attempt_number"`
    MaxAttempts     int       `json:"max_attempts"`
    NextRetryAt     time.Time `json:"next_retry_at"`
    RetryStrategy   string    `json:"retry_strategy"`
    Status          string    `json:"status"`
    FailureType     string    `json:"failure_type"`
    // ... additional fields
}
```

#### PaymentRetryHistory Entity
```go
type PaymentRetryHistory struct {
    ID              uint      `json:"id"`
    PaymentRetryID  uint      `json:"payment_retry_id"`
    AttemptNumber   int       `json:"attempt_number"`
    AttemptedAt     time.Time `json:"attempted_at"`
    Status          string    `json:"status"`
    Duration        int       `json:"duration"`
    ResponseCode    string    `json:"response_code"`
    // ... additional fields
}
```

### 2. Retry Strategies

#### Default Retry Configuration
```go
var DefaultRetryStrategies = map[string]RetryStrategyConfig{
    PaymentGatewayEpay: {
        Gateway:           PaymentGatewayEpay,
        MaxAttempts:       3,
        InitialDelay:      3600,  // 1 hour
        MaxDelay:          86400, // 24 hours
        BackoffFactor:     2.0,
        Strategy:          RetryStrategyExponential,
        FailureTypes:      []string{FailureTypeTemporary, FailureTypeNetwork, FailureTypeGateway},
        TimeoutSeconds:    30,
        EnableAfterHours:  true,
        MaxConcurrent:     5,
    },
    // ... other gateways
}
```

#### Exponential Backoff Algorithm
```go
func (pr *PaymentRetry) calculateExponentialDelay() int {
    if pr.AttemptNumber == 0 {
        return pr.InitialDelay
    }
    
    // Calculate: initial_delay * (backoff_factor ^ attempt_number)
    delay := float64(pr.InitialDelay)
    for i := 0; i < pr.AttemptNumber; i++ {
        delay *= pr.BackoffFactor
    }
    
    // Ensure we don't exceed max delay
    if int(delay) > pr.MaxDelay {
        return pr.MaxDelay
    }
    
    return int(delay)
}
```

### 3. Failure Classification

#### Automatic Failure Type Detection
```go
func (s *paymentRetryService) ClassifyFailure(ctx context.Context, gateway, paymentMethod, errorCode, errorMessage string) string {
    // Permanent failure patterns
    permanentPatterns := []string{
        "invalid card", "card expired", "insufficient funds",
        "card declined", "invalid account", "account closed",
    }
    
    // Network failure patterns
    networkPatterns := []string{
        "timeout", "network", "connection", "dns", "ssl",
    }
    
    // Gateway failure patterns
    gatewayPatterns := []string{
        "gateway", "service unavailable", "server error", "maintenance",
    }
    
    // Classification logic...
}
```

### 4. Task Queue Integration

#### Retry Task Worker
```go
type PaymentRetryWorker struct {
    retryService interfaces.PaymentRetryService
    processor    *queue.TaskProcessor
}

func (w *PaymentRetryWorker) handleProcessPaymentRetry(ctx context.Context, task *asynq.Task) error {
    var payload struct {
        RetryID uint `json:"retry_id"`
    }
    
    // Unmarshal payload and process retry
    // ...
}
```

#### Scheduled Processing
- **Pending Retries**: Processed every 5 minutes
- **Health Checks**: Performed every 15 minutes  
- **Cleanup Tasks**: Daily maintenance at 2 AM
- **Retry Notifications**: Real-time notifications for retry events

## API Endpoints

### Admin Retry Management

#### Get Payment Retries
```http
GET /admin/payments/retries
Query Parameters:
- user_id: Filter by user ID
- gateway: Filter by gateway (epay, epusdt)
- status: Filter by retry status
- failure_type: Filter by failure type
- from_date, to_date: Date range filter
- limit, offset: Pagination
```

#### Get Retry Details
```http
GET /admin/payments/retries/{id}
Response: Detailed retry information with complete history
```

#### Cancel Retry
```http
POST /admin/payments/retries/{id}/cancel
Body: { "reason": "Manual cancellation by admin" }
```

#### Reset Retry
```http
POST /admin/payments/retries/{id}/reset
```

#### Bulk Operations
```http
POST /admin/payments/retries/bulk-cancel
POST /admin/payments/retries/bulk-reset
Body: { "retry_ids": [1,2,3], "reason": "Bulk operation" }
```

#### Statistics and Health
```http
GET /admin/payments/retries/statistics?gateway=epay&days=30
GET /admin/payments/retries/health
```

## Configuration

### Environment Variables
```bash
# Retry System Configuration
RETRY_ENABLED=true
RETRY_MAX_CONCURRENT=100
RETRY_PROCESSING_INTERVAL=5m
RETRY_HEALTH_CHECK_INTERVAL=15m

# Notification Settings
RETRY_NOTIFICATIONS_ENABLED=true
RETRY_NOTIFY_ON_FAILURE=true
RETRY_NOTIFY_ON_SUCCESS=false
```

### Gateway-Specific Configuration
```go
// Runtime configuration updates
retryService.UpdateRetryStrategy(ctx, "epay", "alipay", &RetryStrategyConfig{
    MaxAttempts:      5,
    InitialDelay:     1800, // 30 minutes
    MaxDelay:         43200, // 12 hours
    BackoffFactor:    1.5,
    Strategy:         RetryStrategyExponential,
})
```

## Monitoring and Alerting

### Health Metrics
- **Total Active Retries**: Current number of pending/in-progress retries
- **Overdue Retries**: Retries that should have been processed but weren't
- **Success Rates**: 24-hour and 7-day success rates
- **Gateway Health**: Per-gateway performance metrics

### Automatic Alerts
- High number of pending retries (>50)
- High number of overdue retries (>10)
- Low success rate (<80%)
- Gateway-specific health issues

### Logging
All retry operations are comprehensively logged with structured logging:
```go
logger.Info("Payment retry initiated successfully",
    logger.Uint("retry_id", retry.ID),
    logger.Uint("payment_record_id", paymentRecord.ID),
    logger.String("next_retry_at", retry.NextRetryAt.String()),
)
```

## Testing

### Unit Tests
Comprehensive test coverage for:
- Retry service operations
- Exponential backoff calculations
- Failure classification logic
- Repository operations
- Worker task processing

### Integration Tests
- End-to-end retry workflows
- Database operations
- Queue integration
- API endpoint testing

## Migration

### Database Migration
```bash
# Apply retry tables migration
make migrate-up

# Verify migration status
make migrate-status
```

### Deployment Steps
1. **Deploy Database Changes**: Apply migration 000017
2. **Deploy Application**: Update application with retry services
3. **Configure Retry Strategies**: Set up gateway-specific configurations
4. **Enable Processing**: Start retry workers and processing
5. **Monitor Health**: Verify system health and metrics

## Performance Considerations

### Scalability
- **Horizontal Scaling**: Multiple worker instances can process retries
- **Queue Partitioning**: Separate queues for different priorities
- **Database Optimization**: Proper indexing for retry queries
- **Cache Integration**: Cached strategies and configurations

### Resource Management
- **Memory Usage**: Efficient data structures and object pooling
- **Database Connections**: Connection pooling and query optimization
- **Queue Processing**: Configurable concurrency limits
- **Cleanup**: Automatic cleanup of old retry records

## Security

### Access Control
- **Admin-Only Access**: Retry management restricted to admin users
- **Rate Limiting**: API endpoint rate limiting
- **Input Validation**: Comprehensive request validation
- **Audit Logging**: All admin actions are logged

### Data Protection
- **Sensitive Data**: Payment data is properly masked in logs
- **Encryption**: Sensitive configuration data encryption
- **Access Logs**: Comprehensive access logging for security monitoring

## Troubleshooting

### Common Issues

#### High Retry Queue Depth
**Symptoms**: Large number of pending retries
**Causes**: Gateway issues, network problems, configuration errors
**Solutions**: 
- Check gateway health
- Verify network connectivity
- Review retry configurations
- Scale worker instances

#### Low Success Rates
**Symptoms**: Retries consistently failing
**Causes**: Permanent failures being retried, incorrect classification
**Solutions**:
- Review failure classification rules
- Update retry strategies
- Check gateway-specific issues

#### Performance Issues
**Symptoms**: Slow retry processing
**Causes**: Database bottlenecks, queue overload
**Solutions**:
- Optimize database queries
- Scale worker instances
- Review queue configuration

## Future Enhancements

### Planned Features
1. **Machine Learning Classification**: AI-powered failure classification
2. **Adaptive Retry Strategies**: Dynamic strategy adjustment based on success rates
3. **Advanced Analytics**: Predictive analytics for payment success
4. **Multi-region Support**: Distributed retry processing
5. **Enhanced Notifications**: Rich notifications with payment context

### Performance Optimizations
1. **Caching Layer**: Redis-based caching for configurations
2. **Database Sharding**: Horizontal database scaling
3. **Async Processing**: Non-blocking retry processing
4. **Batch Operations**: Bulk retry processing capabilities

## Conclusion

The Smart Payment Retry Strategy implementation provides a robust, scalable, and intelligent solution for handling payment failures. With comprehensive monitoring, flexible configuration, and powerful admin tools, this system significantly improves payment success rates while reducing operational overhead.

The implementation follows clean architecture principles, provides extensive test coverage, and includes comprehensive documentation for maintenance and future enhancements.