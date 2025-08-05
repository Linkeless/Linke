package framework

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Logger interface defines the logging contract for the framework
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

// Database interface defines the database operations contract
type Database interface {
	GetDB() *gorm.DB
	GetRedis() *redis.Client
	TestConnections(ctx context.Context) error
	TestMySQLConnection(ctx context.Context) error
	TestRedisConnection(ctx context.Context) error
	WithTransaction(fn func(*gorm.DB) error) error
	HealthCheck(ctx context.Context) map[string]bool
	Close() error
}

// Config interface defines the configuration contract
type Config interface {
	GetServerConfig() ServerConfig
	GetDatabaseConfig() DatabaseConfig
	GetRedisConfig() RedisConfig
	GetJWTConfig() JWTConfig
	GetLogConfig() LogConfig
	GetOAuth2Config() OAuth2Config
	GetAPIConfig() APIConfig
}

// ServerConfig interface for server configuration
type ServerConfig interface {
	GetPort() string
}

// DatabaseConfig interface for database configuration
type DatabaseConfig interface {
	GetHost() string
	GetPort() string
	GetUser() string
	GetPassword() string
	GetName() string
}

// RedisConfig interface for Redis configuration
type RedisConfig interface {
	GetHost() string
	GetPort() string
	GetPassword() string
	GetDB() int
}

// JWTConfig interface for JWT configuration
type JWTConfig interface {
	GetSecret() string
	GetExpireHours() int
}

// LogConfig interface for logging configuration
type LogConfig interface {
	GetLevel() string
	GetFormat() string
	GetOutput() string
}

// OAuth2Config interface for OAuth2 configuration
type OAuth2Config interface {
	GetGoogleClientID() string
	GetGoogleClientSecret() string
	GetGoogleRedirectURL() string
	GetGitHubClientID() string
	GetGitHubClientSecret() string
	GetGitHubRedirectURL() string
	GetTelegramBotToken() string
	GetTelegramRedirectURL() string
}

// APIConfig interface for API configuration
type APIConfig interface {
	GetServerToken() string
}

// Repository interface defines the base repository contract
type Repository interface {
	GetDB() *gorm.DB
	BeginTransaction() *gorm.DB
	CommitTransaction(tx *gorm.DB) error
	RollbackTransaction(tx *gorm.DB) error
}

// BaseRepository provides common repository operations (legacy interface - deprecated)
type BaseRepository interface {
	Repository
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id interface{}, entity interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}, entity interface{}) error
	List(ctx context.Context, entities interface{}, offset, limit int) error
	Count(ctx context.Context, entity interface{}) (int64, error)
}

// GenericRepository provides type-safe repository operations with full CRUD, pagination, soft delete, and batch operations
type GenericRepository[T any, ID comparable] interface {
	Repository
	
	// Basic CRUD operations
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id ID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
	
	// Soft delete operations
	SoftDelete(ctx context.Context, id ID) error
	Restore(ctx context.Context, id ID) error  
	HardDelete(ctx context.Context, id ID) error
	
	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*T, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*T, int64, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*T, int64, error)
	
	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*T, int64, error)
	
	// Status management
	UpdateStatus(ctx context.Context, id ID, status string) error
	
	// Statistics operations
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountDeleted(ctx context.Context) (int64, error)
		
	// Batch operations
	BatchDelete(ctx context.Context, ids []ID) (int, []ID, error)     // returns (successCount, failedIDs, error)
	BatchRestore(ctx context.Context, ids []ID) (int, []ID, error)    // returns (successCount, failedIDs, error)
	BatchUpdateStatus(ctx context.Context, ids []ID, status string) (int, []ID, error) // returns (successCount, failedIDs, error)
	
	// Existence checks
	ExistsByID(ctx context.Context, id ID) (bool, error)
	
	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*T, int64, error)
}

// UserScopedRepository extends GenericRepository with user-specific operations
type UserScopedRepository[T any, ID comparable] interface {
	GenericRepository[T, ID]
	
	// User-specific operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*T, int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	GetUserTotalCount(ctx context.Context, userID uint) (int64, error)
}

// TimeBasedRepository extends GenericRepository with time-based operations
type TimeBasedRepository[T any, ID comparable] interface {
	GenericRepository[T, ID]
	
	// Time-based queries
	ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*T, int64, error)
	ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error)
	ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error)
}

// UserScopedTimeBasedRepository extends GenericRepository with both user-specific and time-based operations
type UserScopedTimeBasedRepository[T any, ID comparable] interface {
	GenericRepository[T, ID]
	
	// User-specific operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*T, int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	GetUserTotalCount(ctx context.Context, userID uint) (int64, error)
	
	// Time-based queries
	ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*T, int64, error)
	ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error)
	ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error)
}

// Service interface defines the base service contract
type Service interface {
	GetName() string
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// DomainService interface for domain-specific services
type DomainService interface {
	Service
	GetRepository() Repository
}

// GenericService provides type-safe service operations with full CRUD, validation, and business logic
type GenericService[T any, ID comparable] interface {
	Service
	
	// Basic CRUD operations
	Create(ctx context.Context, req *CreateRequest[T]) (*T, error)
	GetByID(ctx context.Context, id ID) (*T, error)
	Update(ctx context.Context, id ID, req *UpdateRequest[T]) (*T, error)
	Delete(ctx context.Context, id ID) error
	
	// Soft delete operations  
	SoftDelete(ctx context.Context, id ID) error
	Restore(ctx context.Context, id ID) error
	HardDelete(ctx context.Context, id ID) error
	
	// List operations with pagination
	List(ctx context.Context, req *ListRequest) (*ListResponse[T], error)
	ListDeleted(ctx context.Context, req *ListRequest) (*ListResponse[T], error)
	ListByStatus(ctx context.Context, status string, req *ListRequest) (*ListResponse[T], error)
	
	// Search operations
	Search(ctx context.Context, query string, req *ListRequest) (*ListResponse[T], error)
	
	// Status management
	UpdateStatus(ctx context.Context, id ID, status string) (*T, error)
	
	// Statistics operations
	GetStatistics(ctx context.Context) (*StatisticsResponse, error)
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountDeleted(ctx context.Context) (int64, error)
		
	// Batch operations
	BatchDelete(ctx context.Context, ids []ID) (*BatchOperationResponse, error)
	BatchRestore(ctx context.Context, ids []ID) (*BatchOperationResponse, error)
	BatchUpdateStatus(ctx context.Context, ids []ID, status string) (*BatchOperationResponse, error)
	
	// Existence checks
	ExistsByID(ctx context.Context, id ID) (bool, error)
	
	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, req *ListRequest) (*ListResponse[T], error)
	
	// Validation hooks (can be overridden by implementations)
	ValidateCreate(ctx context.Context, req *CreateRequest[T]) error
	ValidateUpdate(ctx context.Context, id ID, req *UpdateRequest[T]) error
	ValidateDelete(ctx context.Context, id ID) error
}

// UserScopedService extends GenericService with user-specific operations
type UserScopedService[T any, ID comparable] interface {
	GenericService[T, ID]
	
	// User-specific operations
	ListByUser(ctx context.Context, userID uint, req *ListRequest) (*ListResponse[T], error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	GetUserTotalCount(ctx context.Context, userID uint) (int64, error)
	DeleteByUser(ctx context.Context, userID uint, ids []ID) (*BatchOperationResponse, error)
	
	// User validation
	ValidateUserAccess(ctx context.Context, userID uint, id ID) error
}

// TimeBasedService extends GenericService with time-based operations
type TimeBasedService[T any, ID comparable] interface {
	GenericService[T, ID]
	
	// Time-based queries
	ListByDateRange(ctx context.Context, field string, start, end time.Time, req *ListRequest) (*ListResponse[T], error)
	ListCreatedAfter(ctx context.Context, after time.Time, req *ListRequest) (*ListResponse[T], error)
	ListUpdatedAfter(ctx context.Context, after time.Time, req *ListRequest) (*ListResponse[T], error)
	
	// Time-based statistics
	GetStatisticsByDateRange(ctx context.Context, start, end time.Time) (*StatisticsResponse, error)
}

// UserScopedTimeBasedService extends GenericService with both user-specific and time-based operations
type UserScopedTimeBasedService[T any, ID comparable] interface {
	GenericService[T, ID]
	
	// User-specific operations
	ListByUser(ctx context.Context, userID uint, req *ListRequest) (*ListResponse[T], error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	GetUserTotalCount(ctx context.Context, userID uint) (int64, error)
	DeleteByUser(ctx context.Context, userID uint, ids []ID) (*BatchOperationResponse, error)
	
	// Time-based queries
	ListByDateRange(ctx context.Context, field string, start, end time.Time, req *ListRequest) (*ListResponse[T], error)
	ListCreatedAfter(ctx context.Context, after time.Time, req *ListRequest) (*ListResponse[T], error)
	ListUpdatedAfter(ctx context.Context, after time.Time, req *ListRequest) (*ListResponse[T], error)
	
	// Combined user + time operations
	ListByUserAndDateRange(ctx context.Context, userID uint, field string, start, end time.Time, req *ListRequest) (*ListResponse[T], error)
	
	// User validation
	ValidateUserAccess(ctx context.Context, userID uint, id ID) error
	
	// Time-based statistics
	GetStatisticsByDateRange(ctx context.Context, start, end time.Time) (*StatisticsResponse, error)
	GetUserStatisticsByDateRange(ctx context.Context, userID uint, start, end time.Time) (*StatisticsResponse, error)
}

// BusinessService extends GenericService with business logic operations
type BusinessService[T any, ID comparable] interface {
	GenericService[T, ID]
	
	// Business validation
	ValidateBusinessRules(ctx context.Context, entity *T) error
	ValidateBusinessRulesForUpdate(ctx context.Context, id ID, req *UpdateRequest[T]) error
	
	// Event publishing
	PublishCreatedEvent(ctx context.Context, entity *T) error
	PublishUpdatedEvent(ctx context.Context, old *T, new *T) error  
	PublishDeletedEvent(ctx context.Context, entity *T) error
	
	// Audit operations
	GetAuditLog(ctx context.Context, id ID, req *ListRequest) (*ListResponse[AuditLogEntry], error)
	
	// Workflow operations
	ProcessWorkflow(ctx context.Context, id ID, action string, params map[string]interface{}) error
}

// Standard request/response DTOs for generic services

// CreateRequest represents a generic create request
type CreateRequest[T any] struct {
	Data     *T                     `json:"data" binding:"required"`
	Options  *CreateOptions         `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateRequest represents a generic update request  
type UpdateRequest[T any] struct {
	Data     *T                     `json:"data" binding:"required"`
	Options  *UpdateOptions         `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ListRequest represents a generic list request
type ListRequest struct {
	Limit   int                    `form:"limit,omitempty" binding:"omitempty,min=1,max=1000" example:"10"`
	Offset  int                    `form:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
	SortBy  string                 `form:"sort_by,omitempty" example:"created_at"`
	SortDir string                 `form:"sort_dir,omitempty" binding:"omitempty,oneof=asc desc" example:"desc"`
	Filters map[string]interface{} `form:"filters,omitempty"`
}

// ListResponse represents a generic list response
type ListResponse[T any] struct {
	Data       []*T  `json:"data"`
	Total      int64 `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	HasMore    bool  `json:"has_more"`
	TotalPages int   `json:"total_pages"`
}

// StatisticsResponse represents a generic statistics response
type StatisticsResponse struct {
	TotalCount    int64                  `json:"total_count"`
	ActiveCount   int64                  `json:"active_count"`
	InactiveCount int64                  `json:"inactive_count"`
	DeletedCount  int64                  `json:"deleted_count"`
	StatusCounts  map[string]int64       `json:"status_counts"`
	CustomStats   map[string]interface{} `json:"custom_stats,omitempty"`
	GeneratedAt   time.Time              `json:"generated_at"`
}

// BatchOperationResponse represents a generic batch operation response
type BatchOperationResponse struct {
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	FailedIDs    []uint `json:"failed_ids,omitempty"`
	Errors       map[string]string `json:"errors,omitempty"`
}

// CreateOptions represents options for create operations
type CreateOptions struct {
	SkipValidation    bool `json:"skip_validation,omitempty"`
	PublishEvents     bool `json:"publish_events,omitempty"`
	EnableAuditLog    bool `json:"enable_audit_log,omitempty"`
	ProcessWorkflows  bool `json:"process_workflows,omitempty"`
}

// UpdateOptions represents options for update operations
type UpdateOptions struct {
	SkipValidation    bool   `json:"skip_validation,omitempty"`
	PublishEvents     bool   `json:"publish_events,omitempty"`
	EnableAuditLog    bool   `json:"enable_audit_log,omitempty"`
	ProcessWorkflows  bool   `json:"process_workflows,omitempty"`
	UpdateMode        string `json:"update_mode,omitempty"` // "merge", "replace"
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID          uint                   `json:"id"`
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	Action      string                 `json:"action"` // create, update, delete, etc.
	UserID      *uint                  `json:"user_id,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// EventPublisher interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishAsync(ctx context.Context, event DomainEvent) error
}

// EventSubscriber interface for subscribing to domain events
type EventSubscriber interface {
	Subscribe(eventTypes []string, handler EventHandler) error
	Unsubscribe(eventTypes []string, handler EventHandler) error
}

// DomainEvent interface for domain events
type DomainEvent interface {
	EventType() string
	EventData() interface{}
	EventTime() time.Time
	EventID() string
	EventVersion() string
	EventSource() string
}

// EventHandler interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
	EventTypes() []string
}

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// Queue interface defines queue operations
type Queue interface {
	Enqueue(ctx context.Context, queueName string, task Task) error
	Dequeue(ctx context.Context, queueName string) (Task, error)
	GetQueueInfo(ctx context.Context, queueName string) (QueueInfo, error)
	Close() error
}

// Task interface for queue tasks
type Task interface {
	GetID() string
	GetType() string
	GetPayload() map[string]interface{}
	GetRetryCount() int
	GetMaxRetries() int
	GetCreatedAt() time.Time
}

// QueueInfo interface for queue information
type QueueInfo interface {
	GetName() string
	GetSize() int
	GetStatus() string
}

// Validator interface for data validation
type Validator interface {
	Validate(data interface{}) error
	ValidateStruct(data interface{}) error
}

// Serializer interface for data serialization
type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
	GetContentType() string
}

// HTTPClient interface for HTTP operations
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
	Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
	Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
}

// HTTPResponse interface for HTTP responses
type HTTPResponse interface {
	GetStatusCode() int
	GetHeaders() map[string]string
	GetBody() []byte
	GetBodyAsString() string
	GetBodyAsJSON(target interface{}) error
}

// Metrics interface for application metrics
type Metrics interface {
	Counter(name string, tags map[string]string) Counter
	Gauge(name string, tags map[string]string) Gauge
	Histogram(name string, tags map[string]string) Histogram
	Timer(name string, tags map[string]string) Timer
}

// Counter interface for counter metrics
type Counter interface {
	Increment(value float64)
	GetValue() float64
	Reset()
}

// Gauge interface for gauge metrics
type Gauge interface {
	Set(value float64)
	GetValue() float64
}

// Histogram interface for histogram metrics
type Histogram interface {
	Observe(value float64)
	GetCount() int64
	GetSum() float64
}

// Timer interface for timer metrics
type Timer interface {
	Record(duration time.Duration)
	Start() func()
	GetCount() int64
	GetSum() time.Duration
}

// HealthChecker interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) HealthStatus
	GetName() string
}

// HealthStatus interface for health status
type HealthStatus interface {
	IsHealthy() bool
	GetStatus() string
	GetMessage() string
	GetDetails() map[string]interface{}
	GetTimestamp() time.Time
}

// Security interface for security operations
type Security interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
	GenerateToken(claims map[string]interface{}) (string, error)
	ValidateToken(token string) (map[string]interface{}, error)
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// FileStorage interface for file operations
type FileStorage interface {
	Store(ctx context.Context, path string, data []byte) error
	Retrieve(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

// Migration interface for database migrations
type Migration interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Status() (string, error)
	Force(version int) error
	Goto(version uint) error
	Steps(n int) error
	Reset() error
}

// ApplicationContext interface for application-wide context
type ApplicationContext interface {
	GetLogger() Logger
	GetDatabase() Database
	GetConfig() Config
	GetCache() Cache
	GetQueue() Queue
	GetEventPublisher() EventPublisher
	GetEventSubscriber() EventSubscriber
	GetMetrics() Metrics
	GetSecurity() Security
	GetFileStorage() FileStorage
	GetHTTPClient() HTTPClient
	GetValidator() Validator
	GetSerializer() Serializer
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Module interface for application modules
type Module interface {
	GetName() string
	GetVersion() string
	GetDependencies() []string
	Initialize(ctx context.Context, app ApplicationContext) error
	Shutdown(ctx context.Context) error
	GetServices() []Service
	GetRoutes() []Route
}

// Route interface for HTTP routes
type Route interface {
	GetMethod() string
	GetPath() string
	GetHandler() interface{}
	GetMiddleware() []interface{}
}

// Middleware interface for HTTP middleware
type Middleware interface {
	Process(ctx context.Context, request interface{}, next func(context.Context, interface{}) (interface{}, error)) (interface{}, error)
}

// Error interfaces for structured error handling

// AppError interface for application errors
type AppError interface {
	error
	GetCode() string
	GetMessage() string
	GetDetails() map[string]interface{}
	GetCause() error
	GetTimestamp() time.Time
}

// ValidationError interface for validation errors
type ValidationError interface {
	AppError
	GetField() string
	GetValue() interface{}
	GetTag() string
}

// NotFoundError interface for not found errors
type NotFoundError interface {
	AppError
	GetResource() string
	GetID() interface{}
}

// UnauthorizedError interface for unauthorized errors
type UnauthorizedError interface {
	AppError
	GetReason() string
}

// ForbiddenError interface for forbidden errors
type ForbiddenError interface {
	AppError
	GetAction() string
	GetResource() string
}
