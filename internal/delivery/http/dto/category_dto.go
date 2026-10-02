package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/category"
)

type CategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func ToCategoryResponse(c *category.Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID(),
		UserID:    c.UserID(),
		Name:      c.Name(),
		CreatedAt: c.CreatedAt(),
	}
}