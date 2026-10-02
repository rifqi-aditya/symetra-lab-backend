package product

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProductComponent represents a hardware accessory required for one unit of product.
type ProductComponent struct {
	id            uuid.UUID
	productID     uuid.UUID
	componentID   *uuid.UUID
	quantity      float64
	markupPercent float64
	createdAt     time.Time
}

func ReconstructProductComponent(
	id, productID uuid.UUID,
	componentID *uuid.UUID,
	quantity, markupPercent float64,
	createdAt time.Time,
) *ProductComponent {
	return &ProductComponent{
		id:            id,
		productID:     productID,
		componentID:   componentID,
		quantity:      quantity,
		markupPercent: markupPercent,
		createdAt:     createdAt,
	}
}

func (pc *ProductComponent) ID() uuid.UUID            { return pc.id }
func (pc *ProductComponent) ProductID() uuid.UUID     { return pc.productID }
func (pc *ProductComponent) ComponentID() *uuid.UUID  { return pc.componentID }
func (pc *ProductComponent) Quantity() float64        { return pc.quantity }
func (pc *ProductComponent) MarkupPercent() float64   { return pc.markupPercent }
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

// Product represents the workshop master product and physical 3D print blueprint.
type Product struct {
	id                     uuid.UUID
	userID                 uuid.UUID
	name                   string
	parentSKU              *string
	sku                    *string
	description            *string
	category               string
	thumbnailURL           *string
	designLink             *string
	defaultWeightGrams     float64
	materialType           string
	defaultPrintTimeHours  float64
	defaultMachineID       *uuid.UUID
	packagingPresetID      *uuid.UUID
	batchSize              int
	packingFeeIDR          int
	baseHPP                float64
	baseSellingPrice       float64
	targetMarginPercent    int
	timesOrdered           int
	components             []*ProductComponent
	packagingItems         []*ProductPackagingItem
	createdAt              time.Time
	updatedAt              time.Time
}

// NewProduct creates a new Product entity with basic domain validations.
func NewProduct(
	userID uuid.UUID,
	name string,
	sku *string,
	category string,
	defaultWeightGrams float64,
	defaultPrintTimeHours float64,
	materialType string,
) (*Product, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrProductNameRequired
	}

	mat := strings.TrimSpace(materialType)
	if mat == "" {
		mat = "PLA"
	}

	now := time.Now()
	return &Product{
		id:                    uuid.New(),
		userID:                userID,
		name:                  cleanName,
		sku:                   sku,
		category:              strings.TrimSpace(category),
		defaultWeightGrams:    defaultWeightGrams,
		defaultPrintTimeHours: defaultPrintTimeHours,
		materialType:          mat,
		batchSize:             1,
		targetMarginPercent:   30,
		createdAt:             now,
		updatedAt:             now,
	}, nil
}

// ReconstructProduct reconstitutes a Product from database persistence.
func ReconstructProduct(
	id, userID uuid.UUID,
	name string,
	parentSKU, sku, description *string,
	category string,
	thumbnailURL, designLink *string,
	defaultWeightGrams float64,
	materialType string,
	defaultPrintTimeHours float64,
	defaultMachineID, packagingPresetID *uuid.UUID,
	batchSize, packingFeeIDR int,
	baseHPP, baseSellingPrice float64,
	targetMarginPercent, timesOrdered int,
	components []*ProductComponent,
	packagingItems []*ProductPackagingItem,
	createdAt, updatedAt time.Time,
) *Product {
	return &Product{
		id:                    id,
		userID:                userID,
		name:                  name,
		parentSKU:             parentSKU,
		sku:                   sku,
		description:           description,
		category:              category,
		thumbnailURL:          thumbnailURL,
		designLink:            designLink,
		defaultWeightGrams:    defaultWeightGrams,
		materialType:          materialType,
		defaultPrintTimeHours: defaultPrintTimeHours,
		defaultMachineID:      defaultMachineID,
		packagingPresetID:     packagingPresetID,
		batchSize:             batchSize,
		packingFeeIDR:         packingFeeIDR,
		baseHPP:               baseHPP,
		baseSellingPrice:      baseSellingPrice,
		targetMarginPercent:   targetMarginPercent,
		timesOrdered:          timesOrdered,
		components:            components,
		packagingItems:        packagingItems,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

// Getters
func (p *Product) ID() uuid.UUID                             { return p.id }
func (p *Product) UserID() uuid.UUID                         { return p.userID }
func (p *Product) Name() string                              { return p.name }
func (p *Product) ParentSKU() *string                        { return p.parentSKU }
func (p *Product) SKU() *string                              { return p.sku }
func (p *Product) Description() *string                      { return p.description }
func (p *Product) Category() string                          { return p.category }
func (p *Product) ThumbnailURL() *string                     { return p.thumbnailURL }
func (p *Product) DesignLink() *string                       { return p.designLink }
func (p *Product) DefaultWeightGrams() float64               { return p.defaultWeightGrams }
func (p *Product) MaterialType() string                      { return p.materialType }
func (p *Product) DefaultPrintTimeHours() float64            { return p.defaultPrintTimeHours }
func (p *Product) DefaultMachineID() *uuid.UUID              { return p.defaultMachineID }
func (p *Product) PackagingPresetID() *uuid.UUID             { return p.packagingPresetID }
func (p *Product) BatchSize() int                            { return p.batchSize }
func (p *Product) PackingFeeIDR() int                        { return p.packingFeeIDR }
func (p *Product) BaseHPP() float64                          { return p.baseHPP }
func (p *Product) BaseSellingPrice() float64                 { return p.baseSellingPrice }
func (p *Product) TargetMarginPercent() int                  { return p.targetMarginPercent }
func (p *Product) TimesOrdered() int                         { return p.timesOrdered }
func (p *Product) Components() []*ProductComponent           { return p.components }
func (p *Product) PackagingItems() []*ProductPackagingItem   { return p.packagingItems }
func (p *Product) CreatedAt() time.Time                      { return p.createdAt }
func (p *Product) UpdatedAt() time.Time                      { return p.updatedAt }

// Business methods
func (p *Product) Validate() error {
	if p.name == "" {
		return ErrProductNameRequired
	}
	if p.defaultWeightGrams < 0 {
		return ErrInvalidWeight
	}
	if p.defaultPrintTimeHours < 0 {
		return ErrInvalidPrintTime
	}
	return nil
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

func (p *Product) UpdateSpecs(
	name string,
	description, thumbnailURL, designLink *string,
	category string,
	weightGrams, printTimeHours float64,
	materialType string,
	machineID, presetID *uuid.UUID,
	batchSize, packingFeeIDR, marginPercent int,
) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return ErrProductNameRequired
	}
	p.name = cleanName
	p.description = description
	p.thumbnailURL = thumbnailURL
	p.designLink = designLink
	p.category = strings.TrimSpace(category)
	p.defaultWeightGrams = weightGrams
	p.defaultPrintTimeHours = printTimeHours
	if strings.TrimSpace(materialType) != "" {
		p.materialType = strings.TrimSpace(materialType)
	}
	p.defaultMachineID = machineID
	p.packagingPresetID = presetID
	if batchSize > 0 {
		p.batchSize = batchSize
	}
	p.packingFeeIDR = packingFeeIDR
	if marginPercent > 0 {
		p.targetMarginPercent = marginPercent
	}
	p.updatedAt = time.Now()
	return nil
}

func (p *Product) SetBOM(components []*ProductComponent, packagingItems []*ProductPackagingItem) {
	p.components = components
	p.packagingItems = packagingItems
	p.updatedAt = time.Now()
}