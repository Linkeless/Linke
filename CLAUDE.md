# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Application Overview:**
Linke is a comprehensive service management platform with subscription-based billing, user management, and server administration. The system supports shadowsocks server management, multi-gateway payments, referral programs, and customer support through an integrated ticket system.

## Development Commands

**Build and Run:**
- `make build` - Build the application to `bin/server`
- `make run` - Run the application directly
- `make dev` - Run in development mode (generates Swagger docs first)
- `make safe-run` - Run with security pre-flight checks (RECOMMENDED)
- `make safe-dev` - Run in development mode with security checks (RECOMMENDED)

**Security:**
- `make security-check` - Run security pre-flight validation
- `go run tools/generate-jwt-key/main.go` - Generate secure JWT key

**Testing and Quality:**
- `make test` - Run all tests with verbose output (`go test -v ./...`)

**Documentation:**
- `make swagger` - Generate Swagger API documentation in `docs/`
- Swagger UI available at `/swagger/*any` when server is running

**Dependencies:**
- `make install` - Install dependencies and required tools (swag)
- `make clean` - Clean build artifacts and docs

**Database Migration (Integrated):**
- `make migrate-help` - Show detailed migration help and usage
- `make migrate-up` - Run all pending migrations
- `make migrate-down` - Rollback one migration
- `make migrate-status` - Show current migration version
- `make migrate-list` - List all applied migrations
- `make migrate-reset` - Drop all tables and re-run migrations (DANGEROUS!)
- `make migrate-create NAME=name` - Create new migration files
- `make migrate-force VERSION=N` - Force set migration version
- `make migrate-goto VERSION=N` - Migrate to specific version
- `make migrate-steps STEPS=N` - Run N migration steps (positive for up, negative for down)
- `make migrate-fix-dirty VERSION=N` - Fix dirty migration state (use version from error message)
- `go run cmd/server/main.go -migrate` - Run migrations only and exit (does not start server)
- All migration commands are integrated into the main server application

**Migration Troubleshooting:**
- If you see "Dirty database version X" error: `make migrate-fix-dirty VERSION=X`
- Or use: `go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=X`
- See `docs/MIGRATION_TROUBLESHOOTING.md` for detailed troubleshooting guide

## Architecture Overview

**Core Application Structure:**
- **Entry Point**: `cmd/server/main.go` - Server initialization with graceful shutdown
- **Configuration**: Environment-based config with `.env` support via `config/config.go`
- **Database Layer**: GORM for MySQL + Redis for caching/queues in `internal/repository/`
- **Business Logic**: Service layer in `internal/service/`
- **API Layer**: Gin HTTP handlers in `internal/handler/`
- **Background Processing**: Redis-based task queue system in `internal/queue/`
- **Module Management**: `internal/modules/manager_simple.go` - Centralized dependency injection and service wiring
- **Route Management**: `internal/routes/manager_simple.go` - Centralized route registration with middleware

**Key Components:**

**Authentication System** (`internal/handler/auth.go`, `internal/service/oauth.go`):
- Multi-provider OAuth2 (Google, GitHub, Telegram)
- Telegram Widget integration for web authentication
- User creation/update with provider-specific ID mapping
- Local account registration with email/password
- JWT token-based authentication with refresh support

**Task Queue System** (`internal/queue/`):
- Redis-backed asynchronous task processing
- Configurable retry logic with dead letter queues
- Built-in handlers for email, notifications, and data processing
- Graceful shutdown and context cancellation support

**User Management** (`internal/model/user.go`, `internal/handler/user.go`):
- Soft delete functionality with restore capabilities
- Multi-provider account linking (Google, GitHub, Telegram IDs)
- CRUD operations with proper error handling
- User status management (active, inactive, banned)
- Admin and regular user role support

**Structured Logging** (`internal/logger/logger.go`):
- Zap-based structured logging with multiple output formats
- Environment-configurable log levels and outputs
- Standardized field helpers for consistent logging

**Invite Code System** (`internal/model/invite_code.go`, `internal/service/invite_code.go`):
- Secure random code generation (32-character hex)
- Flexible usage limits (single-use or multi-use codes)
- Expiration time support with automatic validation
- Status management (active, used, expired, disabled)
- Integration with user registration process
- Admin monitoring and statistics

**Server Group Management** (`internal/model/server_group.go`, `internal/service/server_group.go`):
- Server group organization for network services
- Unique naming constraints with proper validation
- Full CRUD operations with admin controls
- Integration with shadowsocks server management

**Shadowsocks Server Management** (`internal/model/shadowsocks_server.go`, `internal/service/shadowsocks_server.go`):
- Complete shadowsocks server configuration management
- Server group association for organization
- User access control and subscription integration
- Rate limiting and traffic management features

**Payment System** (`internal/service/payment*.go`, `internal/handler/payment.go`):
- Multi-gateway payment processing (EPay, EPUSDT)
- Payment config management for different gateways
- Order tracking and notification handling
- Integration with subscription system for automated billing

**Subscription Management** (`internal/service/subscription*.go`, `internal/service/user_subscription.go`):
- Flexible subscription plan creation and management
- User subscription lifecycle (purchase, renewal, cancellation)
- Integration with payment system for automated billing
- Subscription order processing and tracking

**Referral System** (`internal/service/referral*.go`, `internal/handler/**/referral.go`):
- Referral campaign management with customizable rewards
- Referral tracking and analytics
- Integration with user registration and invite codes
- Campaign statistics and performance monitoring

**Support Ticket System** (`internal/service/ticket*.go`, `internal/handler/**/ticket.go`):
- Multi-tier ticket management (user and admin interfaces)
- Ticket messaging and conversation threading
- Ticket assignment and resolution workflows
- Decoupled ticket architecture for scalability

**Coupon System** (`internal/service/coupon.go`, `internal/handler/**/coupon.go`, `internal/model/coupon.go`):
- Flexible discount system supporting percentage and fixed-amount coupons
- Usage tracking and limitation controls (per-user and global usage limits)
- Time-based validity periods with automatic expiration
- Plan-specific applicability and minimum order requirements
- Public/private coupon visibility management
- Complete audit trail with usage history tracking
- Integration with subscription order creation and payment processing

**Database Migration System** (`internal/migration/migrate.go`, integrated in `cmd/server/main.go`):
- golang-migrate integration for versioned database migrations with automatic `schema_migrations` table management
- Fully integrated migration functionality within main server application (no separate CLI tool required)
- Support for up, down, reset, status, list, force, goto, steps, and fix-dirty operations
- golang-migrate automatically creates and manages `schema_migrations` table following best practices
- Uses same database configuration as main application
- Migration files in `migrations/` directory with SQL schema definitions
- Rich command-line interface with `-migrate-command`, `-migrate-version`, `-migrate-steps` flags
- Migration tracking with dirty state detection and comprehensive error recovery

**Database Architecture:**
- GORM for MySQL with connection pooling (max idle: 10, max open: 100)
- Redis for caching and task queues
- Soft delete middleware for query scoping (`internal/middleware/soft_delete.go`)
- Models: User, InviteCode, SubscriptionPlan, UserSubscription, PaymentRecord, Coupon, CouponUsage, etc.
- **Schema Optimizations**: Foreign key constraints, composite indexes, CHECK constraints, and performance optimizations
- **Audit & Security**: Login tracking, failed attempt monitoring, email verification, and comprehensive audit trails
- **Reporting Views**: `v_active_subscriptions` and `v_revenue_summary` for business analytics

## Environment Configuration

**🚨 CRITICAL SECURITY REQUIREMENTS:**

**JWT Configuration (REQUIRED):**
- `JWT_SECRET` - **CRITICAL**: Must be set with a secure random key (32+ chars)
  - Generate with: `openssl rand -hex 32`
  - Or use: `go run tools/generate-jwt-key/main.go`
  - NEVER use default values in production
- `JWT_EXPIRE_HOURS` - Token expiration time (default: 24)

**Security Checks:**
- Run `make security-check` before deployment
- Use `make safe-run` or `make safe-dev` for development
- The application will refuse to start with insecure JWT configuration

Required environment variables for full functionality:

**Database:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

**Redis:**
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`

**OAuth2 Providers:**
- Google: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- GitHub: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`
- Telegram: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_REDIRECT_URL`

**Logging:**
- `LOG_LEVEL` (debug, info, warn, error), `LOG_FORMAT` (text/json), `LOG_OUTPUT`

**Server:**
- `SERVER_PORT` (default: 8080)

**Server API Token:**
- `SERVER_API_TOKEN` - Authentication token for UniProxy server nodes (minimum 20 characters required)

**Database Migration:**
- `RUN_MIGRATION` (set to "true" to enable auto-migration)

## API Structure

**Authentication Routes** (`/api/v1/auth`):
- `GET /auth/providers` - List available OAuth providers
- `GET /auth/:provider` - Initiate OAuth login
- `GET /auth/:provider/callback` - Handle OAuth callbacks
- `GET /auth/telegram/widget` - Get Telegram login widget HTML
- `POST /auth/register` - Register new user with email/password (supports invite codes)
- `POST /auth/login` - Login with email/password
- `POST /auth/logout` - Logout current user
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/change-password` - Change user password
- `GET /auth/profile` - Get current user profile

**User Routes** (`/api/v1/user`):
- `GET /user/profile` - Get current user's profile
- `PUT /user/profile` - Update current user's profile
- `PUT /user/password` - Change current user's password

**Invite Code Routes** (`/api/v1/invite-codes`):
- `GET /invite-codes/validate/:code` - Validate invite code (public)
- `POST /invite-codes` - Create new invite code (authenticated)
- `GET /invite-codes/my` - Get my invite codes (authenticated)
- `GET /invite-codes/:id` - Get invite code details (authenticated)
- `PUT /invite-codes/:id/status` - Update invite code status (authenticated)
- `DELETE /invite-codes/:id` - Delete invite code (authenticated)

**Admin Routes** (`/api/v1/admin`):
- User management: `/admin/users` with full CRUD operations, search, stats, and batch operations
- Invite code management: `/admin/invite-codes` with statistics
- Server group management: `/admin/server-groups` with full CRUD operations
- Shadowsocks server management: `/admin/shadowsocks-servers` with full CRUD operations
- Subscription plan management: `/admin/subscriptions/plans` with full CRUD operations
- User subscription management: `/admin/subscriptions/users` with management and renewal
- Payment config management: `/admin/payments/configs` with CRUD operations
- Referral management: `/admin/referrals` and `/admin/referral-campaigns` with full CRUD operations
- Ticket management: Admin ticket routes with assignment and resolution
- Coupon management: `/admin/coupons` with full CRUD operations, usage tracking, and statistics

**User-Specific Routes** (`/api/v1/user`):
- Profile management: `/user/profile` and `/user/password`
- Subscription management: `/user/subscriptions` with purchase, cancel, and history
- Payment orders: `/user/payments/orders` with creation and tracking
- Shadowsocks servers: `/user/shadowsocks-servers` for available servers
- Referrals: `/user/referrals` and `/user/referrals/stats`
- Invite codes: `/user/invite-codes` with full management
- Tasks: `/user/tasks` for background task management
- Tickets: User ticket routes for support system
- Coupon operations: `/user/coupons/validate` for coupon validation and `/user/coupons/usages` for usage history

**Public Routes**:
- Subscription plans: `/api/v1/subscription-plans` with public plan listing
- Payment methods: `/api/v1/payments/methods` and `/api/v1/payments/configs`
- Payment notifications: `/api/v1/payments/notify/:gateway`
- Server status: `/api/v1/servers/status`
- Referral campaigns: `/api/v1/referral-campaigns` and `/api/v1/referral/track/:code`
- Public coupons: `/api/v1/coupons` for browsing available public coupons

**Server API Routes** (`/api/v1/server`):
- UniProxy Server API: `/server/UniProxy/health`
- Subscription orders: `/subscription-orders` with authenticated access

**Health Check Routes**:
- `GET /health` - Health check endpoint
- `GET /api/v1/ping` - API ping endpoint

## Development Patterns

**Error Handling:**
- Structured logging with context fields throughout
- Graceful degradation for missing OAuth provider configurations
- Database transaction rollback patterns in handlers

**Database Patterns:**
- Use `internal/middleware/soft_delete.go` scopes for deleted record handling
- Provider-specific user lookups with fallback creation
- Auto-migration runs on server startup
- Transaction-based invite code usage to ensure data consistency

**Handler Organization:**
- **Admin Handlers** (`internal/handler/admin/`): Admin-only operations requiring elevated privileges
- **User Handlers** (`internal/handler/user/`): User-specific operations with authentication
- **Root Handlers** (`internal/handler/`): General API handlers (auth, payments, tasks, server APIs)
- Consistent response formatting using `internal/response/` package
- Structured error handling with proper HTTP status codes

**Module Architecture:**
- `SimpleManager` in `internal/modules/` handles dependency injection
- Services are initialized once and shared across handlers
- Circular dependency resolution through careful service initialization order
- Centralized module configuration and management

**Queue Patterns:**
- Register task handlers in `main.go` during startup
- Tasks include retry logic and dead letter queue routing
- Context-aware processing with graceful shutdown

**Configuration:**
- Environment variables with sensible defaults
- Centralized config loading in `config/config.go`
- Provider feature toggles based on credential availability

**Security Patterns:**
- JWT token-based authentication with refresh mechanism
- User status validation on all authenticated endpoints
- Only active users can access protected resources
- Invite code validation with expiration and usage limits
- Admin-only routes protected by role-based middleware
- OAuth2 data comparison to minimize unnecessary database updates

## Usage Examples

**Creating an Invite Code:**
```bash
curl -X POST http://localhost:8080/api/v1/invite-codes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "max_uses": 5,
    "expires_at": "2024-12-31T23:59:59Z",
    "description": "Friend invitation code"
  }'
```

**Registering with Invite Code:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepassword123",
    "invite_code": "a1b2c3d4e5f6789012345678901234567890abcd"
  }'
```

**Validating an Invite Code:**
```bash
curl -X GET http://localhost:8080/api/v1/invite-codes/validate/a1b2c3d4e5f6789012345678901234567890abcd
```

**Coupon Management Examples:**
```bash
# Get public coupons (no authentication required)
curl -X GET http://localhost:8080/api/v1/coupons

# Create a coupon (admin only)
curl -X POST http://localhost:8080/api/v1/admin/coupons \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "SAVE20",
    "name": "20% Off All Plans",
    "description": "Save 20% on any subscription plan",
    "type": "percentage",
    "value": 20,
    "max_uses": 100,
    "max_uses_per_user": 1,
    "min_order_amount": 10,
    "currency": "USD",
    "valid_until": "2024-12-31T23:59:59Z",
    "is_public": true
  }'

# Validate a coupon (user authentication required)
curl -X POST http://localhost:8080/api/v1/user/coupons/validate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "coupon_code": "SAVE20",
    "plan_id": 1,
    "order_amount": 29.99,
    "currency": "USD"
  }'

# Get coupon usage history (admin only)
curl -X GET http://localhost:8080/api/v1/admin/coupons/1/usages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get my coupon usage history (user authentication required)
curl -X GET http://localhost:8080/api/v1/user/coupons/usages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**UniProxy Server API Usage:**
```bash
# Get server configuration (for UniProxy nodes)
curl -X GET "http://localhost:8080/api/v1/server/UniProxy/config?node_id=1&node_type=shadowsocks&token=uniproxy-server-token-2024-secure-key-example"

# Get users list (for UniProxy nodes)
curl -X GET "http://localhost:8080/api/v1/server/UniProxy/user?node_id=1&node_type=shadowsocks&token=uniproxy-server-token-2024-secure-key-example"

# Health check (for UniProxy nodes)
curl -X GET "http://localhost:8080/api/v1/server/UniProxy/health"

# Push data from UniProxy node
curl -X POST "http://localhost:8080/api/v1/server/UniProxy/push?node_id=1&node_type=shadowsocks&token=uniproxy-server-token-2024-secure-key-example" \
  -H "Content-Type: application/json" \
  -d '{"data": "node statistics or traffic data"}'
```

**Server Group Access Configuration:**
```json
# Allow access to specific server groups
{"server_group_ids": "[1,2,3]"}

# Allow access to all server groups (explicit)
{"server_group_ids": "[0]"}

# No access (secure default)
{"server_group_ids": ""}
```

**Subscription Server Group Assignment:**
- **SubscriptionPlan**: Contains `default_server_group_ids` (JSON) for automatic assignment
- **User Purchase**: Automatically inherits plan's default server groups
- **Admin Creation**: Can override server groups via API
- **Default Behavior**: Empty plan configuration grants access to all groups (`[0]`)
- **Security**: Empty subscription `server_group_ids` denies all access

## Database Schema Best Practices

**Schema Optimizations** (Migrations 000009, 000010 & 000011):
- **Foreign Key Constraints**: Full referential integrity with proper CASCADE/RESTRICT actions
- **Composite Indexes**: Optimized for common query patterns (auth, billing, reporting)
- **Data Validation**: CHECK constraints for status values, business rules (MySQL 8.0.16+)
- **Performance Monitoring**: Gateway response times, retry tracking, login patterns
- **Security Enhancements**: Failed login tracking, account lockout, email verification
- **Audit Trail**: Comprehensive logging for user actions and payment processing
- **Reporting Views**: Pre-built views for subscription analytics and revenue reporting
- **JSON Optimization**: Efficient storage for metadata and configuration data (MySQL 5.7+)
- **MySQL Compatibility**: Fixed syntax issues for MySQL/MariaDB compatibility

See `docs/DATABASE_OPTIMIZATIONS.md` for detailed implementation guide.

## Best Practices

**Authentication:**
- Always use HTTPS in production
- Implement proper token refresh logic in frontend
- Store JWT tokens securely (httpOnly cookies recommended)
- Validate user status on critical operations

**Invite Code Management:**
- Set appropriate expiration times for invite codes
- Monitor invite code usage through admin statistics
- Implement rate limiting on invite code creation
- Use single-use codes for sensitive invitations

**Database Operations:**
- Enable database migration in production with `RUN_MIGRATION=true`
- Monitor database connection pool usage
- Use soft delete for audit trails
- Implement proper indexing for performance

**Server API Token Security:**
- Use strong, randomly generated tokens (minimum 20 characters)
- Rotate tokens regularly in production environments
- Store tokens securely as environment variables
- Never log or expose tokens in responses
- Implement rate limiting for server API endpoints
- Monitor failed authentication attempts

**UniProxy Server API Security:**
- Subscription access validation with multiple security layers:
  - Time-based validation using database NOW() function for timezone consistency
  - Traffic limit enforcement (both database flag and real-time usage check)
  - Trial period expiration validation
  - Server group access control with explicit permission model
- Server group access follows "deny by default" principle:
  - Empty `server_group_ids` denies access (secure default)
  - Use `[0]` in JSON to explicitly grant access to all groups
  - Specific group IDs must be listed for limited access
- Support for both `active` and `trial` subscription statuses

**Logging and Monitoring:**
- Use structured logging with appropriate log levels
- Monitor invite code usage patterns
- Track authentication failures and security events
- Set up alerts for admin operations
- Monitor server API token usage and failed attempts
- Monitor coupon usage patterns and fraud detection

**Coupon Management:**
- Set appropriate expiration times and usage limits for coupons
- Monitor coupon usage statistics and conversion rates
- Implement fraud detection for suspicious coupon usage patterns
- Use percentage coupons carefully to avoid excessive discounts
- Set minimum order amounts to prevent abuse
- Regularly audit coupon effectiveness and ROI
- Archive expired coupons but maintain usage history for analytics