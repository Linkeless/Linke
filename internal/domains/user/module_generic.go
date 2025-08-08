package user

import (
	"go.uber.org/fx"

	"linke/internal/domains/user/adapters/repositories"
	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/adapters/handlers"
	"linke/internal/domains/user/usecases/implementations"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"
)

// ModuleGeneric demonstrates the new generic service-based user domain module
// This shows how to use the new generic service architecture
var ModuleGeneric = fx.Module("user-generic",
	// Provide generic repository implementation
	fx.Provide(
		// Create a generic repository for User entities
		func(db framework.Database, logger framework.Logger) framework.GenericRepository[entities.User, uint] {
			return repository.NewBaseRepository[entities.User, uint](db.GetDB(), logger)
		},
		
		// Create specific user repository if needed for domain-specific queries
		fx.Annotate(
			repositories.NewUserRepository,
			fx.As(new(interfaces.UserRepository)),
		),
	),

	// Provide generic service implementations
	fx.Provide(
		// Provide the new generic UserService implementation
		fx.Annotate(
			func(
				userRepo framework.GenericRepository[entities.User, uint],
				logger framework.Logger,
				eventPub framework.EventPublisher,
				validator framework.Validator,
			) interfaces.UserService {
				return implementations.NewUserServiceGeneric(userRepo, logger, eventPub, validator)
			},
			fx.As(new(interfaces.UserService)),
		),
	),

	// Provide handlers (unchanged)
	fx.Provide(
		handlers.NewUserProfileHandler,
		handlers.NewAdminUserHandler,
	),

	// Module initialization
	fx.Invoke(func(logger framework.Logger) {
		logger.Info("User domain module initialized with generic service architecture")
	}),
)

// GenericServiceProvider demonstrates how to provide generic services to other modules
type GenericServiceProvider struct {
	UserService framework.GenericService[entities.User, uint]
}

func NewGenericServiceProvider(userService interfaces.UserService) *GenericServiceProvider {
	// Since UserService now extends GenericService, we can cast it
	if genericService, ok := userService.(framework.GenericService[entities.User, uint]); ok {
		return &GenericServiceProvider{
			UserService: genericService,
		}
	}
	return nil
}

// GenericServiceModule provides generic user services to other modules
var GenericServiceModule = fx.Module("user-generic-service",
	fx.Provide(NewGenericServiceProvider),
)

// Example of how to use the generic service in another module
var ExampleUsageModule = fx.Module("user-example-usage",
	fx.Invoke(func(userService interfaces.UserService, logger framework.Logger) {
		// Example of using both generic and domain-specific methods
		logger.Info("Example of using generic user service:")
		
		// This would be used in actual application logic, not in module initialization
		// Just showing the interface usage
		
		// Generic operations available:
		// userService.Create(ctx, createRequest)
		// userService.GetByID(ctx, id)  
		// userService.Update(ctx, id, updateRequest)
		// userService.Delete(ctx, id)
		// userService.List(ctx, listRequest)
		// userService.Search(ctx, query, listRequest)
		// userService.GetStatistics(ctx)
		// userService.BatchDelete(ctx, ids)
		// userService.BatchRestore(ctx, ids)
		
		// Domain-specific operations still available:
		// userService.GetUserByEmail(ctx, email)
		// userService.GetActiveUserByID(ctx, id)
		// userService.ListUsersByProvider(ctx, provider, limit, offset)
		// userService.UpdateUserRole(ctx, id, role)
		// userService.GetUserStats(ctx)
		
		// Legacy methods still work for backward compatibility:
		// userService.CreateUser(ctx, user)
		// userService.UpdateUser(ctx, user)
		// userService.ListUsers(ctx, limit, offset)
	}),
)