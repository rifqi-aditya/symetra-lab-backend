package app

import (
	"symetra-lab-backend-v2/internal/delivery/http/handler"
	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/order"
	"symetra-lab-backend-v2/internal/domain/product"
	orderUC "symetra-lab-backend-v2/internal/usecase/order"
)

type OrderModule struct {
	OrderRepo    order.Repository
	OrderHandler *handler.OrderHandler
}

func wireOrderModule(
	orderRepo order.Repository,
	productRepo product.Repository,
	financeRepo finance.Repository,
) *OrderModule {
	listOrdersUC := orderUC.NewListOrdersUseCase(orderRepo)
	getOrderUC := orderUC.NewGetOrderUseCase(orderRepo)
	createOrderUC := orderUC.NewCreateOrderUseCase(orderRepo, productRepo, financeRepo)
	updateOrderStatusUC := orderUC.NewUpdateOrderStatusUseCase(orderRepo, financeRepo)
	updatePaymentStatusUC := orderUC.NewUpdatePaymentStatusUseCase(orderRepo, financeRepo)
	deleteOrderUC := orderUC.NewDeleteOrderUseCase(orderRepo, financeRepo)

	orderHandler := handler.NewOrderHandler(
		listOrdersUC,
		getOrderUC,
		createOrderUC,
		updateOrderStatusUC,
		updatePaymentStatusUC,
		deleteOrderUC,
	)

	return &OrderModule{
		OrderRepo:    orderRepo,
		OrderHandler: orderHandler,
	}
}
