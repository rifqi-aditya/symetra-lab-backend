package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/component"
)

type componentGORM struct {
	ID                   uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID               uuid.UUID `gorm:"column:user_id;type:uuid"`
	Name                 string    `gorm:"column:name"`
	PricePerUnit         float64   `gorm:"column:price_per_unit"`
	DefaultMarkupPercent float64   `gorm:"column:default_markup_percent"`
	Description          *string   `gorm:"column:description"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (componentGORM) TableName() string {
	return "components"
}

type ComponentRepository struct {
	db *gorm.DB
}

func NewComponentRepository(db *gorm.DB) *ComponentRepository {
	return &ComponentRepository{db: db}
}

func mapComponentGORMToDomain(c *componentGORM) *component.Component {
	if c == nil {
		return nil
	}
	return component.ReconstructComponent(
		c.ID,
		c.UserID,
		c.Name,
		c.PricePerUnit,
		c.DefaultMarkupPercent,
		c.Description,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func (r *ComponentRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*component.Component, error) {
	var list []componentGORM
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]*component.Component, len(list))
	for i, c := range list {
		cCopy := c
		res[i] = mapComponentGORMToDomain(&cCopy)
	}
	return res, nil
}

func (r *ComponentRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*component.Component, error) {
	var c componentGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return mapComponentGORMToDomain(&c), nil
}

func (r *ComponentRepository) Create(ctx context.Context, c *component.Component) error {
	m := componentGORM{
		ID:                   c.ID(),
		UserID:               c.UserID(),
		Name:                 c.Name(),
		PricePerUnit:         c.PricePerUnit(),
		DefaultMarkupPercent: c.DefaultMarkupPercent(),
		Description:          c.Description(),
		CreatedAt:            c.CreatedAt(),
		UpdatedAt:            c.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *ComponentRepository) Update(ctx context.Context, c *component.Component) error {
	return r.db.WithContext(ctx).Model(&componentGORM{}).
		Where("id = ?", c.ID()).
		Updates(map[string]interface{}{
			"name":                   c.Name(),
			"price_per_unit":         c.PricePerUnit(),
			"default_markup_percent": c.DefaultMarkupPercent(),
			"description":            c.Description(),
			"updated_at":             c.UpdatedAt(),
		}).Error
}

func (r *ComponentRepository) Delete(ctx context.Context, userID, id uuid.UUID, force bool) error {
	if !force {
		var usageCount int64
		if err := r.db.WithContext(ctx).Table("product_components").Where("component_id = ?", id).Count(&usageCount).Error; err == nil && usageCount > 0 {
			return fmt.Errorf("komponen tidak dapat dihapus karena digunakan dalam %d resep produk", usageCount)
		}
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&componentGORM{}).Error
}