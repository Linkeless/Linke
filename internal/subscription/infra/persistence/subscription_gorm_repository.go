package persistence

import (
	"context"
	"fmt"
	"time"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/repository"
	"linke/internal/subscription/domain/valueobject"

	"gorm.io/gorm"
)

type SubscriptionGormRepository struct {
	db     *gorm.DB
	mapper *SubscriptionMapper
}

func NewSubscriptionGormRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &SubscriptionGormRepository{
		db:     db,
		mapper: NewSubscriptionMapper(),
	}
}

func (r *SubscriptionGormRepository) Save(ctx context.Context, subscription *model.Subscription) error {
	po, err := r.mapper.DomainToPO(subscription)
	if err != nil {
		return fmt.Errorf("failed to map domain to PO: %w", err)
	}

	if po.ID == 0 {
		if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
		
		subscriptionID, _ := valueobject.NewSubscriptionID(po.ID)
		subscription.SetID(*subscriptionID)
	} else {
		if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

func (r *SubscriptionGormRepository) FindByID(ctx context.Context, id valueobject.SubscriptionID) (*model.Subscription, error) {
	var po SubscriptionPO
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to find subscription by ID: %w", err)
	}

	return r.mapper.POToDomain(&po)
}

func (r *SubscriptionGormRepository) FindByUUID(ctx context.Context, uuid valueobject.SubscriptionUUID) (*model.Subscription, error) {
	var po SubscriptionPO
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid.Value()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to find subscription by UUID: %w", err)
	}

	return r.mapper.POToDomain(&po)
}

func (r *SubscriptionGormRepository) FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*model.Subscription, error) {
	var pos []SubscriptionPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find subscriptions by user ID: %w", err)
	}

	subscriptions := make([]*model.Subscription, len(pos))
	for i, po := range pos {
		subscription, err := r.mapper.POToDomain(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to map PO to domain: %w", err)
		}
		subscriptions[i] = subscription
	}

	return subscriptions, nil
}

func (r *SubscriptionGormRepository) FindActiveByUserID(ctx context.Context, userID valueobject.UserID) ([]*model.Subscription, error) {
	var pos []SubscriptionPO
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > ?", userID.Value(), "active", now).
		Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find active subscriptions by user ID: %w", err)
	}

	subscriptions := make([]*model.Subscription, len(pos))
	for i, po := range pos {
		subscription, err := r.mapper.POToDomain(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to map PO to domain: %w", err)
		}
		subscriptions[i] = subscription
	}

	return subscriptions, nil
}

func (r *SubscriptionGormRepository) FindExpiringBefore(ctx context.Context, before time.Time) ([]*model.Subscription, error) {
	var pos []SubscriptionPO
	if err := r.db.WithContext(ctx).
		Where("status = ? AND end_date < ?", "active", before).
		Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}

	subscriptions := make([]*model.Subscription, len(pos))
	for i, po := range pos {
		subscription, err := r.mapper.POToDomain(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to map PO to domain: %w", err)
		}
		subscriptions[i] = subscription
	}

	return subscriptions, nil
}

func (r *SubscriptionGormRepository) FindPendingRenewal(ctx context.Context) ([]*model.Subscription, error) {
	var pos []SubscriptionPO
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("status = ? AND auto_renew = ? AND next_billing_date IS NOT NULL AND next_billing_date <= ?", 
			"active", true, now.Add(24*time.Hour)).
		Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find pending renewal subscriptions: %w", err)
	}

	subscriptions := make([]*model.Subscription, len(pos))
	for i, po := range pos {
		subscription, err := r.mapper.POToDomain(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to map PO to domain: %w", err)
		}
		subscriptions[i] = subscription
	}

	return subscriptions, nil
}

func (r *SubscriptionGormRepository) Delete(ctx context.Context, id valueobject.SubscriptionID) error {
	if err := r.db.WithContext(ctx).Delete(&SubscriptionPO{}, id.Value()).Error; err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionGormRepository) Count(ctx context.Context, filters *repository.SubscriptionFilters) (int64, error) {
	query := r.db.WithContext(ctx).Model(&SubscriptionPO{})
	query = r.applyFilters(query, filters)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	return count, nil
}

func (r *SubscriptionGormRepository) FindWithFilters(ctx context.Context, filters *repository.SubscriptionFilters) ([]*model.Subscription, error) {
	query := r.db.WithContext(ctx).Model(&SubscriptionPO{})
	query = r.applyFilters(query, filters)

	if filters.SortBy != "" {
		sortOrder := "DESC"
		if filters.SortOrder == "asc" {
			sortOrder = "ASC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filters.SortBy, sortOrder))
	} else {
		query = query.Order("created_at DESC")
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var pos []SubscriptionPO
	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find subscriptions with filters: %w", err)
	}

	subscriptions := make([]*model.Subscription, len(pos))
	for i, po := range pos {
		subscription, err := r.mapper.POToDomain(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to map PO to domain: %w", err)
		}
		subscriptions[i] = subscription
	}

	return subscriptions, nil
}

func (r *SubscriptionGormRepository) applyFilters(query *gorm.DB, filters *repository.SubscriptionFilters) *gorm.DB {
	if filters.UserID != nil {
		query = query.Where("user_id = ?", filters.UserID.Value())
	}

	if filters.PlanID != nil {
		query = query.Where("plan_id = ?", filters.PlanID.Value())
	}

	if filters.Status != nil {
		query = query.Where("status = ?", filters.Status.String())
	}

	if filters.Currency != nil {
		query = query.Where("currency = ?", filters.Currency.String())
	}

	if filters.AutoRenew != nil {
		query = query.Where("auto_renew = ?", *filters.AutoRenew)
	}

	if filters.StartDate != nil {
		query = query.Where("start_date >= ?", *filters.StartDate)
	}

	if filters.EndDate != nil {
		query = query.Where("end_date <= ?", filters.EndDate.Add(24*time.Hour))
	}

	now := time.Now()
	if filters.InTrial != nil {
		if *filters.InTrial {
			query = query.Where("trial_end_date IS NOT NULL AND trial_end_date > ?", now)
		} else {
			query = query.Where("trial_end_date IS NULL OR trial_end_date <= ?", now)
		}
	}

	if filters.Expired != nil {
		if *filters.Expired {
			query = query.Where("end_date < ? OR status = ?", now, "expired")
		} else {
			query = query.Where("end_date >= ? AND status != ?", now, "expired")
		}
	}

	if filters.Overdue != nil && *filters.Overdue {
		query = query.Where("next_billing_date < ? AND auto_renew = ? AND status = ?", 
			now.AddDate(0, 0, -7), true, "active")
	}

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("uuid LIKE ? OR notes LIKE ?", searchPattern, searchPattern)
	}

	return query
}