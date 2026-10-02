package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/component"
)

type CreateComponentRequest struct {
	Name                 string   `json:"name"`
	PricePerUnit         float64  `json:"price_per_unit"`
	DefaultMarkupPercent *float64 `json:"default_markup_percent"`
	Description          *string  `json:"description"`
}

type UpdateComponentRequest struct {
	Name                 *string  `json:"name"`
	PricePerUnit         *float64 `json:"price_per_unit"`
	DefaultMarkupPercent *float64 `json:"default_markup_percent"`
	Description          *string  `json:"description"`
}

type ComponentResponse struct {
	ID                     uuid.UUID `json:"id"`
	UserID                 uuid.UUID `json:"user_id"`
	Name                   string    `json:"name"`
	PricePerUnit           float64   `json:"price_per_unit"`
	DefaultMarkupPercent   float64   `json:"default_markup_percent"`
	CalculatedSellingPrice float64   `json:"calculated_selling_price"`
	Description            *string   `json:"description"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func ToComponentResponse(c *component.Component) ComponentResponse {
	return ComponentResponse{
		ID:                     c.ID(),
		UserID:                 c.UserID(),
		Name:                   c.Name(),
		PricePerUnit:           c.PricePerUnit(),
		DefaultMarkupPercent:   c.DefaultMarkupPercent(),
		CalculatedSellingPrice: c.CalculatedSellingPrice(),
		Description:            c.Description(),
		CreatedAt:              c.CreatedAt(),
		UpdatedAt:              c.UpdatedAt(),
	}
}