package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/product"
)

type productComponentGORM struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	ProductID     uuid.UUID `gorm:"column:product_id;type:uuid"`
	ComponentID   *uuid.UUID `gorm:"column:component_id;type:uuid"`
	Quantity      float64   `gorm:"column:quantity"`
	MarkupPercent float64   `gorm:"column:markup_percent"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (productComponentGORM) TableName() string {
	return "product_components"
}

type productPackagingItemGORM struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	ProductID       uuid.UUID `gorm:"column:product_id;type:uuid"`
	PackagingItemID uuid.UUID `gorm:"column:packaging_item_id;type:uuid"`
	QuantityUsed    float64   `gorm:"column:quantity_used"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (productPackagingItemGORM) TableName() string {
	return "product_packaging_items"
}

type productGORM struct {
	ID                    uuid.UUID                  `gorm:"column:id;primaryKey;type:uuid"`
	UserID                uuid.UUID                  `gorm:"column:user_id;type:uuid"`
	Name                  string                     `gorm:"column:name"`
	ParentSKU             *string                    `gorm:"column:parent_sku"`
	SKU                   *string                    `gorm:"column:sku"`
	Description           *string                    `gorm:"column:description"`
	Category              string                     `gorm:"column:category"`
	ThumbnailURL          *string                    `gorm:"column:thumbnail_url"`
	DesignLink            *string                    `gorm:"column:design_link"`
	DefaultWeightGrams    float64                    `gorm:"column:default_weight_grams"`
	MaterialType          string                     `gorm:"column:material_type"`
	DefaultPrintTimeHours float64                    `gorm:"column:default_print_time_hours"`
	DefaultMachineID      *uuid.UUID                 `gorm:"column:default_machine_id;type:uuid"`
	PackagingPresetID     *uuid.UUID                 `gorm:"column:packaging_preset_id;type:uuid"`
	BatchSize             int                        `gorm:"column:batch_size"`
	PackingFeeIDR         int                        `gorm:"column:packing_fee_idr"`
	BaseHPP               float64                    `gorm:"column:base_hpp"`
	BaseSellingPrice      float64                    `gorm:"column:base_selling_price"`
	TargetMarginPercent   int                        `gorm:"column:target_margin_percent"`
	TimesOrdered          int                        `gorm:"column:times_ordered"`
	Components            []productComponentGORM     `gorm:"foreignKey:ProductID"`
	PackagingItems        []productPackagingItemGORM `gorm:"foreignKey:ProductID"`
	CreatedAt             time.Time                  `gorm:"column:created_at"`
	UpdatedAt             time.Time                  `gorm:"column:updated_at"`
}

func (productGORM) TableName() string {
	return "products"
}

func toProductEntity(g *productGORM) *product.Product {
	if g == nil {
		return nil
	}

	comps := make([]*product.ProductComponent, len(g.Components))
	for i, c := range g.Components {
		comps[i] = product.ReconstructProductComponent(c.ID, c.ProductID, c.ComponentID, c.Quantity, c.MarkupPercent, c.CreatedAt)
	}

	packs := make([]*product.ProductPackagingItem, len(g.PackagingItems))
	for i, p := range g.PackagingItems {
		packs[i] = product.ReconstructProductPackagingItem(p.ID, p.ProductID, p.PackagingItemID, p.QuantityUsed, p.CreatedAt)
	}

	return product.ReconstructProduct(
		g.ID,
		g.UserID,
		g.Name,
		g.ParentSKU,
		g.SKU,
		g.Description,
		g.Category,
		g.ThumbnailURL,
		g.DesignLink,
		g.DefaultWeightGrams,
		g.MaterialType,
		g.DefaultPrintTimeHours,
		g.DefaultMachineID,
		g.PackagingPresetID,
		g.BatchSize,
		g.PackingFeeIDR,
		g.BaseHPP,
		g.BaseSellingPrice,
		g.TargetMarginPercent,
		g.TimesOrdered,
		comps,
		packs,
		g.CreatedAt,
		g.UpdatedAt,
	)
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindAll(ctx context.Context, userID uuid.UUID, filter product.Filter) ([]*product.Product, int64, error) {
	var gormProducts []productGORM
	var total int64

	query := r.db.WithContext(ctx).Model(&productGORM{})

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR sku ILIKE ?", s, s)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.ParentSKU != "" {
		query = query.Where("parent_sku = ?", filter.ParentSKU)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&gormProducts).Error; err != nil {
		return nil, 0, err
	}

	entities := make([]*product.Product, len(gormProducts))
	for i := range gormProducts {
		entities[i] = toProductEntity(&gormProducts[i])
	}

	return entities, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*product.Product, error) {
	var g productGORM
	err := r.db.WithContext(ctx).Preload("Components").Preload("PackagingItems").Where("id = ?", id).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toProductEntity(&g), nil
}

func (r *ProductRepository) FindBySKU(ctx context.Context, userID uuid.UUID, sku string) (*product.Product, error) {
	var g productGORM
	err := r.db.WithContext(ctx).Preload("Components").Preload("PackagingItems").Where("sku = ?", sku).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toProductEntity(&g), nil
}

func (r *ProductRepository) FindAllWithoutSKU(ctx context.Context, userID uuid.UUID) ([]*product.Product, error) {
	var gormProducts []productGORM
	err := r.db.WithContext(ctx).Where("sku IS NULL OR sku = ''").Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}
	entities := make([]*product.Product, len(gormProducts))
	for i := range gormProducts {
		entities[i] = toProductEntity(&gormProducts[i])
	}
	return entities, nil
}

func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		g := &productGORM{
			ID:                    p.ID(),
			UserID:                p.UserID(),
			Name:                  p.Name(),
			ParentSKU:             p.ParentSKU(),
			SKU:                   p.SKU(),
			Description:           p.Description(),
			Category:              p.Category(),
			ThumbnailURL:          p.ThumbnailURL(),
			DesignLink:            p.DesignLink(),
			DefaultWeightGrams:    p.DefaultWeightGrams(),
			MaterialType:          p.MaterialType(),
			DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
			DefaultMachineID:      p.DefaultMachineID(),
			PackagingPresetID:     p.PackagingPresetID(),
			BatchSize:             p.BatchSize(),
			PackingFeeIDR:         p.PackingFeeIDR(),
			BaseHPP:               p.BaseHPP(),
			BaseSellingPrice:      p.BaseSellingPrice(),
			TargetMarginPercent:   p.TargetMarginPercent(),
			TimesOrdered:          p.TimesOrdered(),
			CreatedAt:             p.CreatedAt(),
			UpdatedAt:             p.UpdatedAt(),
		}

		if err := tx.Create(g).Error; err != nil {
			return err
		}

		for _, comp := range p.Components() {
			cModel := productComponentGORM{
				ID:            comp.ID(),
				ProductID:     p.ID(),
				ComponentID:   comp.ComponentID(),
				Quantity:      comp.Quantity(),
				MarkupPercent: comp.MarkupPercent(),
				CreatedAt:     comp.CreatedAt(),
			}
			if err := tx.Create(&cModel).Error; err != nil {
				return err
			}
		}

		for _, pack := range p.PackagingItems() {
			pModel := productPackagingItemGORM{
				ID:              pack.ID(),
				ProductID:       p.ID(),
				PackagingItemID: pack.PackagingItemID(),
				QuantityUsed:    pack.QuantityUsed(),
				CreatedAt:       pack.CreatedAt(),
			}
			if err := tx.Create(&pModel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) Update(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":                     p.Name(),
			"parent_sku":              p.ParentSKU(),
			"sku":                     p.SKU(),
			"description":             p.Description(),
			"category":                p.Category(),
			"thumbnail_url":           p.ThumbnailURL(),
			"design_link":             p.DesignLink(),
			"default_weight_grams":    p.DefaultWeightGrams(),
			"material_type":           p.MaterialType(),
			"default_print_time_hours": p.DefaultPrintTimeHours(),
			"default_machine_id":      p.DefaultMachineID(),
			"packaging_preset_id":     p.PackagingPresetID(),
			"batch_size":              p.BatchSize(),
			"packing_fee_idr":         p.PackingFeeIDR(),
			"base_hpp":                p.BaseHPP(),
			"base_selling_price":      p.BaseSellingPrice(),
			"target_margin_percent":   p.TargetMarginPercent(),
			"updated_at":              p.UpdatedAt(),
		}

		if err := tx.Model(&productGORM{}).Where("id = ?", p.ID()).Updates(updates).Error; err != nil {
			return err
		}

		// Replace components
		if err := tx.Where("product_id = ?", p.ID()).Delete(&productComponentGORM{}).Error; err != nil {
			return err
		}
		for _, comp := range p.Components() {
			cModel := productComponentGORM{
				ID:            comp.ID(),
				ProductID:     p.ID(),
				ComponentID:   comp.ComponentID(),
				Quantity:      comp.Quantity(),
				MarkupPercent: comp.MarkupPercent(),
				CreatedAt:     comp.CreatedAt(),
			}
			if err := tx.Create(&cModel).Error; err != nil {
				return err
			}
		}

		// Replace packaging items
		if err := tx.Where("product_id = ?", p.ID()).Delete(&productPackagingItemGORM{}).Error; err != nil {
			return err
		}
		for _, pack := range p.PackagingItems() {
			pModel := productPackagingItemGORM{
				ID:              pack.ID(),
				ProductID:       p.ID(),
				PackagingItemID: pack.PackagingItemID(),
				QuantityUsed:    pack.QuantityUsed(),
				CreatedAt:       pack.CreatedAt(),
			}
			if err := tx.Create(&pModel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	// Check order items existence
	var count int64
	r.db.WithContext(ctx).Table("order_items").Where("product_id = ?", id).Count(&count)
	if count > 0 {
		return product.ErrProductInUseInOrder
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", id).Delete(&productComponentGORM{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", id).Delete(&productPackagingItemGORM{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", id).Delete(&productGORM{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return product.ErrNotFound
		}
		return nil
	})
}

func (r *ProductRepository) GetTotalSoldMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	type SoldResult struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
		TotalSold int       `gorm:"column:total_sold"`
	}

	result := make(map[uuid.UUID]int)

	var shopeeSold []SoldResult
	shopeeQuery := `
		SELECT soi.product_id, COALESCE(SUM(soi.quantity), 0)::integer as total_sold
		FROM shopee_order_items soi
		JOIN shopee_orders so ON so.order_sn = soi.order_sn
		WHERE so.order_status NOT IN ('CANCELLED', 'IN_CANCEL')
		  AND soi.product_id IS NOT NULL
		GROUP BY soi.product_id
	`
	if err := r.db.WithContext(ctx).Raw(shopeeQuery).Scan(&shopeeSold).Error; err == nil {
		for _, s := range shopeeSold {
			result[s.ProductID] += s.TotalSold
		}
	}

	var manualSold []SoldResult
	manualQuery := `
		SELECT oi.product_id, COALESCE(SUM(oi.quantity), 0)::integer as total_sold
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status != 'CANCELLED'
		  AND oi.product_id IS NOT NULL
		GROUP BY oi.product_id
	`
	if err := r.db.WithContext(ctx).Raw(manualQuery).Scan(&manualSold).Error; err == nil {
		for _, m := range manualSold {
			result[m.ProductID] += m.TotalSold
		}
	}

	return result, nil
}

func (r *ProductRepository) GetSoldQtyByID(ctx context.Context, userID, productID uuid.UUID) (int, error) {
	m, err := r.GetTotalSoldMap(ctx, userID)
	if err != nil {
		return 0, err
	}
	return m[productID], nil
}