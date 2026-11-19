package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/product"
)

type productComponentGORM struct {
	ID            uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	ProductID     uuid.UUID  `gorm:"column:product_id;type:uuid"`
	ComponentID   *uuid.UUID `gorm:"column:component_id;type:uuid"`
	Quantity      float64    `gorm:"column:quantity"`
	MarkupPercent float64    `gorm:"column:markup_percent"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
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

type parentProductGORM struct {
	ID           uuid.UUID            `gorm:"column:id;primaryKey;type:uuid"`
	UserID       uuid.UUID            `gorm:"column:user_id;type:uuid"`
	ParentSKU    string               `gorm:"column:parent_sku"`
	Name         string               `gorm:"column:name"`
	CategoryID   *uuid.UUID           `gorm:"column:category_id;type:uuid"`
	Description  *string              `gorm:"column:description"`
	ThumbnailURL *string              `gorm:"column:thumbnail_url"`
	DesignLink   *string              `gorm:"column:design_link"`
	Category     *productCategoryGORM `gorm:"foreignKey:CategoryID"`
	Variants     []productGORM        `gorm:"foreignKey:ParentID"`
	CreatedAt    time.Time            `gorm:"column:created_at"`
	UpdatedAt    time.Time            `gorm:"column:updated_at"`
}

func (parentProductGORM) TableName() string {
	return "parent_products"
}

type productGORM struct {
	ID                    uuid.UUID                  `gorm:"column:id;primaryKey;type:uuid"`
	ParentID              uuid.UUID                  `gorm:"column:parent_id;type:uuid"`
	VariantName           string                     `gorm:"column:variant_name"`
	SKU                   *string                    `gorm:"column:sku"`
	MaterialType          string                     `gorm:"column:material_type"`
	DefaultWeightGrams    float64                    `gorm:"column:default_weight_grams"`
	DefaultPrintTimeHours float64                    `gorm:"column:default_print_time_hours"`
	BatchSize             int                        `gorm:"column:batch_size"`
	TargetMarginPercent   int                        `gorm:"column:target_margin_percent"`
	BaseHPP               float64                    `gorm:"column:base_hpp"`
	BaseSellingPrice      float64                    `gorm:"column:base_selling_price"`
	Parent                *parentProductGORM         `gorm:"foreignKey:ParentID"`
	Components            []productComponentGORM     `gorm:"foreignKey:ProductID"`
	PackagingItems        []productPackagingItemGORM `gorm:"foreignKey:ProductID"`
	CreatedAt             time.Time                  `gorm:"column:created_at"`
	UpdatedAt             time.Time                  `gorm:"column:updated_at"`
}

func (productGORM) TableName() string {
	return "products"
}

func toParentProductEntity(g *parentProductGORM) *product.ParentProduct {
	if g == nil {
		return nil
	}
	catName := ""
	if g.Category != nil {
		catName = g.Category.Name
	}
	var variants []*product.Product
	if len(g.Variants) > 0 {
		variants = make([]*product.Product, len(g.Variants))
		for i := range g.Variants {
			variants[i] = toProductEntity(&g.Variants[i])
		}
	}
	return product.ReconstructParentProduct(
		g.ID,
		g.UserID,
		g.ParentSKU,
		g.Name,
		g.CategoryID,
		catName,
		g.Description,
		g.ThumbnailURL,
		g.DesignLink,
		variants,
		g.CreatedAt,
		g.UpdatedAt,
	)
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

	var parentEntity *product.ParentProduct
	if g.Parent != nil {
		parentEntity = toParentProductEntity(g.Parent)
	}

	return product.ReconstructProduct(
		g.ID,
		g.ParentID,
		g.VariantName,
		g.SKU,
		g.MaterialType,
		g.DefaultWeightGrams,
		g.DefaultPrintTimeHours,
		g.BatchSize,
		g.TargetMarginPercent,
		g.BaseHPP,
		g.BaseSellingPrice,
		parentEntity,
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

	query := r.db.WithContext(ctx).Model(&productGORM{}).
		Joins("JOIN parent_products ON parent_products.id = products.parent_id").
		Preload("Parent.Category").
		Preload("Components").
		Preload("PackagingItems").
		Where("parent_products.user_id = ?", userID)

	if filter.Search != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where("parent_products.name ILIKE ? OR products.sku ILIKE ? OR parent_products.parent_sku ILIKE ?", s, s, s)
	}
	if filter.CategoryID != nil && *filter.CategoryID != uuid.Nil {
		query = query.Where("parent_products.category_id = ?", *filter.CategoryID)
	} else if filter.Category != "" {
		query = query.Joins("LEFT JOIN product_categories ON product_categories.id = parent_products.category_id").
			Where("product_categories.name = ?", filter.Category)
	}
	if filter.ParentSKU != "" {
		query = query.Where("parent_products.parent_sku = ?", filter.ParentSKU)
	}

	if filter.Limit > 0 {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	query = query.Order("products.created_at DESC")

	if err := query.Find(&gormProducts).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		total = int64(len(gormProducts))
	}

	entities := make([]*product.Product, len(gormProducts))
	for i := range gormProducts {
		entities[i] = toProductEntity(&gormProducts[i])
	}

	return entities, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*product.Product, error) {
	var g productGORM
	err := r.db.WithContext(ctx).
		Preload("Parent.Category").
		Preload("Components").
		Preload("PackagingItems").
		Where("id = ?", id).
		First(&g).Error
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
	err := r.db.WithContext(ctx).
		Preload("Parent.Category").
		Preload("Components").
		Preload("PackagingItems").
		Where("UPPER(TRIM(sku)) = UPPER(TRIM(?))", sku).
		First(&g).Error
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
	err := r.db.WithContext(ctx).
		Preload("Parent.Category").
		Preload("Components").
		Preload("PackagingItems").
		Joins("JOIN parent_products ON parent_products.id = products.parent_id").
		Where("parent_products.user_id = ? AND (products.sku IS NULL OR products.sku = '')", userID).
		Find(&gormProducts).Error
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
			ParentID:              p.ParentID(),
			VariantName:           p.VariantName(),
			SKU:                   p.SKU(),
			MaterialType:          p.MaterialType(),
			DefaultWeightGrams:    p.DefaultWeightGrams(),
			DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
			BatchSize:             p.BatchSize(),
			TargetMarginPercent:   p.TargetMarginPercent(),
			BaseHPP:               p.BaseHPP(),
			BaseSellingPrice:      p.BaseSellingPrice(),
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
			"variant_name":             p.VariantName(),
			"sku":                      p.SKU(),
			"material_type":            p.MaterialType(),
			"default_weight_grams":     p.DefaultWeightGrams(),
			"default_print_time_hours": p.DefaultPrintTimeHours(),
			"batch_size":               p.BatchSize(),
			"target_margin_percent":    p.TargetMarginPercent(),
			"base_hpp":                 p.BaseHPP(),
			"base_selling_price":       p.BaseSellingPrice(),
			"updated_at":               p.UpdatedAt(),
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

// Parent Product Methods
func (r *ProductRepository) FindAllParents(ctx context.Context, userID uuid.UUID, filter product.Filter) ([]*product.ParentProduct, int64, error) {
	var gormParents []parentProductGORM
	var total int64

	query := r.db.WithContext(ctx).Model(&parentProductGORM{}).
		Preload("Category").
		Preload("Variants.Components").
		Preload("Variants.PackagingItems").
		Where("user_id = ?", userID)

	if filter.Search != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where("name ILIKE ? OR parent_sku ILIKE ?", s, s)
	}
	if filter.CategoryID != nil && *filter.CategoryID != uuid.Nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	} else if filter.Category != "" {
		query = query.Joins("LEFT JOIN product_categories ON product_categories.id = parent_products.category_id").
			Where("product_categories.name = ?", filter.Category)
	}
	if filter.ParentSKU != "" {
		query = query.Where("parent_sku = ?", filter.ParentSKU)
	}

	if filter.Limit > 0 {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&gormParents).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		total = int64(len(gormParents))
	}

	entities := make([]*product.ParentProduct, len(gormParents))
	for i := range gormParents {
		entities[i] = toParentProductEntity(&gormParents[i])
	}

	return entities, total, nil
}

func (r *ProductRepository) FindParentByID(ctx context.Context, userID, id uuid.UUID) (*product.ParentProduct, error) {
	var g parentProductGORM
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants.Components").
		Preload("Variants.PackagingItems").
		Where("id = ? AND user_id = ?", id, userID).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toParentProductEntity(&g), nil
}

func (r *ProductRepository) FindParentBySKU(ctx context.Context, userID uuid.UUID, parentSKU string) (*product.ParentProduct, error) {
	var g parentProductGORM
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants.Components").
		Preload("Variants.PackagingItems").
		Where("UPPER(TRIM(parent_sku)) = UPPER(TRIM(?)) AND user_id = ?", parentSKU, userID).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toParentProductEntity(&g), nil
}

func (r *ProductRepository) CreateParent(ctx context.Context, pp *product.ParentProduct) error {
	g := &parentProductGORM{
		ID:           pp.ID(),
		UserID:       pp.UserID(),
		ParentSKU:    pp.ParentSKU(),
		Name:         pp.Name(),
		CategoryID:   pp.CategoryID(),
		Description:  pp.Description(),
		ThumbnailURL: pp.ThumbnailURL(),
		DesignLink:   pp.DesignLink(),
		CreatedAt:    pp.CreatedAt(),
		UpdatedAt:    pp.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *ProductRepository) UpdateParent(ctx context.Context, pp *product.ParentProduct) error {
	updates := map[string]interface{}{
		"parent_sku":    pp.ParentSKU(),
		"name":          pp.Name(),
		"category_id":   pp.CategoryID(),
		"description":   pp.Description(),
		"thumbnail_url": pp.ThumbnailURL(),
		"design_link":   pp.DesignLink(),
		"updated_at":    pp.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Model(&parentProductGORM{}).
		Where("id = ? AND user_id = ?", pp.ID(), pp.UserID()).
		Updates(updates).Error
}

func (r *ProductRepository) DeleteParent(ctx context.Context, userID, id uuid.UUID) error {
	// Check if any variant is used in order_items
	var count int64
	r.db.WithContext(ctx).Table("order_items").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.parent_id = ?", id).
		Count(&count)
	if count > 0 {
		return product.ErrProductInUseInOrder
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find variant IDs
		var variantIDs []uuid.UUID
		tx.Model(&productGORM{}).Where("parent_id = ?", id).Pluck("id", &variantIDs)

		if len(variantIDs) > 0 {
			if err := tx.Where("product_id IN ?", variantIDs).Delete(&productComponentGORM{}).Error; err != nil {
				return err
			}
			if err := tx.Where("product_id IN ?", variantIDs).Delete(&productPackagingItemGORM{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", variantIDs).Delete(&productGORM{}).Error; err != nil {
				return err
			}
		}

		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&parentProductGORM{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return product.ErrNotFound
		}
		return nil
	})
}

func (r *ProductRepository) FindVariantsByParentID(ctx context.Context, parentID uuid.UUID) ([]*product.Product, error) {
	var gormProducts []productGORM
	err := r.db.WithContext(ctx).
		Preload("Parent.Category").
		Preload("Components").
		Preload("PackagingItems").
		Where("parent_id = ?", parentID).
		Order("created_at ASC").
		Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}
	entities := make([]*product.Product, len(gormProducts))
	for i := range gormProducts {
		entities[i] = toProductEntity(&gormProducts[i])
	}
	return entities, nil
}

func (r *ProductRepository) GetTotalSoldMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	type SoldResult struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
		TotalSold int       `gorm:"column:total_sold"`
	}

	result := make(map[uuid.UUID]int)

	query := `
		SELECT oi.product_id, COALESCE(SUM(oi.quantity), 0)::integer as total_sold
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status NOT IN ('CANCELLED', 'UNPAID')
		  AND oi.product_id IS NOT NULL
		GROUP BY oi.product_id
	`
	var soldResults []SoldResult
	if err := r.db.WithContext(ctx).Raw(query).Scan(&soldResults).Error; err == nil {
		for _, s := range soldResults {
			result[s.ProductID] = s.TotalSold
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
