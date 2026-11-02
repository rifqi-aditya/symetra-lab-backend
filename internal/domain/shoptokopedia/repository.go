package shoptokopedia

import (
	"context"
)

type Repository interface {
	FindStore(ctx context.Context, openID string) (*ShopTokopediaStore, error)
	FindDefaultStore(ctx context.Context) (*ShopTokopediaStore, error)
	FindAllStores(ctx context.Context) ([]*ShopTokopediaStore, error)
	SaveStore(ctx context.Context, store *ShopTokopediaStore) error
}

