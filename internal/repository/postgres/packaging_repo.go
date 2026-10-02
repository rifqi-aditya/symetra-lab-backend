package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/packaging"
)

type packagingItemGORM struct {
	ID               uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID           uuid.UUID `gorm:"column:user_id;type:uuid"`
	Name             string    `gorm:"column:name"`
	Category         string    `gorm:"column:category"`
	UnitType         string    `gorm:"column:unit_type"`
	PurchasePrice    float64   `gorm:"column:purchase_price"`
	PurchaseQuantity float64   `gorm:"column:purchase_quantity"`
	UnitCost         float64   `gorm:"column:unit_cost"`
	StockQuantity    float64   `gorm:"column:stock_quantity"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (packagingItemGORM) TableName() string {
	return "packaging_items"
}

type packagingPresetItemGORM struct {
	ID              uuid.UUID          `gorm:"column:id;primaryKey;type:uuid"`
	PresetID        uuid.UUID          `gorm:"column:preset_id;type:uuid"`
	PackagingItemID uuid.UUID          `gorm:"column:packaging_item_id;type:uuid"`
	QuantityUsed    float64            `gorm:"column:quantity_used"`
	PackagingItem   *packagingItemGORM `gorm:"foreignKey:PackagingItemID"`
	CreatedAt       time.Time          `gorm:"column:created_at"`
}

func (packagingPresetItemGORM) TableName() string {
	return "packaging_preset_items"
}

type packagingPresetGORM struct {
	ID          uuid.UUID                 `gorm:"column:id;primaryKey;type:uuid"`
	UserID      uuid.UUID                 `gorm:"column:user_id;type:uuid"`
	Name        string                    `gorm:"column:name"`
	Description *string                   `gorm:"column:description"`
	Items       []packagingPresetItemGORM `gorm:"foreignKey:PresetID"`
	CreatedAt   time.Time                 `gorm:"column:created_at"`
	UpdatedAt   time.Time                 `gorm:"column:updated_at"`
}

func (packagingPresetGORM) TableName() string {
	return "packaging_presets"
}

type PackagingRepository struct {
	db *gorm.DB
}

func NewPackagingRepository(db *gorm.DB) *PackagingRepository {
	return &PackagingRepository{db: db}
}

func mapPackagingItemGORMToDomain(p *packagingItemGORM) *packaging.PackagingItem {
	if p == nil {
		return nil
	}
	return packaging.ReconstructPackagingItem(
		p.ID,
		p.UserID,
		p.Name,
		p.Category,
		p.UnitType,
		p.PurchasePrice,
		p.PurchaseQuantity,
		p.UnitCost,
		p.StockQuantity,
		p.CreatedAt,
		p.UpdatedAt,
	)
}

func (r *PackagingRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*packaging.PackagingItem, error) {
	var list []packagingItemGORM
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]*packaging.PackagingItem, len(list))
	for i, p := range list {
		pCopy := p
		res[i] = mapPackagingItemGORMToDomain(&pCopy)
	}
	return res, nil
}

func (r *PackagingRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*packaging.PackagingItem, error) {
	var p packagingItemGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return mapPackagingItemGORMToDomain(&p), nil
}

func (r *PackagingRepository) Create(ctx context.Context, p *packaging.PackagingItem) error {
	m := packagingItemGORM{
		ID:               p.ID(),
		UserID:           p.UserID(),
		Name:             p.Name(),
		Category:         p.Category(),
		UnitType:         p.UnitType(),
		PurchasePrice:    p.PurchasePrice(),
		PurchaseQuantity: p.PurchaseQuantity(),
		UnitCost:         p.UnitCost(),
		StockQuantity:    p.StockQuantity(),
		CreatedAt:        p.CreatedAt(),
		UpdatedAt:        p.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *PackagingRepository) Update(ctx context.Context, p *packaging.PackagingItem) error {
	return r.db.WithContext(ctx).Model(&packagingItemGORM{}).
		Where("id = ?", p.ID()).
		Updates(map[string]interface{}{
			"name":              p.Name(),
			"category":          p.Category(),
			"unit_type":         p.UnitType(),
			"purchase_price":    p.PurchasePrice(),
			"purchase_quantity": p.PurchaseQuantity(),
			"unit_cost":         p.UnitCost(),
			"stock_quantity":    p.StockQuantity(),
			"updated_at":        p.UpdatedAt(),
		}).Error
}

func (r *PackagingRepository) Delete(ctx context.Context, userID, id uuid.UUID, force bool) error {
	if !force {
		var presetUsage int64
		r.db.WithContext(ctx).Table("packaging_preset_items").Where("packaging_item_id = ?", id).Count(&presetUsage)
		if presetUsage > 0 {
			return fmt.Errorf("packaging item sedang digunakan di %d preset", presetUsage)
		}
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&packagingItemGORM{}).Error
}

func (r *PackagingRepository) FindAllPresets(ctx context.Context, userID uuid.UUID) ([]*packaging.PackagingPreset, error) {
	var list []packagingPresetGORM
	err := r.db.WithContext(ctx).Preload("Items.PackagingItem").Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	res := make([]*packaging.PackagingPreset, len(list))
	for i, preset := range list {
		items := make([]*packaging.PackagingPresetItem, len(preset.Items))
		for j, it := range preset.Items {
			items[j] = packaging.ReconstructPresetItem(
				it.ID,
				it.PresetID,
				it.PackagingItemID,
				it.QuantityUsed,
				mapPackagingItemGORMToDomain(it.PackagingItem),
				it.CreatedAt,
			)
		}
		res[i] = packaging.ReconstructPackagingPreset(
			preset.ID,
			preset.UserID,
			preset.Name,
			preset.Description,
			items,
			preset.CreatedAt,
			preset.UpdatedAt,
		)
	}
	return res, nil
}

func (r *PackagingRepository) FindPresetByID(ctx context.Context, userID, id uuid.UUID) (*packaging.PackagingPreset, error) {
	var preset packagingPresetGORM
	err := r.db.WithContext(ctx).Preload("Items.PackagingItem").Where("id = ?", id).First(&preset).Error
	if err != nil {
		return nil, err
	}
	items := make([]*packaging.PackagingPresetItem, len(preset.Items))
	for j, it := range preset.Items {
		items[j] = packaging.ReconstructPresetItem(
			it.ID,
			it.PresetID,
			it.PackagingItemID,
			it.QuantityUsed,
			mapPackagingItemGORMToDomain(it.PackagingItem),
			it.CreatedAt,
		)
	}
	return packaging.ReconstructPackagingPreset(
		preset.ID,
		preset.UserID,
		preset.Name,
		preset.Description,
		items,
		preset.CreatedAt,
		preset.UpdatedAt,
	), nil
}

func (r *PackagingRepository) CreatePreset(ctx context.Context, p *packaging.PackagingPreset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		presetG := packagingPresetGORM{
			ID:          p.ID(),
			UserID:      p.UserID(),
			Name:        p.Name(),
			Description: p.Description(),
			CreatedAt:   p.CreatedAt(),
			UpdatedAt:   p.UpdatedAt(),
		}
		if err := tx.Create(&presetG).Error; err != nil {
			return err
		}

		for _, it := range p.Items() {
			itemG := packagingPresetItemGORM{
				ID:              it.ID(),
				PresetID:        p.ID(),
				PackagingItemID: it.PackagingItemID(),
				QuantityUsed:    it.QuantityUsed(),
				CreatedAt:       it.CreatedAt(),
			}
			if err := tx.Create(&itemG).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PackagingRepository) UpdatePreset(ctx context.Context, p *packaging.PackagingPreset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&packagingPresetGORM{}).Where("id = ?", p.ID()).Updates(map[string]interface{}{
			"name":        p.Name(),
			"description": p.Description(),
			"updated_at":  p.UpdatedAt(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("preset_id = ?", p.ID()).Delete(&packagingPresetItemGORM{}).Error; err != nil {
			return err
		}

		for _, it := range p.Items() {
			itemG := packagingPresetItemGORM{
				ID:              it.ID(),
				PresetID:        p.ID(),
				PackagingItemID: it.PackagingItemID(),
				QuantityUsed:    it.QuantityUsed(),
				CreatedAt:       it.CreatedAt(),
			}
			if err := tx.Create(&itemG).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PackagingRepository) DeletePreset(ctx context.Context, userID, id uuid.UUID, force bool) error {
	if !force {
		var productCount int64
		r.db.WithContext(ctx).Table("products").Where("packaging_preset_id = ?", id).Count(&productCount)
		if productCount > 0 {
			return fmt.Errorf("preset kemasan sedang digunakan oleh %d produk", productCount)
		}
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("preset_id = ?", id).Delete(&packagingPresetItemGORM{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&packagingPresetGORM{}).Error
	})
}