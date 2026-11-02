package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"symetra-lab-backend-v2/internal/domain/marketplacecred"
	"symetra-lab-backend-v2/pkg/crypto"
)

type marketplaceCredentialGORM struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	Platform     string    `gorm:"column:platform;uniqueIndex;type:varchar(50)"`
	PartnerID    string    `gorm:"column:partner_id;type:varchar(100)"`
	PartnerKey   string    `gorm:"column:partner_key;type:text"` // Ciphertext AES-256-GCM
	PushKey      string    `gorm:"column:push_key;type:text"`    // Ciphertext AES-256-GCM
	RedirectURL  string    `gorm:"column:redirect_url;type:text"`
	IsProduction bool      `gorm:"column:is_production"`
	IsActive     bool      `gorm:"column:is_active"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (marketplaceCredentialGORM) TableName() string {
	return "marketplace_credentials"
}

type MarketplaceCredentialRepository struct {
	db            *gorm.DB
	encryptionKey []byte
}

func NewMarketplaceCredentialRepository(db *gorm.DB, encryptionPassphrase string) *MarketplaceCredentialRepository {
	// Pastikan skema tabel terbentuk
	_ = db.AutoMigrate(&marketplaceCredentialGORM{})

	var key []byte
	if encryptionPassphrase != "" {
		key = crypto.DeriveKey32(encryptionPassphrase)
	} else {
		// Fallback default dev key jika env passphrase belum di-set
		key = crypto.DeriveKey32("symetra-lab-default-master-key-32")
	}

	return &MarketplaceCredentialRepository{
		db:            db,
		encryptionKey: key,
	}
}

func (r *MarketplaceCredentialRepository) FindByPlatform(ctx context.Context, platform marketplacecred.Platform) (*marketplacecred.MarketplaceCredential, error) {
	var m marketplaceCredentialGORM
	err := r.db.WithContext(ctx).
		Where("platform = ? AND is_active = ?", string(platform), true).
		First(&m).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, marketplacecred.ErrCredentialNotFound
		}
		return nil, err
	}

	return r.toDomain(&m)
}

func (r *MarketplaceCredentialRepository) FindAll(ctx context.Context) ([]*marketplacecred.MarketplaceCredential, error) {
	var models []marketplaceCredentialGORM
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	results := make([]*marketplacecred.MarketplaceCredential, 0, len(models))
	for _, m := range models {
		domainEntity, err := r.toDomain(&m)
		if err != nil {
			// Jika gagal dekripsi (misal data korup), lewati atau log
			continue
		}
		results = append(results, domainEntity)
	}

	return results, nil
}

func (r *MarketplaceCredentialRepository) Save(ctx context.Context, cred *marketplacecred.MarketplaceCredential) error {
	if cred == nil {
		return errors.New("credential cannot be nil")
	}

	// Enkripsi PartnerKey sebelum disimpan ke basis data
	encryptedKey, err := crypto.Encrypt(cred.PartnerKey(), r.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt partner key: %w", err)
	}

	encryptedPushKey := ""
	if cred.PushKey() != "" {
		enc, err := crypto.Encrypt(cred.PushKey(), r.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt push key: %w", err)
		}
		encryptedPushKey = enc
	}

	m := marketplaceCredentialGORM{
		ID:           cred.ID(),
		Platform:     string(cred.Platform()),
		PartnerID:    cred.PartnerID(),
		PartnerKey:   encryptedKey,
		PushKey:      encryptedPushKey,
		RedirectURL:  cred.RedirectURL(),
		IsProduction: cred.IsProduction(),
		IsActive:     cred.IsActive(),
		CreatedAt:    cred.CreatedAt(),
		UpdatedAt:    cred.UpdatedAt(),
	}

	// Upsert atomik native PostgreSQL dengan ON CONFLICT (platform)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "platform"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"partner_id",
				"partner_key",
				"push_key",
				"redirect_url",
				"is_production",
				"is_active",
				"updated_at",
			}),
		}).
		Create(&m).Error
}

func (r *MarketplaceCredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&marketplaceCredentialGORM{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return marketplacecred.ErrCredentialNotFound
	}
	return nil
}

func (r *MarketplaceCredentialRepository) toDomain(m *marketplaceCredentialGORM) (*marketplacecred.MarketplaceCredential, error) {
	// Dekripsi PartnerKey
	decryptedKey, err := crypto.Decrypt(m.PartnerKey, r.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt partner key for platform %s: %w", m.Platform, err)
	}

	// Dekripsi PushKey jika ada
	decryptedPushKey := ""
	if m.PushKey != "" {
		if dec, err := crypto.Decrypt(m.PushKey, r.encryptionKey); err == nil {
			decryptedPushKey = dec
		}
	}

	return marketplacecred.ReconstructMarketplaceCredential(
		m.ID,
		marketplacecred.Platform(m.Platform),
		m.PartnerID,
		decryptedKey,
		decryptedPushKey,
		m.RedirectURL,
		m.IsProduction,
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

