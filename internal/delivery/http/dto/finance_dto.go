package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/finance"
	financeUC "symetra-lab-backend-v2/internal/usecase/finance"
)

// ─── FinanceTransaction DTOs ───────────────────────────────────────────────────

type CreateTransactionRequest struct {
	Type            string  `json:"type"` // INCOME | EXPENSE | CAPITAL_IN
	Category        string  `json:"category"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"` // Format: "2006-01-02"
	ReferenceType   string  `json:"reference_type,omitempty"`
	ReferenceID     string  `json:"reference_id,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

type FinanceTransactionResponse struct {
	ID              uuid.UUID `json:"id"`
	Type            string    `json:"type"`
	Category        string    `json:"category"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionDate string    `json:"transaction_date"`
	ReferenceType   string    `json:"reference_type,omitempty"`
	ReferenceID     string    `json:"reference_id,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func ToFinanceTransactionResponse(t *finance.FinanceTransaction) FinanceTransactionResponse {
	return FinanceTransactionResponse{
		ID:              t.ID(),
		Type:            string(t.Type()),
		Category:        string(t.Category()),
		Amount:          t.Amount(),
		Description:     t.Description(),
		TransactionDate: t.TransactionDate().Format("2006-01-02"),
		ReferenceType:   t.ReferenceType(),
		ReferenceID:     t.ReferenceID(),
		Notes:           t.Notes(),
		CreatedAt:       t.CreatedAt(),
		UpdatedAt:       t.UpdatedAt(),
	}
}

type FinanceTransactionListQuery struct {
	Type     string `query:"type"`
	Category string `query:"category"`
	DateFrom string `query:"date_from"` // Format: "2006-01-02"
	DateTo   string `query:"date_to"`
}

type FinanceSummaryResponse struct {
	TotalIncome    float64            `json:"total_income"`
	TotalExpense   float64            `json:"total_expense"`
	TotalCapitalIn float64            `json:"total_capital_in"`
	NetCashFlow    float64            `json:"net_cash_flow"`
	ByCategory     map[string]float64 `json:"by_category"`
	PeriodFrom     string             `json:"period_from"`
	PeriodTo       string             `json:"period_to"`
}

func ToFinanceSummaryResponse(s *finance.FinanceSummary, from, to time.Time) FinanceSummaryResponse {
	byCat := make(map[string]float64, len(s.ByCategory))
	for k, v := range s.ByCategory {
		byCat[string(k)] = v
	}
	return FinanceSummaryResponse{
		TotalIncome:    s.TotalIncome,
		TotalExpense:   s.TotalExpense,
		TotalCapitalIn: s.TotalCapitalIn,
		NetCashFlow:    s.NetCashFlow,
		ByCategory:     byCat,
		PeriodFrom:     from.Format("2006-01-02"),
		PeriodTo:       to.Format("2006-01-02"),
	}
}

// ─── PurchaseOrder DTOs ────────────────────────────────────────────────────────

type CreatePurchaseOrderItemRequest struct {
	ItemType  string  `json:"item_type"` // FILAMENT | HARDWARE | PACKAGING | OTHER
	ItemName  string  `json:"item_name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"` // kg, pcs, roll, dll.
	UnitPrice float64 `json:"unit_price"`
}

type CreatePurchaseOrderRequest struct {
	SupplierName string                           `json:"supplier_name"`
	PurchaseDate string                           `json:"purchase_date"` // Format: "2006-01-02"
	Notes        string                           `json:"notes,omitempty"`
	Items        []CreatePurchaseOrderItemRequest `json:"items"`
}

type PurchaseOrderItemResponse struct {
	ID         uuid.UUID `json:"id"`
	ItemType   string    `json:"item_type"`
	ItemName   string    `json:"item_name"`
	Quantity   float64   `json:"quantity"`
	Unit       string    `json:"unit"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
}

type PurchaseOrderResponse struct {
	ID                  uuid.UUID                   `json:"id"`
	SupplierName        string                      `json:"supplier_name"`
	PurchaseDate        string                      `json:"purchase_date"`
	TotalAmount         float64                     `json:"total_amount"`
	Status              string                      `json:"status"`
	Notes               string                      `json:"notes,omitempty"`
	LinkedTransactionID *uuid.UUID                  `json:"linked_transaction_id,omitempty"`
	Items               []PurchaseOrderItemResponse `json:"items"`
	CreatedAt           time.Time                   `json:"created_at"`
}

func ToPurchaseOrderItemResponse(i *finance.PurchaseOrderItem) PurchaseOrderItemResponse {
	return PurchaseOrderItemResponse{
		ID:         i.ID(),
		ItemType:   string(i.ItemType()),
		ItemName:   i.ItemName(),
		Quantity:   i.Quantity(),
		Unit:       i.Unit(),
		UnitPrice:  i.UnitPrice(),
		TotalPrice: i.TotalPrice(),
	}
}

func ToPurchaseOrderResponse(po *finance.PurchaseOrder) PurchaseOrderResponse {
	items := make([]PurchaseOrderItemResponse, len(po.Items()))
	for i := range po.Items() {
		it := po.Items()[i]
		items[i] = ToPurchaseOrderItemResponse(&it)
	}
	return PurchaseOrderResponse{
		ID:                  po.ID(),
		SupplierName:        po.SupplierName(),
		PurchaseDate:        po.PurchaseDate().Format("2006-01-02"),
		TotalAmount:         po.TotalAmount(),
		Status:              po.Status(),
		Notes:               po.Notes(),
		LinkedTransactionID: po.LinkedTransactionID(),
		Items:               items,
		CreatedAt:           po.CreatedAt(),
	}
}

// ─── CapitalRecord DTOs ────────────────────────────────────────────────────────

type CreateCapitalRecordRequest struct {
	Type        string  `json:"type"` // INITIAL | ADDITION
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	RecordDate  string  `json:"record_date"` // Format: "2006-01-02"
	Notes       string  `json:"notes,omitempty"`
}

type CapitalRecordResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Type                string     `json:"type"`
	Amount              float64    `json:"amount"`
	Description         string     `json:"description"`
	RecordDate          string     `json:"record_date"`
	LinkedTransactionID *uuid.UUID `json:"linked_transaction_id,omitempty"`
	Notes               string     `json:"notes,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

func ToCapitalRecordResponse(cr *finance.CapitalRecord) CapitalRecordResponse {
	return CapitalRecordResponse{
		ID:                  cr.ID(),
		Type:                string(cr.RecordType()),
		Amount:              cr.Amount(),
		Description:         cr.Description(),
		RecordDate:          cr.RecordDate().Format("2006-01-02"),
		LinkedTransactionID: cr.LinkedTransactionID(),
		Notes:               cr.Notes(),
		CreatedAt:           cr.CreatedAt(),
	}
}

// ─── Input mappers (Request → UseCase Input) ─────────────────────────────────

func ToCreateTransactionInput(req CreateTransactionRequest) (financeUC.CreateTransactionInput, error) {
	date, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		return financeUC.CreateTransactionInput{}, err
	}
	return financeUC.CreateTransactionInput{
		Type:            finance.TransactionType(req.Type),
		Category:        finance.TransactionCategory(req.Category),
		Amount:          req.Amount,
		Description:     req.Description,
		TransactionDate: date,
		ReferenceType:   req.ReferenceType,
		ReferenceID:     req.ReferenceID,
		Notes:           req.Notes,
	}, nil
}

func ToCreatePurchaseOrderInput(req CreatePurchaseOrderRequest) (financeUC.CreatePurchaseOrderInput, error) {
	date, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return financeUC.CreatePurchaseOrderInput{}, err
	}
	items := make([]financeUC.CreatePurchaseOrderItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = financeUC.CreatePurchaseOrderItemInput{
			ItemType:  finance.PurchaseItemType(it.ItemType),
			ItemName:  it.ItemName,
			Quantity:  it.Quantity,
			Unit:      it.Unit,
			UnitPrice: it.UnitPrice,
		}
	}
	return financeUC.CreatePurchaseOrderInput{
		SupplierName: req.SupplierName,
		PurchaseDate: date,
		Notes:        req.Notes,
		Items:        items,
	}, nil
}

func ToCreateCapitalRecordInput(req CreateCapitalRecordRequest) (financeUC.CreateCapitalRecordInput, error) {
	date, err := time.Parse("2006-01-02", req.RecordDate)
	if err != nil {
		return financeUC.CreateCapitalRecordInput{}, err
	}
	return financeUC.CreateCapitalRecordInput{
		Type:        finance.CapitalRecordType(req.Type),
		Amount:      req.Amount,
		Description: req.Description,
		RecordDate:  date,
		Notes:       req.Notes,
	}, nil
}
