package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/order"
)

type orderGORM struct {
	ID              uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"`
	UserID          uuid.UUID       `gorm:"column:user_id;type:uuid"`
	OrderNumber     string          `gorm:"column:order_number"`
	CustomerName    string          `gorm:"column:customer_name"`
	CustomerContact string          `gorm:"column:customer_contact"`
	TotalRevenue    float64         `gorm:"column:total_revenue"`
	TotalHPP        float64         `gorm:"column:total_hpp"`
	TotalProfit     float64         `gorm:"column:total_profit"`
	Status          string          `gorm:"column:status"`
	Notes           string          `gorm:"column:notes"`
	Source          string          `gorm:"column:source"`
	PaymentStatus   string          `gorm:"column:payment_status"`
	StartedAt       *time.Time      `gorm:"column:started_at"`
	CompletedAt     *time.Time      `gorm:"column:completed_at"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
	Items           []orderItemGORM `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
}

func (orderGORM) TableName() string {
	return "orders"
}

type orderItemGORM struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	OrderID          uuid.UUID  `gorm:"column:order_id;type:uuid"`
	ProductID        *uuid.UUID `gorm:"column:product_id;type:uuid"`
	ProductName      string     `gorm:"column:product_name"`
	Quantity         int        `gorm:"column:quantity"`
	SellingPrice     float64    `gorm:"column:selling_price"`
	HPP              float64    `gorm:"column:hpp"`
	WeightGrams      float64    `gorm:"column:weight_grams"`
	PrintTimeHours   float64    `gorm:"column:print_time_hours"`
	MachineID        *uuid.UUID `gorm:"column:machine_id;type:uuid"`
	EnergyCost       float64    `gorm:"column:energy_cost"`
	DepreciationCost float64    `gorm:"column:depreciation_cost"`
	MaintenanceCost  float64    `gorm:"column:maintenance_cost"`
	PackingFee       float64    `gorm:"column:packing_fee"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (orderItemGORM) TableName() string {
	return "order_items"
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func mapOrderGORMToDomain(g *orderGORM) *order.Order {
	if g == nil {
		return nil
	}

	items := make([]order.OrderItem, len(g.Items))
	for i, it := range g.Items {
		items[i] = order.ReconstructOrderItem(
			it.ID,
			it.OrderID,
			it.ProductID,
			it.ProductName,
			it.Quantity,
			it.SellingPrice,
			it.HPP,
			it.WeightGrams,
			it.PrintTimeHours,
			it.MachineID,
			it.EnergyCost,
			it.DepreciationCost,
			it.MaintenanceCost,
			it.PackingFee,
			it.CreatedAt,
		)
	}

	return order.ReconstructOrder(
		g.ID,
		g.UserID,
		g.OrderNumber,
		g.CustomerName,
		g.CustomerContact,
		g.TotalRevenue,
		g.TotalHPP,
		g.TotalProfit,
		g.Status,
		g.Notes,
		g.Source,
		g.PaymentStatus,
		g.StartedAt,
		g.CompletedAt,
		items,
		g.CreatedAt,
		g.UpdatedAt,
	)
}

func (r *OrderRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	var gormOrders []orderGORM
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&gormOrders).Error
	if err != nil {
		return nil, err
	}

	results := make([]*order.Order, len(gormOrders))
	for i := range gormOrders {
		results[i] = mapOrderGORMToDomain(&gormOrders[i])
	}
	return results, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*order.Order, error) {
	var g orderGORM
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND user_id = ?", id, userID).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}
	return mapOrderGORMToDomain(&g), nil
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		g := orderGORM{
			ID:              o.ID(),
			UserID:          o.UserID(),
			OrderNumber:     o.OrderNumber(),
			CustomerName:    o.CustomerName(),
			CustomerContact: o.CustomerContact(),
			TotalRevenue:    o.TotalRevenue(),
			TotalHPP:        o.TotalHPP(),
			TotalProfit:     o.TotalProfit(),
			Status:          o.Status(),
			Notes:           o.Notes(),
			Source:          o.Source(),
			PaymentStatus:   o.PaymentStatus(),
			StartedAt:       o.StartedAt(),
			CompletedAt:     o.CompletedAt(),
			CreatedAt:       o.CreatedAt(),
			UpdatedAt:       o.UpdatedAt(),
		}

		if err := tx.Create(&g).Error; err != nil {
			return err
		}

		for _, it := range o.Items() {
			itemGORM := orderItemGORM{
				ID:               it.ID(),
				OrderID:          o.ID(),
				ProductID:        it.ProductID(),
				ProductName:      it.ProductName(),
				Quantity:         it.Quantity(),
				SellingPrice:     it.SellingPrice(),
				HPP:              it.HPP(),
				WeightGrams:      it.WeightGrams(),
				PrintTimeHours:   it.PrintTimeHours(),
				MachineID:        it.MachineID(),
				EnergyCost:       it.EnergyCost(),
				DepreciationCost: it.DepreciationCost(),
				MaintenanceCost:  it.MaintenanceCost(),
				PackingFee:       it.PackingFee(),
				CreatedAt:        it.CreatedAt(),
			}
			if err := tx.Create(&itemGORM).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, userID, id uuid.UUID, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	if status == "IN_PRODUCTION" {
		updates["started_at"] = now
	}
	if status == "COMPLETED" {
		updates["completed_at"] = now
	}

	res := r.db.WithContext(ctx).Model(&orderGORM{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return order.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error {
	res := r.db.WithContext(ctx).Model(&orderGORM{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"payment_status": paymentStatus,
			"updated_at":     time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return order.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&orderItemGORM{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&orderGORM{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return order.ErrOrderNotFound
		}
		return nil
	})
}
