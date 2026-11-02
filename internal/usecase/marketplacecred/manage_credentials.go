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
			isProd := cred.IsProduction()
			if uc.cfg != nil && (strings.EqualFold(uc.cfg.AppEnv, "production") || strings.EqualFold(uc.cfg.AppEnv, "prod")) {
				isProd = true
			}

			if redirectURL == "" || (isProd && strings.Contains(redirectURL, "localhost")) {
				if uc.cfg != nil && uc.cfg.ShopeeRedirectURL != "" && (!isProd || !strings.Contains(uc.cfg.ShopeeRedirectURL, "localhost")) {
					redirectURL = uc.cfg.ShopeeRedirectURL
				} else if isProd {
					redirectURL = "https://symetra-lab-backend.vercel.app/api/v1/shopee/callback"
				} else {
					redirectURL = "http://localhost:8080/api/v1/shopee/callback"
				}
			}

			return pkgShopee.NewClient(
				partnerIDInt,
				cred.PartnerKey(),
				isProd,
				redirectURL,
			), nil
		}
	}

	return nil, errors.New("kredensial Shopee Partner belum dikonfigurasi di database. Silakan isi melalui tombol Kredensial API di halaman Marketplace")
}

// GetShopeePushKey mengembalikan PartnerKey aktif dari database
func (uc *UseCase) GetShopeePushKey(ctx context.Context) string {
	cred, err := uc.repo.FindByPlatform(ctx, domainCred.PlatformShopee)
	if err == nil && cred != nil && cred.PartnerKey() != "" {
		return cred.PartnerKey()
	}
	return ""
}

