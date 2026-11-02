package marketplacecred

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/config"
	domainCred "symetra-lab-backend-v2/internal/domain/marketplacecred"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type UseCase struct {
	repo domainCred.Repository
	cfg  *config.Config
}

func NewUseCase(repo domainCred.Repository, cfg *config.Config) *UseCase {
	return &UseCase{
		repo: repo,
		cfg:  cfg,
	}
}

// ListCredentials mengembalikan semua kredensial yang tersimpan di DB
func (uc *UseCase) ListCredentials(ctx context.Context) ([]*domainCred.MarketplaceCredential, error) {
	return uc.repo.FindAll(ctx)
}

// SaveCredential menyimpan atau memperbarui kredensial untuk platform tertentu
func (uc *UseCase) SaveCredential(
	ctx context.Context,
	platformStr, partnerID, partnerKey, redirectURL string,
	isProduction bool,
) (*domainCred.MarketplaceCredential, error) {
	platform := domainCred.Platform(strings.ToUpper(strings.TrimSpace(platformStr)))
	cred, err := domainCred.NewMarketplaceCredential(platform, partnerID, partnerKey, redirectURL, isProduction)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, cred); err != nil {
		return nil, err
	}

	return uc.repo.FindByPlatform(ctx, platform)
}

// DeleteCredential menghapus kredensial berdasarkan ID
func (uc *UseCase) DeleteCredential(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

// GetShopeeClient membuat atau mengembalikan instance Shopee Client yang dikonfigurasi secara dinamis
// Prioritas:
// 1. Data kredensial aktif dari Database (`marketplace_credentials`)
// 2. Fallback kredensial dari `.env` (`config.Config`)
func (uc *UseCase) GetShopeeClient(ctx context.Context) (*pkgShopee.Client, error) {
	cred, err := uc.repo.FindByPlatform(ctx, domainCred.PlatformShopee)
	if err == nil && cred != nil && cred.PartnerID() != "" && cred.PartnerKey() != "" {
		partnerIDInt, parseErr := strconv.ParseInt(cred.PartnerID(), 10, 64)
		if parseErr == nil && partnerIDInt > 0 {
			redirectURL := cred.RedirectURL()
			if redirectURL == "" && uc.cfg != nil {
				redirectURL = uc.cfg.ShopeeRedirectURL
			}
			return pkgShopee.NewClient(
				partnerIDInt,
				cred.PartnerKey(),
				cred.IsProduction(),
				redirectURL,
			), nil
		}
	}

	// Fallback ke .env
	if uc.cfg != nil && uc.cfg.ShopeePartnerID > 0 && uc.cfg.ShopeePartnerKey != "" {
		return pkgShopee.NewClient(
			uc.cfg.ShopeePartnerID,
			uc.cfg.ShopeePartnerKey,
			uc.cfg.ShopeeIsProduction,
			uc.cfg.ShopeeRedirectURL,
		), nil
	}

	return nil, errors.New("kredensial Shopee Partner belum dikonfigurasi di database maupun .env")
}

// GetShopeePushKey mengembalikan PushPartnerKey atau PartnerKey aktif (dari DB atau .env)
func (uc *UseCase) GetShopeePushKey(ctx context.Context) string {
	cred, err := uc.repo.FindByPlatform(ctx, domainCred.PlatformShopee)
	if err == nil && cred != nil && cred.PartnerKey() != "" {
		return cred.PartnerKey()
	}

	if uc.cfg != nil {
		if uc.cfg.ShopeePushPartnerKey != "" {
			return uc.cfg.ShopeePushPartnerKey
		}
		return uc.cfg.ShopeePartnerKey
	}
	return ""
}

