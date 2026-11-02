package app

import (
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/delivery/http/handler"
	"symetra-lab-backend-v2/internal/domain/dashboard"
	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/order"
	postgresRepo "symetra-lab-backend-v2/internal/repository/postgres"
	dashboardUC "symetra-lab-backend-v2/internal/usecase/dashboard"
	financeUC "symetra-lab-backend-v2/internal/usecase/finance"
)

type FinanceModule struct {
	FinanceRepo   finance.Repository
	DashboardRepo dashboard.Repository

	FinanceHandler   *handler.FinanceHandler
	DashboardHandler *handler.DashboardHandler
}

func wireFinanceModule(db *gorm.DB, orderRepo order.Repository) *FinanceModule {
	financeRepo := postgresRepo.NewFinanceRepository(db)
	dashboardRepo := postgresRepo.NewDashboardRepository(db)

	listTransactionsUC := financeUC.NewListTransactionsUseCase(financeRepo)
	createTransactionUC := financeUC.NewCreateTransactionUseCase(financeRepo)
	deleteTransactionUC := financeUC.NewDeleteTransactionUseCase(financeRepo)
	getSummaryUC := financeUC.NewGetSummaryUseCase(financeRepo)
	createPurchaseOrderUC := financeUC.NewCreatePurchaseOrderUseCase(financeRepo)
	listPurchaseOrdersUC := financeUC.NewListPurchaseOrdersUseCase(financeRepo)
	getPurchaseOrderUC := financeUC.NewGetPurchaseOrderUseCase(financeRepo)
	deletePurchaseOrderUC := financeUC.NewDeletePurchaseOrderUseCase(financeRepo)
	createCapitalRecordUC := financeUC.NewCreateCapitalRecordUseCase(financeRepo)
	listCapitalRecordsUC := financeUC.NewListCapitalRecordsUseCase(financeRepo)
	getCapitalRecordUC := financeUC.NewGetCapitalRecordByIDUseCase(financeRepo)
	updateCapitalRecordUC := financeUC.NewUpdateCapitalRecordUseCase(financeRepo)
	deleteCapitalRecordUC := financeUC.NewDeleteCapitalRecordUseCase(financeRepo)
	getTotalCapitalUC := financeUC.NewGetTotalCapitalUseCase(financeRepo)
	getCapitalSummaryUC := financeUC.NewGetCapitalSummaryUseCase(financeRepo)
	getOrderAllocUC := financeUC.NewGetOrderAllocationUseCase(orderRepo)
	listCashAccountsUC := financeUC.NewListCashAccountsUseCase(financeRepo)

	financeHandler := handler.NewFinanceHandler(
		listTransactionsUC,
		createTransactionUC,
		deleteTransactionUC,
		getSummaryUC,
		createPurchaseOrderUC,
		listPurchaseOrdersUC,
		getPurchaseOrderUC,
		deletePurchaseOrderUC,
		createCapitalRecordUC,
		listCapitalRecordsUC,
		getCapitalRecordUC,
		updateCapitalRecordUC,
		deleteCapitalRecordUC,
		getTotalCapitalUC,
		getCapitalSummaryUC,
		getOrderAllocUC,
		listCashAccountsUC,
	)

	getDashboardOverviewUC := dashboardUC.NewGetDashboardOverviewUseCase(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(getDashboardOverviewUC)

	return &FinanceModule{
		FinanceRepo:      financeRepo,
		DashboardRepo:    dashboardRepo,
		FinanceHandler:   financeHandler,
		DashboardHandler: dashboardHandler,
	}
}

