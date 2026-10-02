package component

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Component represents non-3D printed physical accessories/hardware.
type Component struct {
	id                   uuid.UUID
	userID               uuid.UUID
	name                 string
	pricePerUnit         float64
	defaultMarkupPercent float64
	description          *string
	createdAt            time.Time
	updatedAt            time.Time
}

func NewComponent(userID uuid.UUID, name string, pricePerUnit, markupPercent float64, description *string) (*Component, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrComponentNameRequired
	}
	if pricePerUnit < 0 {
		return nil, ErrInvalidPrice
	}
	now := time.Now()
	return &Component{
		id:                   uuid.New(),
		userID:               userID,
		name:                 cleanName,
		pricePerUnit:         pricePerUnit,
		defaultMarkupPercent: markupPercent,
		description:          description,
		createdAt:            now,
		updatedAt:            now,
	}, nil
}

func ReconstructComponent(
	id, userID uuid.UUID,
	name string,
	pricePerUnit, defaultMarkupPercent float64,
	description *string,
	createdAt, updatedAt time.Time,
) *Component {
	return &Component{
		id:                   id,
		userID:               userID,
		name:                 name,
		pricePerUnit:         pricePerUnit,
		defaultMarkupPercent: defaultMarkupPercent,
		description:          description,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}
}

func (c *Component) ID() uuid.UUID                  { return c.id }
func (c *Component) UserID() uuid.UUID              { return c.userID }
func (c *Component) Name() string                   { return c.name }
func (c *Component) PricePerUnit() float64          { return c.pricePerUnit }
func (c *Component) DefaultMarkupPercent() float64  { return c.defaultMarkupPercent }
func (c *Component) Description() *string           { return c.description }
func (c *Component) CreatedAt() time.Time           { return c.createdAt }
func (c *Component) UpdatedAt() time.Time           { return c.updatedAt }

func (c *Component) Update(name string, pricePerUnit, markupPercent float64, description *string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return ErrComponentNameRequired
	}
	if pricePerUnit < 0 {
		return ErrInvalidPrice
	}
	c.name = cleanName
	c.pricePerUnit = pricePerUnit
	c.defaultMarkupPercent = markupPercent
	c.description = description
	c.updatedAt = time.Now()
	return nil
}

// CalculatedSellingPrice returns estimated selling price per unit after markup.
func (c *Component) CalculatedSellingPrice() float64 {
	markup := c.defaultMarkupPercent
	if markup < 0 {
		markup = 0
	}
	raw := c.pricePerUnit * (1.0 + (markup / 100.0))
	return math.Round(raw*100) / 100
}