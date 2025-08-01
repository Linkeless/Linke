package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
)

// PaymentConfigGormRepository implements PaymentConfigRepository using GORM
type PaymentConfigGormRepository struct {
	db     *gorm.DB
	mapper *PaymentConfigMapper
}

// NewPaymentConfigGormRepository creates a new PaymentConfigGormRepository
func NewPaymentConfigGormRepository(db *gorm.DB) repository.PaymentConfigRepository {
	return &PaymentConfigGormRepository{
		db:     db,
		mapper: NewPaymentConfigMapper(),
	}
}

// Save saves a payment config aggregate
func (r *PaymentConfigGormRepository) Save(ctx context.Context, config *aggregate.PaymentConfig) error {
	po, err := r.mapper.ToPersistentObject(config)
	if err != nil {
		return fmt.Errorf("failed to convert payment config to persistent object: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return fmt.Errorf("failed to save payment config: %w", err)
	}

	return nil
}

// Update updates a payment config aggregate
func (r *PaymentConfigGormRepository) Update(ctx context.Context, config *aggregate.PaymentConfig) error {
	po, err := r.mapper.ToPersistentObject(config)
	if err != nil {
		return fmt.Errorf("failed to convert payment config to persistent object: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return fmt.Errorf("failed to update payment config: %w", err)
	}

	return nil
}

// FindByID finds a payment config by its ID
func (r *PaymentConfigGormRepository) FindByID(ctx context.Context, id valueobject.PaymentConfigID) (*aggregate.PaymentConfig, error) {
	var po PaymentConfigPO
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found with ID: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find payment config by ID: %w", err)
	}

	return r.mapper.ToAggregate(&po)
}

// FindByGateway finds a payment config by gateway
func (r *PaymentConfigGormRepository) FindByGateway(ctx context.Context, gateway valueobject.PaymentGateway) (*aggregate.PaymentConfig, error) {
	var po PaymentConfigPO
	if err := r.db.WithContext(ctx).Where("gateway = ?", gateway.Value()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found with gateway: %s", gateway.String())
		}
		return nil, fmt.Errorf("failed to find payment config by gateway: %w", err)
	}

	return r.mapper.ToAggregate(&po)
}

// FindAll finds all payment configs
func (r *PaymentConfigGormRepository) FindAll(ctx context.Context) ([]*aggregate.PaymentConfig, error) {
	var pos []*PaymentConfigPO
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find all payment configs: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindActive finds all active payment configs
func (r *PaymentConfigGormRepository) FindActive(ctx context.Context) ([]*aggregate.PaymentConfig, error) {
	var pos []*PaymentConfigPO
	if err := r.db.WithContext(ctx).Where("is_enabled = ?", true).Order("sort_order ASC, created_at ASC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find active payment configs: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindByCurrency finds payment configs that support a specific currency
func (r *PaymentConfigGormRepository) FindByCurrency(ctx context.Context, currency valueobject.Currency) ([]*aggregate.PaymentConfig, error) {
	var pos []*PaymentConfigPO
	// Use JSON_CONTAINS for MySQL or similar function for other databases
	// For simplicity, using LIKE pattern matching
	currencyPattern := fmt.Sprintf("%%%q%%", currency.Code())
	
	if err := r.db.WithContext(ctx).Where("supported_currencies LIKE ?", currencyPattern).Order("sort_order ASC, created_at ASC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find payment configs by currency: %w", err)
	}

	// Filter results in application to ensure exact match
	configs, err := r.mapper.ToAggregateList(pos)
	if err != nil {
		return nil, err
	}

	var filteredConfigs []*aggregate.PaymentConfig
	for _, config := range configs {
		if config.SupportsCurrency(currency) {
			filteredConfigs = append(filteredConfigs, config)
		}
	}

	return filteredConfigs, nil
}

// FindByMethod finds payment configs that support a specific payment method
func (r *PaymentConfigGormRepository) FindByMethod(ctx context.Context, method valueobject.PaymentMethod) ([]*aggregate.PaymentConfig, error) {
	var pos []*PaymentConfigPO
	// Use JSON_CONTAINS for MySQL or similar function for other databases
	// For simplicity, using LIKE pattern matching
	methodPattern := fmt.Sprintf("%%%q%%", method.Value())
	
	if err := r.db.WithContext(ctx).Where("supported_methods LIKE ?", methodPattern).Order("sort_order ASC, created_at ASC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find payment configs by method: %w", err)
	}

	// Filter results in application to ensure exact match
	configs, err := r.mapper.ToAggregateList(pos)
	if err != nil {
		return nil, err
	}

	var filteredConfigs []*aggregate.PaymentConfig
	for _, config := range configs {
		if config.SupportsMethod(method) {
			filteredConfigs = append(filteredConfigs, config)
		}
	}

	return filteredConfigs, nil
}

// FindWithFilters finds payment configs with filters
func (r *PaymentConfigGormRepository) FindWithFilters(ctx context.Context, filters repository.PaymentConfigFilters) ([]*aggregate.PaymentConfig, int64, error) {
	query := r.db.WithContext(ctx).Model(&PaymentConfigPO{})

	// Apply filters
	if filters.Gateway != nil {
		query = query.Where("gateway = ?", filters.Gateway.Value())
	}

	if filters.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filters.IsEnabled)
	}

	if filters.Currency != nil {
		currencyPattern := fmt.Sprintf("%%%q%%", filters.Currency.Code())
		query = query.Where("supported_currencies LIKE ?", currencyPattern)
	}

	if filters.Method != nil {
		methodPattern := fmt.Sprintf("%%%q%%", filters.Method.Value())
		query = query.Where("supported_methods LIKE ?", methodPattern)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment configs: %w", err)
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "sort_order"
	}

	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}

	validSortFields := map[string]bool{
		"sort_order":  true,
		"created_at":  true,
		"updated_at":  true,
		"gateway":     true,
		"name":        true,
		"is_enabled":  true,
	}

	if !validSortFields[sortBy] {
		sortBy = "sort_order"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply secondary sort for consistency
	if sortBy != "created_at" {
		query = query.Order("created_at ASC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var pos []*PaymentConfigPO
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find payment configs: %w", err)
	}

	configs, err := r.mapper.ToAggregateList(pos)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert payment configs: %w", err)
	}

	// Apply additional filtering in application if needed
	if filters.Currency != nil || filters.Method != nil {
		var filteredConfigs []*aggregate.PaymentConfig
		for _, config := range configs {
			include := true
			
			if filters.Currency != nil && !config.SupportsCurrency(*filters.Currency) {
				include = false
			}
			
			if filters.Method != nil && !config.SupportsMethod(*filters.Method) {
				include = false
			}
			
			if include {
				filteredConfigs = append(filteredConfigs, config)
			}
		}
		configs = filteredConfigs
		totalCount = int64(len(configs))
	}

	return configs, totalCount, nil
}

// Delete soft deletes a payment config
func (r *PaymentConfigGormRepository) Delete(ctx context.Context, id valueobject.PaymentConfigID) error {
	if err := r.db.WithContext(ctx).Delete(&PaymentConfigPO{}, id.Value()).Error; err != nil {
		return fmt.Errorf("failed to delete payment config: %w", err)
	}

	return nil
}

// Exists checks if a payment config exists by ID
func (r *PaymentConfigGormRepository) Exists(ctx context.Context, id valueobject.PaymentConfigID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentConfigPO{}).Where("id = ?", id.Value()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment config existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByGateway checks if a payment config exists by gateway
func (r *PaymentConfigGormRepository) ExistsByGateway(ctx context.Context, gateway valueobject.PaymentGateway) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentConfigPO{}).Where("gateway = ?", gateway.Value()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment config gateway existence: %w", err)
	}

	return count > 0, nil
}

// Count returns the total count of payment configs
func (r *PaymentConfigGormRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentConfigPO{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payment configs: %w", err)
	}

	return count, nil
}

// CountActive returns the count of active payment configs
func (r *PaymentConfigGormRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentConfigPO{}).Where("is_enabled = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active payment configs: %w", err)
	}

	return count, nil
}