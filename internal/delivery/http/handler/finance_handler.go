package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	"symetra-lab-backend-v2/internal/domain/finance"
	financeUC "symetra-lab-backend-v2/internal/usecase/finance"
)

type FinanceHandler struct {
	listTxUC     *financeUC.ListTransactionsUseCase
	createTxUC   *financeUC.CreateTransactionUseCase
	deleteTxUC   *financeUC.DeleteTransactionUseCase
	getSummaryUC *financeUC.GetSummaryUseCase

	createPOUC *financeUC.CreatePurchaseOrderUseCase
	listPOUC   *financeUC.ListPurchaseOrdersUseCase
	getPOUC    *financeUC.GetPurchaseOrderUseCase
	deletePOUC *financeUC.DeletePurchaseOrderUseCase

	createCapUC     *financeUC.CreateCapitalRecordUseCase
	listCapUC       *financeUC.ListCapitalRecordsUseCase
	getCapUC        *financeUC.GetCapitalRecordByIDUseCase
	updateCapUC     *financeUC.UpdateCapitalRecordUseCase
	deleteCapUC     *financeUC.DeleteCapitalRecordUseCase
	getTotalCapUC   *financeUC.GetTotalCapitalUseCase
	getCapSummaryUC *financeUC.GetCapitalSummaryUseCase

	getOrderAllocUC    *financeUC.GetOrderAllocationUseCase
	listCashAccountsUC *financeUC.ListCashAccountsUseCase
}

func NewFinanceHandler(
	listTxUC *financeUC.ListTransactionsUseCase,
	createTxUC *financeUC.CreateTransactionUseCase,
	deleteTxUC *financeUC.DeleteTransactionUseCase,
	getSummaryUC *financeUC.GetSummaryUseCase,
	createPOUC *financeUC.CreatePurchaseOrderUseCase,
	listPOUC *financeUC.ListPurchaseOrdersUseCase,
	getPOUC *financeUC.GetPurchaseOrderUseCase,
	deletePOUC *financeUC.DeletePurchaseOrderUseCase,
	createCapUC *financeUC.CreateCapitalRecordUseCase,
	listCapUC *financeUC.ListCapitalRecordsUseCase,
	getCapUC *financeUC.GetCapitalRecordByIDUseCase,
	updateCapUC *financeUC.UpdateCapitalRecordUseCase,
	deleteCapUC *financeUC.DeleteCapitalRecordUseCase,
	getTotalCapUC *financeUC.GetTotalCapitalUseCase,
	getCapSummaryUC *financeUC.GetCapitalSummaryUseCase,
	getOrderAllocUC *financeUC.GetOrderAllocationUseCase,
	listCashAccountsUC *financeUC.ListCashAccountsUseCase,
) *FinanceHandler {
	return &FinanceHandler{
		listTxUC:           listTxUC,
		createTxUC:         createTxUC,
		deleteTxUC:         deleteTxUC,
		getSummaryUC:       getSummaryUC,
		createPOUC:         createPOUC,
		listPOUC:           listPOUC,
		getPOUC:            getPOUC,
		deletePOUC:         deletePOUC,
		createCapUC:        createCapUC,
		listCapUC:          listCapUC,
		getCapUC:           getCapUC,
		updateCapUC:        updateCapUC,
		deleteCapUC:        deleteCapUC,
		getTotalCapUC:      getTotalCapUC,
		getCapSummaryUC:    getCapSummaryUC,
		getOrderAllocUC:    getOrderAllocUC,
		listCashAccountsUC: listCashAccountsUC,
	}
}

// Zona waktu Indonesia Barat (WIB = UTC+7)
var wibLocation = time.FixedZone("WIB", 7*3600)

func parseWIBDateRange(fromStr, toStr string) (*time.Time, *time.Time) {
	var fromPtr, toPtr *time.Time
	if fromStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", fromStr, wibLocation); err == nil {
			fromPtr = &t
		}
	}
	if toStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", toStr, wibLocation); err == nil {
			end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, wibLocation)
			toPtr = &end
		}
	}
	return fromPtr, toPtr
}

// ─── Transactions ─────────────────────────────────────────────────────────────

// GET /api/v1/finance/transactions?type=EXPENSE&category=FILAMENT&date_from=2026-01-01&date_to=2026-12-31
func (h *FinanceHandler) ListTransactions(c echo.Context) error {
	var q dto.FinanceTransactionListQuery
	if err := c.Bind(&q); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid query params", err.Error()))
	}

	filter := finance.TransactionFilter{
		Type:     finance.TransactionType(q.Type),
		Category: finance.TransactionCategory(q.Category),
	}
	fromPtr, toPtr := parseWIBDateRange(q.DateFrom, q.DateTo)
	filter.DateFrom = fromPtr
	filter.DateTo = toPtr

	txs, err := h.listTxUC.Execute(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve transactions", err.Error()))
	}

	resp := make([]dto.FinanceTransactionResponse, len(txs))
	for i, tx := range txs {
		resp[i] = dto.ToFinanceTransactionResponse(tx)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

// POST /api/v1/finance/transactions
func (h *FinanceHandler) CreateTransaction(c echo.Context) error {
	var req dto.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid request body", err.Error()))
	}

	input, err := dto.ToCreateTransactionInput(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid date format (use YYYY-MM-DD)", err.Error()))
	}

	tx, err := h.createTxUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create transaction", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToFinanceTransactionResponse(tx)))
}

// DELETE /api/v1/finance/transactions/:id
func (h *FinanceHandler) DeleteTransaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid transaction ID", err.Error()))
	}

	if err := h.deleteTxUC.Execute(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete transaction", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Transaction deleted"}))
}

// GET /api/v1/finance/summary?date_from=2026-01-01&date_to=2026-12-31
func (h *FinanceHandler) GetSummary(c echo.Context) error {
	fromStr := c.QueryParam("date_from")
	if fromStr == "" {
		fromStr = c.QueryParam("from")
	}
	toStr := c.QueryParam("date_to")
	if toStr == "" {
		toStr = c.QueryParam("to")
	}

	now := time.Now().In(wibLocation)
	// Default: rentang seluruh waktu jika tidak ada filter parameter
	from := time.Date(2000, 1, 1, 0, 0, 0, 0, wibLocation)
	to := now.AddDate(1, 0, 0)

	fromPtr, toPtr := parseWIBDateRange(fromStr, toStr)
	if fromPtr != nil {
		from = *fromPtr
	}
	if toPtr != nil {
		to = *toPtr
	}

	summary, err := h.getSummaryUC.Execute(c.Request().Context(), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to get summary", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToFinanceSummaryResponse(summary, from, to)))
}

// ─── Purchase Orders ──────────────────────────────────────────────────────────

// GET /api/v1/finance/purchase-orders
func (h *FinanceHandler) ListPurchaseOrders(c echo.Context) error {
	pos, err := h.listPOUC.Execute(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve purchase orders", err.Error()))
	}
	resp := make([]dto.PurchaseOrderResponse, len(pos))
	for i, po := range pos {
		resp[i] = dto.ToPurchaseOrderResponse(po)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

// GET /api/v1/finance/purchase-orders/:id
func (h *FinanceHandler) GetPurchaseOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid ID", err.Error()))
	}
	po, err := h.getPOUC.Execute(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Purchase order not found", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.Success(dto.ToPurchaseOrderResponse(po)))
}

// POST /api/v1/finance/purchase-orders
func (h *FinanceHandler) CreatePurchaseOrder(c echo.Context) error {
	var req dto.CreatePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid request body", err.Error()))
	}

	input, err := dto.ToCreatePurchaseOrderInput(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid date format (use YYYY-MM-DD)", err.Error()))
	}

	po, err := h.createPOUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create purchase order", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToPurchaseOrderResponse(po)))
}

// DELETE /api/v1/finance/purchase-orders/:id
func (h *FinanceHandler) DeletePurchaseOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid ID", err.Error()))
	}
	if err := h.deletePOUC.Execute(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete purchase order", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Purchase order deleted"}))
}

// ─── Capital Records ──────────────────────────────────────────────────────────

// GET /api/v1/finance/capital
func (h *FinanceHandler) ListCapitalRecords(c echo.Context) error {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	search := strings.TrimSpace(c.QueryParam("search"))
	typeFilter := strings.TrimSpace(c.QueryParam("type"))
	fromStr := c.QueryParam("date_from")
	toStr := c.QueryParam("date_to")

	page := 0
	limit := 0
	isPaginated := false
	if pageStr != "" || limitStr != "" {
		isPaginated = true
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			page = 1
		}
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			limit = 25
		}
	}

	fromPtr, toPtr := parseWIBDateRange(fromStr, toStr)

	filter := finance.CapitalFilter{
		Type:     finance.CapitalRecordType(typeFilter),
		Search:   search,
		DateFrom: fromPtr,
		DateTo:   toPtr,
		Page:     page,
		Limit:    limit,
	}

	records, totalCount, err := h.listCapUC.ExecuteFiltered(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve capital records", err.Error()))
	}

	resp := make([]dto.CapitalRecordResponse, len(records))
	for i, cr := range records {
		resp[i] = dto.ToCapitalRecordResponse(cr)
	}

	if isPaginated {
		totalPages := 0
		if totalCount > 0 && limit > 0 {
			totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
		}
		return c.JSON(http.StatusOK, dto.SuccessPaginated(resp, totalCount, page, limit, totalPages))
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

// GET /api/v1/finance/capital/:id
func (h *FinanceHandler) GetCapitalRecordByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid capital record ID", err.Error()))
	}

	cr, err := h.getCapUC.Execute(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Capital record not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToCapitalRecordResponse(cr)))
}

// POST /api/v1/finance/capital
func (h *FinanceHandler) CreateCapitalRecord(c echo.Context) error {
	var req dto.CreateCapitalRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid request body", err.Error()))
	}

	input, err := dto.ToCreateCapitalRecordInput(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid date format (use YYYY-MM-DD)", err.Error()))
	}

	cr, err := h.createCapUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create capital record", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToCapitalRecordResponse(cr)))
}

// PUT /api/v1/finance/capital/:id
func (h *FinanceHandler) UpdateCapitalRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid capital record ID", err.Error()))
	}

	var req dto.UpdateCapitalRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid request body", err.Error()))
	}

	input, err := dto.ToUpdateCapitalRecordInput(id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid date format (use YYYY-MM-DD)", err.Error()))
	}

	cr, err := h.updateCapUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update capital record", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToCapitalRecordResponse(cr)))
}

// GET /api/v1/finance/capital/total
func (h *FinanceHandler) GetTotalCapital(c echo.Context) error {
	total, err := h.getTotalCapUC.Execute(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to get total capital", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.Success(map[string]float64{"total_capital": total}))
}

// GET /api/v1/finance/capital/summary
func (h *FinanceHandler) GetCapitalSummary(c echo.Context) error {
	summary, err := h.getCapSummaryUC.Execute(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to get capital summary", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.Success(dto.CapitalSummaryResponse{
		TotalCapital:      summary.TotalCapital,
		InitialCapital:    summary.InitialCapital,
		AdditionalCapital: summary.AdditionalCapital,
		TotalCount:        summary.TotalCount,
	}))
}

// DELETE /api/v1/finance/capital/:id
func (h *FinanceHandler) DeleteCapitalRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid capital record ID", err.Error()))
	}

	if err := h.deleteCapUC.Execute(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete capital record", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Capital record deleted successfully"}))
}

// GET /api/v1/finance/allocations?channel=ALL&date_from=2026-01-01&date_to=2026-12-31
func (h *FinanceHandler) GetOrderAllocations(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	channel := c.QueryParam("channel")
	if channel == "" {
		channel = "ALL"
	}

	fromPtr, toPtr := parseWIBDateRange(c.QueryParam("date_from"), c.QueryParam("date_to"))

	summary, err := h.getOrderAllocUC.Execute(c.Request().Context(), userID, channel, fromPtr, toPtr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve order allocations", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.OrderAllocationSummaryResponse{
		TotalOrders:          summary.TotalOrders,
		TotalGrossSales:      summary.TotalGrossSales,
		TotalChannelFees:     summary.TotalChannelFees,
		TotalNetRevenue:      summary.TotalNetRevenue,
		TotalCOGS:            summary.TotalCOGS,
		TotalNetProfit:       summary.TotalNetProfit,
		AverageProfitMargin:  summary.AverageProfitMargin,
		FundFilament:         summary.FundFilament,
		FundComponent:        summary.FundComponent,
		FundPackaging:        summary.FundPackaging,
		FundElectricity:      summary.FundElectricity,
		FundMaintenance:      summary.FundMaintenance,
		FundDepreciation:     summary.FundDepreciation,
		FundNetProfit:        summary.FundNetProfit,
		UnmappedItemsCount:   summary.UnmappedItemsCount,
		AllocatedFilament:     summary.AllocatedFilament,
		AllocatedComponent:    summary.AllocatedComponent,
		AllocatedPackaging:    summary.AllocatedPackaging,
		AllocatedElectricity:  summary.AllocatedElectricity,
		AllocatedMaintenance:  summary.AllocatedMaintenance,
		AllocatedDepreciation: summary.AllocatedDepreciation,
		AllocatedNetProfit:    summary.AllocatedNetProfit,
		SpentFilament:         summary.SpentFilament,
		SpentComponent:        summary.SpentComponent,
		SpentPackaging:        summary.SpentPackaging,
		SpentElectricity:      summary.SpentElectricity,
		SpentMaintenance:      summary.SpentMaintenance,
		SpentDepreciation:     summary.SpentDepreciation,
		SpentNetProfit:        summary.SpentNetProfit,
		PendingOrdersCount:    summary.PendingOrdersCount,
		PendingGrossSales:     summary.PendingGrossSales,
		PendingEscrowAmount:   summary.PendingEscrowAmount,
		PendingNetProfit:      summary.PendingNetProfit,
		TotalCompletedOrders:  summary.TotalOrders,
		TotalEscrowNetIn:      summary.TotalNetRevenue,
		TotalMarketplaceFees:  summary.TotalChannelFees,
		TotalHPP:              summary.TotalCOGS,
		KasFilamen:            summary.FundFilament,
		KasKomponen:           summary.FundComponent,
		KasPacking:            summary.FundPackaging,
		KasListrik:            summary.FundElectricity,
		KasMaintenance:        summary.FundMaintenance,
		KasDepresiasi:         summary.FundDepreciation,
		KasLabaBersih:         summary.FundNetProfit,
	}))
}

// GET /api/v1/finance/cash-accounts
func (h *FinanceHandler) ListCashAccounts(c echo.Context) error {
	accounts, err := h.listCashAccountsUC.Execute(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve cash accounts", err.Error()))
	}
	resp := make([]dto.CashAccountResponse, len(accounts))
	for i, ca := range accounts {
		resp[i] = dto.ToCashAccountResponse(ca)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}
