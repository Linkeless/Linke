package services

import (
	"linke/internal/shared/framework" 
	
	"go.uber.org/fx"
)

// ServiceModule provides the service layer module for dependency injection
var ServiceModule = fx.Module("services",
	// Provide service factory functions
	fx.Provide(
		// Generic service factories
		NewBaseServiceFactory,
		NewUserScopedServiceFactory,
		NewBusinessServiceFactory,
	),
)

// ServiceFactory interfaces for creating generic services
type BaseServiceFactory interface {
	CreateService(
		name string,
		repository framework.GenericRepository[any, any],
		logger framework.Logger,
		eventPub framework.EventPublisher,
		validator framework.Validator,
	) framework.GenericService[any, any]
}

type UserScopedServiceFactory interface {
	CreateUserScopedService(
		name string,
		repository framework.UserScopedRepository[any, any],
		logger framework.Logger,
		eventPub framework.EventPublisher,
		validator framework.Validator,
	) framework.UserScopedService[any, any]
}

type BusinessServiceFactory interface {
	CreateBusinessService(
		name string,
		repository framework.GenericRepository[any, any],
		logger framework.Logger,
		eventPub framework.EventPublisher,
		validator framework.Validator,
	) framework.BusinessService[any, any]
}

// Factory implementations
type baseServiceFactory struct{}

func NewBaseServiceFactory() BaseServiceFactory {
	return &baseServiceFactory{}
}

func (f *baseServiceFactory) CreateService(
	name string,
	repository framework.GenericRepository[any, any],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) framework.GenericService[any, any] {
	return NewBaseService(name, repository, logger, eventPub, validator)
}

type userScopedServiceFactory struct{}

func NewUserScopedServiceFactory() UserScopedServiceFactory {
	return &userScopedServiceFactory{}
}

func (f *userScopedServiceFactory) CreateUserScopedService(
	name string,
	repository framework.UserScopedRepository[any, any],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) framework.UserScopedService[any, any] {
	return NewUserScopedService(name, repository, logger, eventPub, validator)
}

type businessServiceFactory struct{}

func NewBusinessServiceFactory() BusinessServiceFactory {
	return &businessServiceFactory{}
}

func (f *businessServiceFactory) CreateBusinessService(
	name string,
	repository framework.GenericRepository[any, any],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) framework.BusinessService[any, any] {
	return NewBusinessService(name, repository, logger, eventPub, validator)
}