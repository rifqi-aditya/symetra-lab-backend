package category

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	id        uuid.UUID
	userID    uuid.UUID
	name      string
	createdAt time.Time
}

func NewCategory(userID uuid.UUID, name string) (*Category, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrCategoryNameEmpty
	}
	return &Category{
		id:        uuid.New(),
		userID:    userID,
		name:      cleanName,
		createdAt: time.Now(),
	}, nil
}

func ReconstructCategory(id, userID uuid.UUID, name string, createdAt time.Time) *Category {
	return &Category{
		id:        id,
		userID:    userID,
		name:      name,
		createdAt: createdAt,
	}
}

func (c *Category) ID() uuid.UUID        { return c.id }
func (c *Category) UserID() uuid.UUID    { return c.userID }
func (c *Category) Name() string         { return c.name }
func (c *Category) CreatedAt() time.Time { return c.createdAt }

func (c *Category) UpdateName(name string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return ErrCategoryNameEmpty
	}
	c.name = cleanName
	return nil
}