package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/order"
	"symetra-lab-backend-v2/internal/domain/product"
)

// LedgerSynchronizer defines the interface needed to keep the 8-bucket cash accounts in sync.
type LedgerSynchronizer interface {
	SyncRunningBalances(ctx context.Context) error
}

type ListOrdersUseCase struct {
	repo order.Repository
}

func NewListOrdersUseCase(repo order.Repository) *ListOrdersUseCase {
	return &ListOrdersUseCase{repo: repo}
}

func (uc *ListOrdersUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	return uc.repo.FindAll(ctx, userID)
}

type GetOrderUseCase struct {
	repo order.Repository
}

func NewGetOrderUseCase(repo order.Repository) *GetOrderUseCase {
	return &GetOrderUseCase{repo: repo}
}

func (uc *GetOrderUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*order.Order, error) {
	return uc.repo.FindByID(ctx, userID, id)
}

func (uc *GetOrderUseCase) ExecuteByOrderNumber(ctx context.Context, userID uuid.UUID, orderNumber string) (*order.Order, error) {
	return uc.repo.FindByOrderNumber(ctx, userID, orderNumber)
}

type CreateOrderItemInput struct {
	ProductID      *uuid.UUID
	ProductName    string
	Quantity       int
	SellingPrice   float64
	HPP            float64
	WeightGrams    float64
	PrintTimeHours float64
	MachineID      *uuid.UUID
}

type CreateOrderInput struct {
	CustomerName    string
	CustomerContact string
	Notes           string
	Source          string
	PaymentStatus   string
	Items           []CreateOrderItemInput
}

type CreateOrderUseCase struct {
	orderRepo   order.Repository
	productRepo product.Repository
	syncLedger  LedgerSynchronizer
}

func NewCreateOrderUseCase(orderRepo order.Repository, productRepo product.Repository, syncLedger LedgerSynchronizer) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		syncLedger:  syncLedger,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateOrderInput) (*order.Order, error) {
	domainItems := make([]order.OrderItem, len(input.Items))

	for i, item := range input.Items {
		name := item.ProductName
		price := item.SellingPrice
		hpp := item.HPP
		weight := item.WeightGrams
		printTime := item.PrintTimeHours
		machineID := item.MachineID

		// If product ID supplied, enrich snapshot from product details if needed
		if item.ProductID != nil && *item.ProductID != uuid.Nil {
			p, err := uc.productRepo.FindByID(ctx, userID, *item.ProductID)
			if err == nil && p != nil {
				if name == "" {
					name = p.Name()
				}
				if price <= 0 {
					price = p.BaseSellingPrice()
				}
				if hpp <= 0 {
					hpp = p.BaseHPP()
				}
				if weight <= 0 {
					weight = p.DefaultWeightGrams()
				}
				if printTime <= 0 {
					printTime = p.DefaultPrintTimeHours()
				}
				if machineID == nil {
					machineID = p.DefaultMachineID()
				}
			}
		}

		domainItems[i] = order.ReconstructOrderItem(
			uuid.New(),
			uuid.Nil,
			item.ProductID,
			name,
			item.Quantity,
			price,
			hpp,
			weight,
			printTime,
			machineID,
			0, 0, 0, 0,
			time.Now(),
		)
	}

	ord, err := order.NewOrder(
		userID,
		input.CustomerName,
		input.CustomerContact,
		input.Notes,
		input.Source,
		input.PaymentStatus,
		domainItems,
	)
	if err != nil {
		return nil, err
	}

	if nextOrderNum, errSeq := uc.orderRepo.GetNextManualOrderNumber(ctx, userID, time.Now()); errSeq == nil && nextOrderNum != "" {
		ord.SetOrderNumber(nextOrderNum)
	}

	if err := uc.orderRepo.Create(ctx, ord); err != nil {
		return nil, err
	}

	if uc.syncLedger != nil && ord.PaymentStatus() == "PAID" {
		_ = uc.syncLedger.SyncRunningBalances(ctx)
	}

	return uc.orderRepo.FindByID(ctx, userID, ord.ID())
}

type UpdateOrderStatusUseCase struct {
	repo       order.Repository
	syncLedger LedgerSynchronizer
}

func NewUpdateOrderStatusUseCase(repo order.Repository, syncLedger LedgerSynchronizer) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{repo: repo, syncLedger: syncLedger}
}

func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, userID, id uuid.UUID, status string) error {
	if err := uc.repo.UpdateStatus(ctx, userID, id, status); err != nil {
		return err
	}
	if uc.syncLedger != nil {
		_ = uc.syncLedger.SyncRunningBalances(ctx)
	}
	return nil
}

type UpdatePaymentStatusUseCase struct {
	repo       order.Repository
	syncLedger LedgerSynchronizer
}

func NewUpdatePaymentStatusUseCase(repo order.Repository, syncLedger LedgerSynchronizer) *UpdatePaymentStatusUseCase {
	return &UpdatePaymentStatusUseCase{repo: repo, syncLedger: syncLedger}
}

func (uc *UpdatePaymentStatusUseCase) Execute(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error {
	if err := uc.repo.UpdatePaymentStatus(ctx, userID, id, paymentStatus); err != nil {
		return err
	}
	if uc.syncLedger != nil {
		_ = uc.syncLedger.SyncRunningBalances(ctx)
	}
	return nil
}

type DeleteOrderUseCase struct {
	repo       order.Repository
	syncLedger LedgerSynchronizer
}

func NewDeleteOrderUseCase(repo order.Repository, syncLedger LedgerSynchronizer) *DeleteOrderUseCase {
	return &DeleteOrderUseCase{repo: repo, syncLedger: syncLedger}
}

func (uc *DeleteOrderUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	if err := uc.repo.Delete(ctx, userID, id); err != nil {
		return err
	}
	if uc.syncLedger != nil {
		_ = uc.syncLedger.SyncRunningBalances(ctx)
	}
	return nil
}
