package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/category"
)

type productCategoryGORM struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid"`
	Name      string    `gorm:"column:name;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (productCategoryGORM) TableName() string {
	return "product_categories"
}

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*category.Category, error) {
	var list []productCategoryGORM
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]*category.Category, len(list))
	for i, c := range list {
		res[i] = category.ReconstructCategory(c.ID, c.UserID, c.Name, c.CreatedAt)
	}
	return res, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*category.Category, error) {
	var c productCategoryGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return category.ReconstructCategory(c.ID, c.UserID, c.Name, c.CreatedAt), nil
}

func (r *CategoryRepository) Create(ctx context.Context, c *category.Category) error {
	m := productCategoryGORM{
		ID:        c.ID(),
		UserID:    c.UserID(),
		Name:      c.Name(),
		CreatedAt: c.CreatedAt(),
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *CategoryRepository) Update(ctx context.Context, c *category.Category) error {
	return r.db.WithContext(ctx).Model(&productCategoryGORM{}).
		Where("id = ?", c.ID()).
		Update("name", c.Name()).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&productCategoryGORM{}).Error
}