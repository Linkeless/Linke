# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Build and Run:**
- `make build` - Build the application to `bin/server`
- `make run` - Run the application directly
- `make dev` - Run in development mode (generates Swagger docs first)

**Testing and Quality:**
- `make test` - Run all tests with verbose output (`go test -v ./...`)

**Documentation:**
- `make swagger` - Generate Swagger API documentation in `docs/`
- Swagger UI available at `/swagger/*any` when server is running

**Dependencies:**
- `make install` - Install dependencies and required tools (swag)
- `make clean` - Clean build artifacts and docs

## Architecture Overview

**Core Application Structure:**
- **Entry Point**: `cmd/server/main.go` - Server initialization with graceful shutdown
- **Configuration**: Environment-based config with `.env` support via `config/config.go`
- **Database Layer**: GORM for MySQL + Redis for caching/queues in `internal/repository/`
- **Business Logic**: Service layer in `internal/service/`
- **API Layer**: Gin HTTP handlers in `internal/handler/`
- **Background Processing**: Redis-based task queue system in `internal/queue/`

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

**Database Architecture:**
- GORM with auto-migration support (`internal/migration/migration.go`)
- Connection pooling configuration (max idle: 10, max open: 100)
- Soft delete middleware for query scoping (`internal/middleware/soft_delete.go`)
- Models: User, InviteCode with proper relationships

## Environment Configuration

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
- User management: `/admin/users` with full CRUD operations
- Invite code management: `/admin/invite-codes` with statistics
- Task management: `POST /tasks`, `GET /tasks/status`

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

**Logging and Monitoring:**
- Use structured logging with appropriate log levels
- Monitor invite code usage patterns
- Track authentication failures and security events
- Set up alerts for admin operations