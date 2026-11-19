package packaging

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PackagingItem represents raw materials used for product and shipping packaging.
type PackagingItem struct {
	id            uuid.UUID
	userID        uuid.UUID
	name          string
	unitType      string
	unitCost      float64
	stockQuantity float64
	createdAt     time.Time
	updatedAt     time.Time
}

func NewPackagingItem(
	userID uuid.UUID,
	name, unitType string,
	unitCost, stockQuantity float64,
) (*PackagingItem, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrInvalidQuantity
	}
	uType := strings.TrimSpace(unitType)
	if uType == "" {
		uType = "PCS"
	}

	now := time.Now()
	return &PackagingItem{
		id:            uuid.New(),
		userID:        userID,
		name:          cleanName,
		unitType:      uType,
		unitCost:      unitCost,
		stockQuantity: stockQuantity,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func ReconstructPackagingItem(
	id, userID uuid.UUID,
	name, unitType string,
	unitCost, stockQuantity float64,
	createdAt, updatedAt time.Time,
) *PackagingItem {
	return &PackagingItem{
		id:            id,
		userID:        userID,
		name:          name,
		unitType:      unitType,
		unitCost:      unitCost,
		stockQuantity: stockQuantity,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (p *PackagingItem) ID() uuid.UUID             { return p.id }
func (p *PackagingItem) UserID() uuid.UUID         { return p.userID }
func (p *PackagingItem) Name() string              { return p.name }
func (p *PackagingItem) UnitType() string          { return p.unitType }
func (p *PackagingItem) UnitCost() float64         { return p.unitCost }
func (p *PackagingItem) StockQuantity() float64    { return p.stockQuantity }
func (p *PackagingItem) CreatedAt() time.Time      { return p.createdAt }
func (p *PackagingItem) UpdatedAt() time.Time      { return p.updatedAt }

// Backward compatibility stubs
func (p *PackagingItem) Category() string          { return "OTHER" }
func (p *PackagingItem) PurchasePrice() float64    { return p.unitCost }
func (p *PackagingItem) PurchaseQuantity() float64 { return 1.0 }
func (p *PackagingItem) CalculateUnitCost() float64 { return p.unitCost }

func (p *PackagingItem) Update(name, unitType string, unitCost, stockQuantity float64) error {
	cleanName := strings.TrimSpace(name)
	if cleanName != "" {
		p.name = cleanName
	}
	if unitType != "" {
		p.unitType = strings.TrimSpace(unitType)
	}
	if unitCost >= 0 {
		p.unitCost = unitCost
	}
	if stockQuantity >= 0 {
		p.stockQuantity = stockQuantity
	}
	p.updatedAt = time.Now()
	return nil
}

// PackagingPresetItem represents an individual item requirement inside a preset bundle.
type PackagingPresetItem struct {
	id              uuid.UUID
	presetID        uuid.UUID
	packagingItemID uuid.UUID
	quantityUsed    float64
	item            *PackagingItem
	createdAt       time.Time
}

func ReconstructPresetItem(
	id, presetID, packagingItemID uuid.UUID,
	quantityUsed float64,
	item *PackagingItem,
	createdAt time.Time,
) *PackagingPresetItem {
	return &PackagingPresetItem{
		id:              id,
		presetID:        presetID,
		packagingItemID: packagingItemID,
		quantityUsed:    quantityUsed,
		item:            item,
		createdAt:       createdAt,
	}
}

func (pi *PackagingPresetItem) ID() uuid.UUID              { return pi.id }
func (pi *PackagingPresetItem) PresetID() uuid.UUID        { return pi.presetID }
func (pi *PackagingPresetItem) PackagingItemID() uuid.UUID { return pi.packagingItemID }
func (pi *PackagingPresetItem) QuantityUsed() float64      { return pi.quantityUsed }
func (pi *PackagingPresetItem) Item() *PackagingItem       { return pi.item }
func (pi *PackagingPresetItem) CreatedAt() time.Time       { return pi.createdAt }

// PackagingPreset represents bundled packaging sets (e.g. "Keychain packing", "Mini diorama").
type PackagingPreset struct {
	id          uuid.UUID
	userID      uuid.UUID
	name        string
	description *string
	items       []*PackagingPresetItem
	createdAt   time.Time
	updatedAt   time.Time
}

func NewPackagingPreset(
	userID uuid.UUID,
	name string,
	description *string,
	items []*PackagingPresetItem,
) (*PackagingPreset, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrPresetNotFound
	}
	now := time.Now()
	presetID := uuid.New()
	for _, it := range items {
		it.presetID = presetID
		if it.id == uuid.Nil {
			it.id = uuid.New()
		}
		it.createdAt = now
	}
	return &PackagingPreset{
		id:          presetID,
		userID:      userID,
		name:        cleanName,
		description: description,
		items:       items,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructPackagingPreset(
	id, userID uuid.UUID,
	name string,
	description *string,
	items []*PackagingPresetItem,
	createdAt, updatedAt time.Time,
) *PackagingPreset {
	return &PackagingPreset{
		id:          id,
		userID:      userID,
		name:        name,
		description: description,
		items:       items,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *PackagingPreset) ID() uuid.UUID                    { return p.id }
func (p *PackagingPreset) UserID() uuid.UUID                { return p.userID }
func (p *PackagingPreset) Name() string                     { return p.name }
func (p *PackagingPreset) Description() *string             { return p.description }
func (p *PackagingPreset) Items() []*PackagingPresetItem    { return p.items }
func (p *PackagingPreset) CreatedAt() time.Time             { return p.createdAt }
func (p *PackagingPreset) UpdatedAt() time.Time             { return p.updatedAt }

func (p *PackagingPreset) Update(name string, description *string, items []*PackagingPresetItem) {
	if strings.TrimSpace(name) != "" {
		p.name = strings.TrimSpace(name)
	}
	p.description = description
	p.items = items
	p.updatedAt = time.Now()
}

func (p *PackagingPreset) TotalCost() float64 {
	var total float64
	for _, it := range p.items {
		if it.item != nil {
			total += math.Round(it.quantityUsed*it.item.unitCost*100) / 100
		}
	}
	return math.Round(total*100) / 100
}