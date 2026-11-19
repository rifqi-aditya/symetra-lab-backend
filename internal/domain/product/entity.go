package product

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProductComponent represents a hardware accessory required for one unit of product.
type ProductComponent struct {
	id          uuid.UUID
	productID   uuid.UUID
	componentID *uuid.UUID
	quantity    float64
	createdAt   time.Time
}

func ReconstructProductComponent(
	id, productID uuid.UUID,
	componentID *uuid.UUID,
	quantity float64,
	createdAt time.Time,
) *ProductComponent {
	return &ProductComponent{
		id:          id,
		productID:   productID,
		componentID: componentID,
		quantity:    quantity,
		createdAt:   createdAt,
	}
}

func (pc *ProductComponent) ID() uuid.UUID            { return pc.id }
func (pc *ProductComponent) ProductID() uuid.UUID     { return pc.productID }
func (pc *ProductComponent) ComponentID() *uuid.UUID  { return pc.componentID }
func (pc *ProductComponent) Quantity() float64        { return pc.quantity }
func (pc *ProductComponent) MarkupPercent() float64   { return 0 }
func (pc *ProductComponent) CreatedAt() time.Time     { return pc.createdAt }

// ProductPackagingItem represents custom packaging material required for one unit of product.
type ProductPackagingItem struct {
	id              uuid.UUID
	productID       uuid.UUID
	packagingItemID uuid.UUID
	quantityUsed    float64
	createdAt       time.Time
}

func ReconstructProductPackagingItem(
	id, productID, packagingItemID uuid.UUID,
	quantityUsed float64,
	createdAt time.Time,
) *ProductPackagingItem {
	return &ProductPackagingItem{
		id:              id,
		productID:       productID,
		packagingItemID: packagingItemID,
		quantityUsed:    quantityUsed,
		createdAt:       createdAt,
	}
}

func (pi *ProductPackagingItem) ID() uuid.UUID              { return pi.id }
func (pi *ProductPackagingItem) ProductID() uuid.UUID       { return pi.productID }
func (pi *ProductPackagingItem) PackagingItemID() uuid.UUID { return pi.packagingItemID }
func (pi *ProductPackagingItem) QuantityUsed() float64      { return pi.quantityUsed }
func (pi *ProductPackagingItem) CreatedAt() time.Time       { return pi.createdAt }

// ParentProduct represents the master catalog / etalase listing (3D design master).
type ParentProduct struct {
	id           uuid.UUID
	userID       uuid.UUID
	parentSKU    string
	name         string
	categoryID   *uuid.UUID
	categoryName string // Populated when joined with product_categories
	description  *string
	thumbnailURL *string
	designLink   *string
	variants     []*Product
	createdAt    time.Time
	updatedAt    time.Time
}

func NewParentProduct(
	userID uuid.UUID,
	name string,
	parentSKU string,
	categoryID *uuid.UUID,
	description, thumbnailURL, designLink *string,
) (*ParentProduct, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrProductNameRequired
	}
	cleanSKU := strings.TrimSpace(strings.ToUpper(parentSKU))
	if cleanSKU == "" {
		cleanSKU = "PAR-" + strings.ToUpper(uuid.New().String()[:8])
	}
	now := time.Now()
	return &ParentProduct{
		id:           uuid.New(),
		userID:       userID,
		parentSKU:    cleanSKU,
		name:         cleanName,
		categoryID:   categoryID,
		description:  description,
		thumbnailURL: thumbnailURL,
		designLink:   designLink,
		variants:     make([]*Product, 0),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructParentProduct(
	id, userID uuid.UUID,
	parentSKU, name string,
	categoryID *uuid.UUID,
	categoryName string,
	description, thumbnailURL, designLink *string,
	variants []*Product,
	createdAt, updatedAt time.Time,
) *ParentProduct {
	return &ParentProduct{
		id:           id,
		userID:       userID,
		parentSKU:    parentSKU,
		name:         name,
		categoryID:   categoryID,
		categoryName: categoryName,
		description:  description,
		thumbnailURL: thumbnailURL,
		designLink:   designLink,
		variants:     variants,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// ParentProduct Getters
func (pp *ParentProduct) ID() uuid.UUID             { return pp.id }
func (pp *ParentProduct) UserID() uuid.UUID         { return pp.userID }
func (pp *ParentProduct) ParentSKU() string         { return pp.parentSKU }
func (pp *ParentProduct) Name() string              { return pp.name }
func (pp *ParentProduct) CategoryID() *uuid.UUID    { return pp.categoryID }
func (pp *ParentProduct) CategoryName() string      { return pp.categoryName }
func (pp *ParentProduct) Description() *string      { return pp.description }
func (pp *ParentProduct) ThumbnailURL() *string     { return pp.thumbnailURL }
func (pp *ParentProduct) DesignLink() *string       { return pp.designLink }
func (pp *ParentProduct) Variants() []*Product      { return pp.variants }
func (pp *ParentProduct) CreatedAt() time.Time      { return pp.createdAt }
func (pp *ParentProduct) UpdatedAt() time.Time      { return pp.updatedAt }

func (pp *ParentProduct) SetVariants(v []*Product) {
	pp.variants = v
}

func (pp *ParentProduct) Update(
	name string,
	parentSKU string,
	categoryID *uuid.UUID,
	description, thumbnailURL, designLink *string,
) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return ErrProductNameRequired
	}
	pp.name = cleanName
	if strings.TrimSpace(parentSKU) != "" {
		pp.parentSKU = strings.TrimSpace(strings.ToUpper(parentSKU))
	}
	pp.categoryID = categoryID
	pp.description = description
	pp.thumbnailURL = thumbnailURL
	pp.designLink = designLink
	pp.updatedAt = time.Now()
	return nil
}

// Product represents a physical sellable variant line of a 3D print.
type Product struct {
	id                    uuid.UUID
	parentID              uuid.UUID
	variantName           string
	sku                   *string
	materialType          string
	defaultWeightGrams    float64
	defaultPrintTimeHours float64
	batchSize             int
	targetMarginPercent   int
	baseHPP               float64
	baseSellingPrice      float64
	components            []*ProductComponent
	packagingItems        []*ProductPackagingItem
	parent                *ParentProduct
	createdAt             time.Time
	updatedAt             time.Time
}

// NewProduct creates a new Product variant entity.
func NewProduct(
	parentID uuid.UUID,
	variantName string,
	sku *string,
	materialType string,
	defaultWeightGrams float64,
	defaultPrintTimeHours float64,
) (*Product, error) {
	if parentID == uuid.Nil {
		return nil, ErrParentIDRequired
	}
	cleanVar := strings.TrimSpace(variantName)
	if cleanVar == "" {
		cleanVar = "Standard"
	}
	mat := strings.TrimSpace(materialType)
	if mat == "" {
		mat = "PLA"
	}
	now := time.Now()
	return &Product{
		id:                    uuid.New(),
		parentID:              parentID,
		variantName:           cleanVar,
		sku:                   sku,
		materialType:          mat,
		defaultWeightGrams:    defaultWeightGrams,
		defaultPrintTimeHours: defaultPrintTimeHours,
		batchSize:             1,
		targetMarginPercent:   30,
		createdAt:             now,
		updatedAt:             now,
	}, nil
}

// ReconstructProduct reconstitutes a Product variant from persistence.
func ReconstructProduct(
	id, parentID uuid.UUID,
	variantName string,
	sku *string,
	materialType string,
	defaultWeightGrams, defaultPrintTimeHours float64,
	batchSize, targetMarginPercent int,
	baseHPP, baseSellingPrice float64,
	parent *ParentProduct,
	components []*ProductComponent,
	packagingItems []*ProductPackagingItem,
	createdAt, updatedAt time.Time,
) *Product {
	return &Product{
		id:                    id,
		parentID:              parentID,
		variantName:           variantName,
		sku:                   sku,
		materialType:          materialType,
		defaultWeightGrams:    defaultWeightGrams,
		defaultPrintTimeHours: defaultPrintTimeHours,
		batchSize:             batchSize,
		targetMarginPercent:   targetMarginPercent,
		baseHPP:               baseHPP,
		baseSellingPrice:      baseSellingPrice,
		parent:                parent,
		components:            components,
		packagingItems:        packagingItems,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

// Getters
func (p *Product) ID() uuid.UUID                           { return p.id }
func (p *Product) ParentID() uuid.UUID                     { return p.parentID }
func (p *Product) VariantName() string                     { return p.variantName }
func (p *Product) SKU() *string                            { return p.sku }
func (p *Product) MaterialType() string                    { return p.materialType }
func (p *Product) DefaultWeightGrams() float64             { return p.defaultWeightGrams }
func (p *Product) DefaultPrintTimeHours() float64          { return p.defaultPrintTimeHours }
func (p *Product) BatchSize() int                          { return p.batchSize }
func (p *Product) TargetMarginPercent() int                { return p.targetMarginPercent }
func (p *Product) BaseHPP() float64                        { return p.baseHPP }
func (p *Product) BaseSellingPrice() float64               { return p.baseSellingPrice }
func (p *Product) Parent() *ParentProduct                  { return p.parent }
func (p *Product) Components() []*ProductComponent         { return p.components }
func (p *Product) PackagingItems() []*ProductPackagingItem { return p.packagingItems }
func (p *Product) CreatedAt() time.Time                    { return p.createdAt }
func (p *Product) UpdatedAt() time.Time                    { return p.updatedAt }

// Backward compatibility helpers for ParentProduct attributes
func (p *Product) UserID() uuid.UUID {
	if p.parent != nil {
		return p.parent.UserID()
	}
	return uuid.Nil
}

func (p *Product) Name() string {
	if p.parent != nil {
		if p.variantName != "" && p.variantName != "Standard" {
			return p.parent.Name() + " (" + p.variantName + ")"
		}
		return p.parent.Name()
	}
	return p.variantName
}

func (p *Product) ParentSKU() *string {
	if p.parent != nil {
		sku := p.parent.ParentSKU()
		return &sku
	}
	return nil
}

func (p *Product) Description() *string {
	if p.parent != nil {
		return p.parent.Description()
	}
	return nil
}

func (p *Product) Category() string {
	if p.parent != nil {
		return p.parent.CategoryName()
	}
	return ""
}

func (p *Product) CategoryID() *uuid.UUID {
	if p.parent != nil {
		return p.parent.CategoryID()
	}
	return nil
}

func (p *Product) ThumbnailURL() *string {
	if p.parent != nil {
		return p.parent.ThumbnailURL()
	}
	return nil
}

func (p *Product) DesignLink() *string {
	if p.parent != nil {
		return p.parent.DesignLink()
	}
	return nil
}

// PackagingPresetID & PackingFeeIDR & TimesOrdered compatibility stubs
func (p *Product) PackagingPresetID() *uuid.UUID { return nil }
func (p *Product) PackingFeeIDR() int            { return 0 }
func (p *Product) TimesOrdered() int             { return 0 }

// Setters and mutators
func (p *Product) SetParent(parent *ParentProduct) {
	p.parent = parent
	if parent != nil {
		p.parentID = parent.ID()
	}
}

func (p *Product) SetSKU(sku *string) {
	p.sku = sku
	p.updatedAt = time.Now()
}

func (p *Product) SetPricing(hpp, sellingPrice float64) {
	p.baseHPP = hpp
	p.baseSellingPrice = sellingPrice
	p.updatedAt = time.Now()
}

func (p *Product) SetBOM(components []*ProductComponent, packagingItems []*ProductPackagingItem) {
	p.components = components
	p.packagingItems = packagingItems
	p.updatedAt = time.Now()
}

func (p *Product) UpdateSpecs(
	variantName string,
	weightGrams, printTimeHours float64,
	materialType string,
	batchSize, marginPercent int,
) error {
	cleanVar := strings.TrimSpace(variantName)
	if cleanVar != "" {
		p.variantName = cleanVar
	}
	p.defaultWeightGrams = weightGrams
	p.defaultPrintTimeHours = printTimeHours
	if strings.TrimSpace(materialType) != "" {
		p.materialType = strings.TrimSpace(materialType)
	}
	if batchSize > 0 {
		p.batchSize = batchSize
	}
	if marginPercent > 0 {
		p.targetMarginPercent = marginPercent
	}
	p.updatedAt = time.Now()
	return nil
}

func (p *Product) Validate() error {
	if p.parentID == uuid.Nil {
		return ErrParentIDRequired
	}
	if p.defaultWeightGrams < 0 {
		return ErrInvalidWeight
	}
	if p.defaultPrintTimeHours < 0 {
		return ErrInvalidPrintTime
	}
	return nil
}
