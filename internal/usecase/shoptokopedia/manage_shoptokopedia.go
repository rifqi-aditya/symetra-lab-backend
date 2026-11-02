package shoptokopedia

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/config"
	domainOrder "symetra-lab-backend-v2/internal/domain/order"
	domainStore "symetra-lab-backend-v2/internal/domain/shoptokopedia"
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
	pkgTokopedia "symetra-lab-backend-v2/pkg/shoptokopedia"
)

var (
	ErrPartnerNotConfig = errors.New("shop tokopedia credentials are not configured")
	ErrStoreNotFound    = errors.New("shop tokopedia store not found")
)

type LedgerSynchronizer interface {
	SyncRunningBalances(ctx context.Context) error
}

type OrderSaver interface {
	SaveStandardMarketplaceOrder(ctx context.Context, ord *domainOrder.StandardMarketplaceOrder) error
}

type ShopTokopediaUseCases struct {
	cfg        *config.Config
	repo       domainStore.Repository
	client     *pkgTokopedia.Client
	skuMatcher *usecaseProduct.SKUMatcher
	ledgerSync LedgerSynchronizer
}

func NewShopTokopediaUseCases(
	cfg *config.Config,
	repo domainStore.Repository,
	client *pkgTokopedia.Client,
	skuMatcher *usecaseProduct.SKUMatcher,
	ledgerSync LedgerSynchronizer,
) *ShopTokopediaUseCases {
	return &ShopTokopediaUseCases{
		cfg:        cfg,
		repo:       repo,
		client:     client,
		skuMatcher: skuMatcher,
		ledgerSync: ledgerSync,
	}
}

// 1. Auth & Store Onboarding
func (uc *ShopTokopediaUseCases) GetAuthURL(serviceID, state string) (string, error) {
	if uc.client == nil || uc.client.AppKey == "" {
		return "", ErrPartnerNotConfig
	}
	return uc.client.BuildAuthURL(serviceID, state)
}

func (uc *ShopTokopediaUseCases) HandleCallback(ctx context.Context, authCode string) (*domainStore.ShopTokopediaStore, error) {
	if uc.client == nil {
		return nil, ErrPartnerNotConfig
	}

	tokenResp, err := uc.client.GetAccessToken(authCode)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	accessExpires := now.Add(time.Duration(tokenResp.Data.AccessTokenExpireIn) * time.Second)
	refreshExpires := now.Add(time.Duration(tokenResp.Data.RefreshTokenExpireIn) * time.Second)

	sellerName := tokenResp.Data.SellerName
	if sellerName == "" {
		sellerName = fmt.Sprintf("Tokopedia Store (%s)", tokenResp.Data.OpenID)
	}

	store := domainStore.ReconstructStore(
		0,
		tokenResp.Data.OpenID,
		sellerName,
		tokenResp.Data.SellerBaseRegion,
		tokenResp.Data.AccessToken,
		tokenResp.Data.RefreshToken,
		accessExpires,
		refreshExpires,
		now,
		now,
	)

	if err := uc.repo.SaveStore(ctx, store); err != nil {
		return nil, err
	}

	return uc.repo.FindStore(ctx, store.OpenID())
}

func (uc *ShopTokopediaUseCases) RefreshToken(ctx context.Context, openID string) (*domainStore.ShopTokopediaStore, error) {
	store, err := uc.repo.FindStore(ctx, openID)
	if err != nil {
		return nil, err
	}

	if uc.client == nil {
		return nil, ErrPartnerNotConfig
	}

	resp, err := uc.client.RefreshAccessToken(store.RefreshToken())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	store.UpdateTokens(
		resp.Data.AccessToken,
		resp.Data.RefreshToken,
		now.Add(time.Duration(resp.Data.AccessTokenExpireIn)*time.Second),
		now.Add(time.Duration(resp.Data.RefreshTokenExpireIn)*time.Second),
	)

	if err := uc.repo.SaveStore(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

func (uc *ShopTokopediaUseCases) GetValidStore(ctx context.Context, openID string) (*domainStore.ShopTokopediaStore, error) {
	var store *domainStore.ShopTokopediaStore
	var err error

	if openID != "" {
		store, err = uc.repo.FindStore(ctx, openID)
		if err != nil && errors.Is(err, ErrStoreNotFound) {
			store, err = uc.repo.FindDefaultStore(ctx)
		}
	} else {
		store, err = uc.repo.FindDefaultStore(ctx)
	}

	if err != nil {
		return nil, err
	}

	if store.IsTokenExpired() {
		return uc.RefreshToken(ctx, store.OpenID())
	}

	return store, nil
}

// 2. Order Ingestion & Matching
func (uc *ShopTokopediaUseCases) ProcessStandardOrder(
	ctx context.Context,
	stdOrder *domainOrder.StandardMarketplaceOrder,
	cache *usecaseProduct.CostingContext,
) error {
	if stdOrder == nil {
		return nil
	}

	// 1. Match all items against workshop products
	for i := range stdOrder.Items {
		item := &stdOrder.Items[i]
		matchRes := uc.skuMatcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
			UserID:         uuid.Nil,
			ModelSKU:       item.ModelSKU,
			ItemSKU:        item.ItemSKU,
			ChannelItemID:  item.ChannelItemID,
			ChannelModelID: item.ChannelModelID,
			ItemName:       item.ItemName,
			Quantity:       item.Quantity,
			SellingPrice:   item.EffectivePrice(),
			ExistingProdID: item.ProductID,
		}, cache)

		item.ProductID = matchRes.ProductID
		item.MatchedSKU = matchRes.MatchedSKU
		item.MappingStatus = matchRes.MappingStatus
		if matchRes.ProductName != "" {
			item.ItemName = matchRes.ProductName
		}
	}

	// 2. Explicit ledger sync wakeup if configured
	if uc.ledgerSync != nil {
		_ = uc.ledgerSync.SyncRunningBalances(ctx)
	}

	return nil
}

