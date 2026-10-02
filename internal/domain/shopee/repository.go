package shopee

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindShop(ctx context.Context, shopID uint64) (*ShopeeShop, error)
	FindDefaultShop(ctx context.Context) (*ShopeeShop, error)
	FindAllShops(ctx context.Context) ([]*ShopeeShop, error)
	SaveShop(ctx context.Context, shop *ShopeeShop) error

	FindOrder(ctx context.Context, orderSN string) (*ShopeeOrder, error)
	FindOrders(ctx context.Context, status, carrier, search string) ([]*ShopeeOrder, error)
	SaveOrder(ctx context.Context, order *ShopeeOrder) error
	LinkSKU(ctx context.Context, itemID, modelID uint64, productID uuid.UUID) error

	GetCashflowSummary(ctx context.Context) (*CashflowSummary, error)
	RecalculateFinances(ctx context.Context) (int, error)
}
