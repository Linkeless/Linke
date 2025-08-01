# User Domain Refactoring - DDD Best Practices

## Overview

The User aggregate has been refactored according to strict DDD principles to eliminate the identified violations and create a more maintainable, domain-focused architecture.

## Refactored Architecture

### 1. Aggregate Structure
```
internal/user/domain/
├── aggregate/
│   └── user.go                 # Refactored User aggregate root
├── entity/
│   ├── oauth_account.go        # OAuth account entity
│   └── user_profile.go         # User profile entity
├── valueobject/
│   ├── username.go             # Username value object
│   ├── display_name.go         # Display name value object
│   ├── avatar_url.go           # Avatar URL value object
│   ├── invite_code.go          # Invite code value object
│   ├── user_errors.go          # Domain-specific errors
│   └── (existing VOs...)       # Email, Password, UserStatus, etc.
├── service/
│   ├── authentication_service.go    # Authentication domain service
│   └── user_business_rules.go       # Business rules service
├── factory/
│   └── user_factory.go         # User factory for complex creation
└── event/
    └── user_events.go          # Domain events (existing)
```

## Key Improvements

### 1. **Eliminated Primitive Obsession**
**Before**: `username string`, `name string`, `avatar string`
**After**: 
- `Username` value object with validation rules
- `DisplayName` value object with length constraints  
- `AvatarURL` value object with URL validation

### 2. **Extracted Entities from Aggregate**
**Before**: OAuth providers stored as `map[string]valueobject.ProviderID`
**After**: 
- `OAuthAccount` entity with proper identity and behavior
- `UserProfile` entity managing profile-related concerns

### 3. **Separated Authentication Concerns**
**Before**: Authentication logic mixed within User aggregate
**After**: 
- `AuthenticationService` domain service for auth logic
- `UserBusinessRules` service for permission validation

### 4. **Improved Value Objects**
**Before**: Primitive types with minimal validation
**After**:
- `InviteCode` value object instead of separate `*uint` and `*string`
- Enhanced validation and business rules in all value objects

### 5. **Better Error Handling**
**Before**: Generic `errors.New()` calls
**After**: 
- `UserDomainError` typed errors with error codes
- Predefined domain errors for common scenarios
- Field-specific validation errors

### 6. **Factory Pattern Implementation**
**Before**: Constructor functions with many parameters
**After**: 
- `UserFactory` with request objects for complex creation
- Type-safe parameter objects eliminate primitive obsession
- Clear separation of creation concerns

## Domain Design Principles Applied

### 1. **Single Responsibility Principle**
- User aggregate: Core identity and coordination
- OAuthAccount entity: OAuth-specific concerns  
- UserProfile entity: Profile management
- AuthenticationService: Authentication rules

### 2. **Ubiquitous Language**
- `DisplayName` instead of generic `name`
- `AvatarURL` instead of generic `avatar`
- `InviteCode` instead of separate ID/code fields

### 3. **Invariant Protection**
- Value objects enforce data integrity at creation
- Business rules service validates operations
- Domain errors provide clear failure reasons

### 4. **Rich Domain Model**
- Behavior-focused entities with meaningful methods
- Domain services for cross-entity logic
- Factory for complex creation scenarios

## Benefits Achieved

### 1. **Type Safety**
- Eliminated primitive parameter confusion
- Compile-time validation of domain concepts
- Clear API contracts through request objects

### 2. **Maintainability**
- Separated concerns reduce cognitive load
- Domain logic is centralized and testable
- Clear boundaries between entities

### 3. **Extensibility**
- Easy to add new OAuth providers via entity
- Profile fields can be extended in dedicated entity
- Business rules are centralized and configurable

### 4. **Testing**
- Each component has focused responsibilities
- Domain services can be unit tested independently
- Value objects have isolated validation logic

## Migration Strategy

### Phase 1: Parallel Implementation
1. Create new domain structure alongside existing model
2. Add adapter layer to convert between old/new models
3. Gradually migrate application services to use new aggregate

### Phase 2: Repository Updates
1. Update repository interfaces to work with new entities  
2. Modify persistence layer to handle entity decomposition
3. Add migration scripts for data structure changes

### Phase 3: Complete Migration
1. Replace all usages of old User model
2. Remove old model and conversion adapters
3. Update integration tests and documentation

## Usage Examples

### Creating a User
```go
factory := factory.NewUserFactory()

// Local user
user, err := factory.CreateLocalUser("user@example.com", "password123")

// OAuth user  
user, err := factory.CreateOAuthUser(factory.CreateOAuthUserRequest{
    Email:      "user@example.com",
    Provider:   "google", 
    ProviderID: "google-id-123",
    // ... other fields
})
```

### Authentication
```go
authService := service.NewAuthenticationService()
err := authService.ValidateCredentials(
    user.Password(), 
    "candidate-password",
    user.Status(),
    user.Provider(),
)
```

### Business Rules
```go
rules := service.NewUserBusinessRules()
err := rules.CanDeleteUser(
    currentUserID,
    targetUserID, 
    currentUserRole,
    targetUserRole,
    isLastAdmin,
)
```

This refactoring transforms the User model from an anemic domain model with primitive obsession into a rich domain model that properly encapsulates business logic and enforces domain invariants.