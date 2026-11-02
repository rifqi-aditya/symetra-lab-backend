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
	platformStr, partnerID, partnerKey, pushKey, redirectURL string,
	isProduction bool,
) (*domainCred.MarketplaceCredential, error) {
	platform := domainCred.Platform(strings.ToUpper(strings.TrimSpace(platformStr)))

	// Tangani kasus KEEP_EXISTING jika kredensial sudah ada sebelumnya
	existing, _ := uc.repo.FindByPlatform(ctx, platform)
	if existing != nil {
		if partnerKey == "KEEP_EXISTING" || partnerKey == "" {
			partnerKey = existing.PartnerKey()
		}
		if pushKey == "KEEP_EXISTING" {
			pushKey = existing.PushKey()
		}
	} else {
		if partnerKey == "KEEP_EXISTING" || partnerKey == "" {
			if uc.cfg != nil && uc.cfg.ShopeePartnerKey != "" {
				partnerKey = uc.cfg.ShopeePartnerKey
			} else if platform == domainCred.PlatformShopee {
				partnerKey = "shpk676b516461596a455257534a557973446969466f44455268686f6d727756"
			}
		}
		if pushKey == "KEEP_EXISTING" || pushKey == "" {
			if uc.cfg != nil && uc.cfg.ShopeePushPartnerKey != "" {
				pushKey = uc.cfg.ShopeePushPartnerKey
			} else if platform == domainCred.PlatformShopee {
				pushKey = "4d68734e79635058547376725a6b505765555279544867476d57456f54774455"
			}
		}
	}

	cred, err := domainCred.NewMarketplaceCredential(platform, partnerID, partnerKey, redirectURL, isProduction)
	if err != nil {
		return nil, err
	}
	cred.SetPushKey(pushKey)

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

	// Fallback to verified credentials if DB is unconfigured or corrupted
	if uc.cfg != nil && uc.cfg.ShopeePartnerID > 0 && uc.cfg.ShopeePartnerKey != "" {
		return pkgShopee.NewClient(
			uc.cfg.ShopeePartnerID,
			uc.cfg.ShopeePartnerKey,
			uc.cfg.ShopeeIsProduction,
			uc.cfg.ShopeeRedirectURL,
		), nil
	}
	return pkgShopee.NewClient(
		2044497,
		"shpk676b516461596a455257534a557973446969466f44455268686f6d727756",
		true,
		"https://symetra-lab-backend.vercel.app/api/v1/shopee/callback",
	), nil
}

// GetShopeePushKey mengembalikan PushKey aktif (atau fallback ke PartnerKey) dari database
func (uc *UseCase) GetShopeePushKey(ctx context.Context) string {
	cred, err := uc.repo.FindByPlatform(ctx, domainCred.PlatformShopee)
	if err == nil && cred != nil {
		if cred.PushKey() != "" {
			return cred.PushKey()
		}
		if cred.PartnerKey() != "" {
			return cred.PartnerKey()
		}
	}
	return ""
}

// GetShopeePartnerKey mengembalikan PartnerKey aktif dari database
func (uc *UseCase) GetShopeePartnerKey(ctx context.Context) string {
	cred, err := uc.repo.FindByPlatform(ctx, domainCred.PlatformShopee)
	if err == nil && cred != nil && cred.PartnerKey() != "" {
		return cred.PartnerKey()
	}
	return ""
}

