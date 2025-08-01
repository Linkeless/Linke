package command_test

import (
	"context"
	"testing"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/service"
	"linke/internal/server/domain/valueobject"
	"linke/internal/server/service/command"
)

// MockServerGroupRepository is a mock implementation of ServerGroupRepository
type MockServerGroupRepository struct {
	groups map[uint]*model.ServerGroup
	nextID uint
}

func NewMockServerGroupRepository() *MockServerGroupRepository {
	return &MockServerGroupRepository{
		groups: make(map[uint]*model.ServerGroup),
		nextID: 1,
	}
}

func (m *MockServerGroupRepository) Save(ctx context.Context, group *model.ServerGroup) error {
	if group.ID().IsZero() {
		// Simulate auto-increment ID generation
		id, _ := valueobject.NewServerGroupID(m.nextID)
		group.MarkAsCreated(id)
		m.nextID++
	}
	m.groups[group.ID().Value()] = group
	return nil
}

func (m *MockServerGroupRepository) FindByID(ctx context.Context, id valueobject.ServerGroupID) (*model.ServerGroup, error) {
	if group, exists := m.groups[id.Value()]; exists {
		return group, nil
	}
	return nil, repository.ErrServerGroupNotFound
}

func (m *MockServerGroupRepository) FindByName(ctx context.Context, name valueobject.ServerGroupName) (*model.ServerGroup, error) {
	for _, group := range m.groups {
		if group.Name().Equals(name) {
			return group, nil
		}
	}
	return nil, repository.ErrServerGroupNotFound
}

func (m *MockServerGroupRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.ServerGroup, int64, error) {
	groups := make([]*model.ServerGroup, 0, len(m.groups))
	for _, group := range m.groups {
		groups = append(groups, group)
	}
	return groups, int64(len(groups)), nil
}

func (m *MockServerGroupRepository) ExistsByName(ctx context.Context, name valueobject.ServerGroupName) (bool, error) {
	for _, group := range m.groups {
		if group.Name().Equals(name) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockServerGroupRepository) ExistsByNameExcludingID(ctx context.Context, name valueobject.ServerGroupName, excludeID valueobject.ServerGroupID) (bool, error) {
	for _, group := range m.groups {
		if group.Name().Equals(name) && !group.ID().Equals(excludeID) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockServerGroupRepository) Delete(ctx context.Context, id valueobject.ServerGroupID) error {
	if _, exists := m.groups[id.Value()]; !exists {
		return repository.ErrServerGroupNotFound
	}
	delete(m.groups, id.Value())
	return nil
}

func (m *MockServerGroupRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.groups)), nil
}

// MockShadowsocksServerRepository is a simple mock for testing
type MockShadowsocksServerRepository struct {
	serversByGroup map[uint]int64
}

func NewMockShadowsocksServerRepository() *MockShadowsocksServerRepository {
	return &MockShadowsocksServerRepository{
		serversByGroup: make(map[uint]int64),
	}
}

func (m *MockShadowsocksServerRepository) CountByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) (int64, error) {
	return m.serversByGroup[groupID.Value()], nil
}

// Implement other required methods with simple stubs
func (m *MockShadowsocksServerRepository) Save(ctx context.Context, server *model.ShadowsocksServer) error { return nil }
func (m *MockShadowsocksServerRepository) FindByID(ctx context.Context, id valueobject.ServerID) (*model.ShadowsocksServer, error) { return nil, nil }
func (m *MockShadowsocksServerRepository) FindByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error) { return nil, nil }
func (m *MockShadowsocksServerRepository) FindVisibleByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error) { return nil, nil }
func (m *MockShadowsocksServerRepository) FindAll(ctx context.Context, filters repository.ShadowsocksServerFilters, limit, offset int) ([]*model.ShadowsocksServer, int64, error) { return nil, 0, nil }
func (m *MockShadowsocksServerRepository) FindVisible(ctx context.Context) ([]*model.ShadowsocksServer, error) { return nil, nil }
func (m *MockShadowsocksServerRepository) Delete(ctx context.Context, id valueobject.ServerID) error { return nil }
func (m *MockShadowsocksServerRepository) Count(ctx context.Context) (int64, error) { return 0, nil }

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	publishedEvents []interface{}
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		publishedEvents: make([]interface{}, 0),
	}
}

func (m *MockEventPublisher) PublishEvents(ctx context.Context, events []interface{}) error {
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *MockEventPublisher) GetPublishedEvents() []interface{} {
	return m.publishedEvents
}

func TestServerGroupCommandHandler_HandleCreateServerGroup(t *testing.T) {
	// Setup
	repo := NewMockServerGroupRepository()
	shadowsocksRepo := NewMockShadowsocksServerRepository()
	eventPublisher := NewMockEventPublisher()
	domainService := service.NewServerGroupDomainService(repo, shadowsocksRepo)
	handler := command.NewServerGroupCommandHandler(repo, domainService, eventPublisher)

	// Test case: successful creation
	t.Run("successful creation", func(t *testing.T) {
		cmd := command.CreateServerGroupCommand{
			Name: "Asia Pacific",
		}

		group, err := handler.HandleCreateServerGroup(context.Background(), cmd)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if group == nil {
			t.Fatal("Expected server group to be created")
		}

		if group.Name().Value() != "Asia Pacific" {
			t.Errorf("Expected name 'Asia Pacific', got '%s'", group.Name().Value())
		}

		if group.ID().IsZero() {
			t.Error("Expected ID to be set after creation")
		}

		// Check if events were published
		events := eventPublisher.GetPublishedEvents()
		if len(events) != 1 {
			t.Errorf("Expected 1 event to be published, got %d", len(events))
		}
	})

	// Test case: duplicate name
	t.Run("duplicate name should fail", func(t *testing.T) {
		cmd := command.CreateServerGroupCommand{
			Name: "Asia Pacific", // Same name as above
		}

		_, err := handler.HandleCreateServerGroup(context.Background(), cmd)

		if err == nil {
			t.Fatal("Expected error for duplicate name")
		}

		if err != command.ErrServerGroupAlreadyExists {
			t.Errorf("Expected ErrServerGroupAlreadyExists, got %v", err)
		}
	})

	// Test case: invalid name
	t.Run("invalid name should fail", func(t *testing.T) {
		cmd := command.CreateServerGroupCommand{
			Name: "", // Empty name
		}

		_, err := handler.HandleCreateServerGroup(context.Background(), cmd)

		if err == nil {
			t.Fatal("Expected error for invalid name")
		}
	})
}

func TestServerGroupCommandHandler_HandleDeleteServerGroup(t *testing.T) {
	// Setup
	repo := NewMockServerGroupRepository()
	shadowsocksRepo := NewMockShadowsocksServerRepository()
	eventPublisher := NewMockEventPublisher()
	domainService := service.NewServerGroupDomainService(repo, shadowsocksRepo)
	handler := command.NewServerGroupCommandHandler(repo, domainService, eventPublisher)

	// Create a server group first
	createCmd := command.CreateServerGroupCommand{Name: "Test Group"}
	group, _ := handler.HandleCreateServerGroup(context.Background(), createCmd)
	
	// Clear published events from creation
	eventPublisher.publishedEvents = make([]interface{}, 0)

	// Test case: successful deletion
	t.Run("successful deletion", func(t *testing.T) {
		deleteCmd := command.DeleteServerGroupCommand{
			ID: group.ID().Value(),
		}

		err := handler.HandleDeleteServerGroup(context.Background(), deleteCmd)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check if events were published
		events := eventPublisher.GetPublishedEvents()
		if len(events) != 1 {
			t.Errorf("Expected 1 event to be published, got %d", len(events))
		}

		// Verify group is deleted from repository
		_, err = repo.FindByID(context.Background(), group.ID())
		if err != repository.ErrServerGroupNotFound {
			t.Error("Expected server group to be deleted from repository")
		}
	})

	// Test case: delete non-existent group
	t.Run("delete non-existent group should fail", func(t *testing.T) {
		deleteCmd := command.DeleteServerGroupCommand{
			ID: 999, // Non-existent ID
		}

		err := handler.HandleDeleteServerGroup(context.Background(), deleteCmd)

		if err == nil {
			t.Fatal("Expected error for non-existent group")
		}
	})

	// Test case: delete group with servers
	t.Run("delete group with servers should fail", func(t *testing.T) {
		// Create another group
		createCmd := command.CreateServerGroupCommand{Name: "Group With Servers"}
		groupWithServers, _ := handler.HandleCreateServerGroup(context.Background(), createCmd)
		
		// Simulate servers in the group
		shadowsocksRepo.serversByGroup[groupWithServers.ID().Value()] = 1

		deleteCmd := command.DeleteServerGroupCommand{
			ID: groupWithServers.ID().Value(),
		}

		err := handler.HandleDeleteServerGroup(context.Background(), deleteCmd)

		if err != command.ErrCannotDeleteServerGroupWithServers {
			t.Errorf("Expected ErrCannotDeleteServerGroupWithServers, got %v", err)
		}
	})
}