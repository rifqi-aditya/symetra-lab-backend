package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/finance"
)

// ─── GORM Structs ─────────────────────────────────────────────────────────────

type financeTransactionGORM struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	Type            string     `gorm:"column:type"`
	Category        string     `gorm:"column:category"`
	CashAccountID   *uuid.UUID `gorm:"column:cash_account_id;type:uuid"`
	Amount          float64    `gorm:"column:amount"`
	Description     string     `gorm:"column:description"`
	TransactionDate time.Time  `gorm:"column:transaction_date"`
	ReferenceType   string     `gorm:"column:reference_type"`
	ReferenceID     string     `gorm:"column:reference_id"`
	Notes           string     `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (financeTransactionGORM) TableName() string { return "finance_transactions" }

type cashAccountGORM struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	Name            string    `gorm:"column:name"`
	Code            string    `gorm:"column:code"`
	Description     string    `gorm:"column:description"`
	Color           string    `gorm:"column:color"`
	AllocatedAmount float64   `gorm:"column:allocated_amount"`
	SpentAmount     float64   `gorm:"column:spent_amount"`
	CurrentBalance  float64   `gorm:"column:current_balance"`
	IsActive        bool      `gorm:"column:is_active"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (cashAccountGORM) TableName() string { return "cash_accounts" }

type purchaseOrderGORM struct {
	ID                  uuid.UUID               `gorm:"column:id;primaryKey;type:uuid"`
	SupplierName        string                  `gorm:"column:supplier_name"`
	PurchaseDate        time.Time               `gorm:"column:purchase_date"`
	TotalAmount         float64                 `gorm:"column:total_amount"`
	Status              string                  `gorm:"column:status"`
	Notes               string                  `gorm:"column:notes"`
	LinkedTransactionID *uuid.UUID              `gorm:"column:linked_transaction_id;type:uuid"`
	CreatedAt           time.Time               `gorm:"column:created_at"`
	UpdatedAt           time.Time               `gorm:"column:updated_at"`
	Items               []purchaseOrderItemGORM `gorm:"foreignKey:PurchaseOrderID;references:ID;constraint:OnDelete:CASCADE"`
}

func (purchaseOrderGORM) TableName() string { return "purchase_orders" }

type purchaseOrderItemGORM struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	PurchaseOrderID uuid.UUID `gorm:"column:purchase_order_id;type:uuid"`
	ItemType        string    `gorm:"column:item_type"`
	ItemName        string    `gorm:"column:item_name"`
	Quantity        float64   `gorm:"column:quantity"`
	Unit            string    `gorm:"column:unit"`
	UnitPrice       float64   `gorm:"column:unit_price"`
	TotalPrice      float64   `gorm:"column:total_price"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (purchaseOrderItemGORM) TableName() string { return "purchase_order_items" }

type capitalRecordGORM struct {
	ID                  uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	Type                string     `gorm:"column:type"`
	Amount              float64    `gorm:"column:amount"`
	Description         string     `gorm:"column:description"`
	RecordDate          time.Time  `gorm:"column:record_date"`
	LinkedTransactionID *uuid.UUID `gorm:"column:linked_transaction_id;type:uuid"`
	Notes               string     `gorm:"column:notes"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (capitalRecordGORM) TableName() string { return "capital_records" }

// ─── Repository Struct ────────────────────────────────────────────────────────

type FinanceRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func mapFinanceTxGORMToDomain(g *financeTransactionGORM) *finance.FinanceTransaction {
	return finance.ReconstructFinanceTransaction(
		g.ID,
		finance.TransactionType(g.Type),
		finance.TransactionCategory(g.Category),
		g.CashAccountID,
		g.Amount,
		g.Description,
		g.TransactionDate,
		g.ReferenceType,
		g.ReferenceID,
		g.Notes,
		g.CreatedAt,
		g.UpdatedAt,
	)
}

func mapPurchaseOrderItemGORMToDomain(g *purchaseOrderItemGORM) finance.PurchaseOrderItem {
	return finance.ReconstructPurchaseOrderItem(
		g.ID, g.PurchaseOrderID,
		finance.PurchaseItemType(g.ItemType),
		g.ItemName, g.Quantity, g.Unit,
		g.UnitPrice, g.TotalPrice, g.CreatedAt,
	)
}

func mapPurchaseOrderGORMToDomain(g *purchaseOrderGORM) *finance.PurchaseOrder {
	items := make([]finance.PurchaseOrderItem, len(g.Items))
	for i, it := range g.Items {
		items[i] = mapPurchaseOrderItemGORMToDomain(&it)
	}
	return finance.ReconstructPurchaseOrder(
		g.ID, g.SupplierName, g.PurchaseDate,
		g.TotalAmount, g.Status, g.Notes,
		g.LinkedTransactionID, items,
		g.CreatedAt, g.UpdatedAt,
	)
}

func mapCapitalRecordGORMToDomain(g *capitalRecordGORM) *finance.CapitalRecord {
	return finance.ReconstructCapitalRecord(
		g.ID,
		finance.CapitalRecordType(g.Type),
		g.Amount,
		g.Description,
		g.RecordDate,
		g.LinkedTransactionID,
		g.Notes,
		g.CreatedAt,
		g.UpdatedAt,
	)
}

// ─── FinanceTransaction Methods ───────────────────────────────────────────────

func (r *FinanceRepository) CreateTransaction(ctx context.Context, t *finance.FinanceTransaction) error {
	g := financeTransactionGORM{
		ID:              t.ID(),
		Type:            string(t.Type()),
		Category:        string(t.Category()),
		CashAccountID:   t.CashAccountID(),
		Amount:          t.Amount(),
		Description:     t.Description(),
		TransactionDate: t.TransactionDate(),
		ReferenceType:   t.ReferenceType(),
		ReferenceID:     t.ReferenceID(),
		Notes:           t.Notes(),
		CreatedAt:       t.CreatedAt(),
		UpdatedAt:       t.UpdatedAt(),
	}
	if err := r.db.WithContext(ctx).Create(&g).Error; err != nil {
		return err
	}
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

// CreateTransactionIfNotExists: idempotent insert — skip if reference already exists
func (r *FinanceRepository) CreateTransactionIfNotExists(ctx context.Context, t *finance.FinanceTransaction) error {
	g := financeTransactionGORM{
		ID:              t.ID(),
		Type:            string(t.Type()),
		Category:        string(t.Category()),
		CashAccountID:   t.CashAccountID(),
		Amount:          t.Amount(),
		Description:     t.Description(),
		TransactionDate: t.TransactionDate(),
		ReferenceType:   t.ReferenceType(),
		ReferenceID:     t.ReferenceID(),
		Notes:           t.Notes(),
		CreatedAt:       t.CreatedAt(),
		UpdatedAt:       t.UpdatedAt(),
	}

	var existing financeTransactionGORM
	err := r.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ?", t.ReferenceType(), t.ReferenceID()).
		First(&existing).Error
	if err == nil {
		// Sudah ada, skip
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := r.db.WithContext(ctx).Create(&g).Error; err != nil {
		return err
	}
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

// UpsertShopeeEscrowTransaction creates or updates an escrow transaction with the exact settlement date
func (r *FinanceRepository) UpsertShopeeEscrowTransaction(ctx context.Context, t *finance.FinanceTransaction) error {
	var existing financeTransactionGORM
	err := r.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ?", t.ReferenceType(), t.ReferenceID()).
		First(&existing).Error
	if err == nil {
		// Update transaction_date dan amount jika ada pembaruan tanggal pencairan
		updateFields := map[string]interface{}{
			"transaction_date": t.TransactionDate(),
			"amount":           t.Amount(),
			"updated_at":       time.Now(),
		}
		if t.Description() != "" {
			updateFields["description"] = t.Description()
		}
		if err := r.db.WithContext(ctx).Model(&existing).Updates(updateFields).Error; err != nil {
			return err
		}
		_ = SyncCashAccountLedger(ctx, r.db)
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	g := financeTransactionGORM{
		ID:              t.ID(),
		Type:            string(t.Type()),
		Category:        string(t.Category()),
		CashAccountID:   t.CashAccountID(),
		Amount:          t.Amount(),
		Description:     t.Description(),
		TransactionDate: t.TransactionDate(),
		ReferenceType:   t.ReferenceType(),
		ReferenceID:     t.ReferenceID(),
		Notes:           t.Notes(),
		CreatedAt:       t.CreatedAt(),
		UpdatedAt:       t.UpdatedAt(),
	}

	if err := r.db.WithContext(ctx).Create(&g).Error; err != nil {
		return err
	}
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

func (r *FinanceRepository) FindTransactionByID(ctx context.Context, id uuid.UUID) (*finance.FinanceTransaction, error) {
	var g financeTransactionGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, finance.ErrTransactionNotFound
		}
		return nil, err
	}
	return mapFinanceTxGORMToDomain(&g), nil
}

func (r *FinanceRepository) FindTransactions(ctx context.Context, filter finance.TransactionFilter) ([]*finance.FinanceTransaction, error) {
	query := r.db.WithContext(ctx).Model(&financeTransactionGORM{})
	if filter.Type != "" {
		query = query.Where("type = ?", string(filter.Type))
	}
	if filter.Category != "" {
		query = query.Where("category = ?", string(filter.Category))
	}
	if filter.DateFrom != nil {
		query = query.Where("transaction_date >= ?", filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("transaction_date <= ?", filter.DateTo)
	}
	var list []financeTransactionGORM
	if err := query.Order("transaction_date DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*finance.FinanceTransaction, len(list))
	for i := range list {
		results[i] = mapFinanceTxGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *FinanceRepository) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txRecord financeTransactionGORM
		if err := tx.Where("id = ?", id).First(&txRecord).Error; err == nil {
			// Jika berasal dari modal usaha (CAPITAL_RECORD), hapus juga dari capital_records
			if txRecord.ReferenceType == "CAPITAL_RECORD" && txRecord.ReferenceID != "" {
				if crID, parseErr := uuid.Parse(txRecord.ReferenceID); parseErr == nil {
					_ = tx.Where("id = ?", crID).Delete(&capitalRecordGORM{}).Error
				}
			}
		}

		if err := tx.Where("id = ?", id).Delete(&financeTransactionGORM{}).Error; err != nil {
			return err
		}
		_ = SyncCashAccountLedger(ctx, tx)
		return nil
	})
}

func (r *FinanceRepository) GetSummary(ctx context.Context, from, to time.Time) (*finance.FinanceSummary, error) {
	type row struct {
		Type     string  `gorm:"column:type"`
		Category string  `gorm:"column:category"`
		Total    float64 `gorm:"column:total"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT type, category, COALESCE(SUM(amount), 0) AS total
		FROM finance_transactions
		WHERE transaction_date >= ? AND transaction_date <= ?
		GROUP BY type, category
	`, from, to).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	summary := &finance.FinanceSummary{
		ByCategory: make(map[finance.TransactionCategory]float64),
	}
	for _, rw := range rows {
		summary.ByCategory[finance.TransactionCategory(rw.Category)] += rw.Total
		switch finance.TransactionType(rw.Type) {
		case finance.TypeIncome:
			summary.TotalIncome += rw.Total
		case finance.TypeExpense:
			summary.TotalExpense += rw.Total
		case finance.TypeCapitalIn:
			summary.TotalCapitalIn += rw.Total
		}
	}
	summary.NetCashFlow = summary.TotalIncome - summary.TotalExpense
	return summary, nil
}

// ─── PurchaseOrder Methods ────────────────────────────────────────────────────

func (r *FinanceRepository) CreatePurchaseOrder(ctx context.Context, po *finance.PurchaseOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		g := purchaseOrderGORM{
			ID:                  po.ID(),
			SupplierName:        po.SupplierName(),
			PurchaseDate:        po.PurchaseDate(),
			TotalAmount:         po.TotalAmount(),
			Status:              po.Status(),
			Notes:               po.Notes(),
			LinkedTransactionID: po.LinkedTransactionID(),
			CreatedAt:           po.CreatedAt(),
			UpdatedAt:           po.UpdatedAt(),
		}
		if err := tx.Create(&g).Error; err != nil {
			return err
		}
		for _, item := range po.Items() {
			ig := purchaseOrderItemGORM{
				ID:              item.ID(),
				PurchaseOrderID: po.ID(),
				ItemType:        string(item.ItemType()),
				ItemName:        item.ItemName(),
				Quantity:        item.Quantity(),
				Unit:            item.Unit(),
				UnitPrice:       item.UnitPrice(),
				TotalPrice:      item.TotalPrice(),
				CreatedAt:       item.CreatedAt(),
			}
			if err := tx.Create(&ig).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *FinanceRepository) FindPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*finance.PurchaseOrder, error) {
	var g purchaseOrderGORM
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, finance.ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	return mapPurchaseOrderGORMToDomain(&g), nil
}

func (r *FinanceRepository) FindPurchaseOrders(ctx context.Context) ([]*finance.PurchaseOrder, error) {
	var list []purchaseOrderGORM
	if err := r.db.WithContext(ctx).Preload("Items").Order("purchase_date DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*finance.PurchaseOrder, len(list))
	for i := range list {
		results[i] = mapPurchaseOrderGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *FinanceRepository) UpdatePurchaseOrderTransaction(ctx context.Context, poID, txID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&purchaseOrderGORM{}).
		Where("id = ?", poID).
		Updates(map[string]interface{}{
			"linked_transaction_id": txID,
			"updated_at":            time.Now(),
		}).Error
}

func (r *FinanceRepository) DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete linked transaction in finance_transactions
		var po purchaseOrderGORM
		if err := tx.Where("id = ?", id).First(&po).Error; err == nil {
			if po.LinkedTransactionID != nil {
				_ = tx.Where("id = ?", *po.LinkedTransactionID).Delete(&financeTransactionGORM{}).Error
			}
		}
		_ = tx.Where("reference_type = 'PURCHASE_ORDER' AND reference_id = ?", id.String()).Delete(&financeTransactionGORM{}).Error

		// 2. Delete purchase_order_items
		_ = tx.Where("purchase_order_id = ?", id).Delete(&purchaseOrderItemGORM{}).Error

		// 3. Delete purchase_order
		return tx.Where("id = ?", id).Delete(&purchaseOrderGORM{}).Error
	})
}

// ─── CapitalRecord Methods ────────────────────────────────────────────────────

func (r *FinanceRepository) CreateCapitalRecord(ctx context.Context, cr *finance.CapitalRecord) error {
	g := capitalRecordGORM{
		ID:                  cr.ID(),
		Type:                string(cr.RecordType()),
		Amount:              cr.Amount(),
		Description:         cr.Description(),
		RecordDate:          cr.RecordDate(),
		LinkedTransactionID: cr.LinkedTransactionID(),
		Notes:               cr.Notes(),
		CreatedAt:           cr.CreatedAt(),
		UpdatedAt:           cr.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&g).Error
}

func (r *FinanceRepository) FindCapitalRecords(ctx context.Context) ([]*finance.CapitalRecord, error) {
	var list []capitalRecordGORM
	if err := r.db.WithContext(ctx).Order("record_date DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*finance.CapitalRecord, len(list))
	for i := range list {
		results[i] = mapCapitalRecordGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *FinanceRepository) DeleteCapitalRecord(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Hapus transaksi terkait di finance_transactions jika ada
		_ = tx.Where("reference_type = 'CAPITAL_RECORD' AND reference_id = ?", id.String()).Delete(&financeTransactionGORM{}).Error

		// 2. Hapus dari capital_records
		if err := tx.Where("id = ?", id).Delete(&capitalRecordGORM{}).Error; err != nil {
			return err
		}
		_ = SyncCashAccountLedger(ctx, tx)
		return nil
	})
}

func (r *FinanceRepository) GetTotalCapital(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&capitalRecordGORM{}).
		Where("type IN ?", []string{"INITIAL", "ADDITION"}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// ─── CashAccount Methods ─────────────────────────────────────────────────────────

func (r *FinanceRepository) ListCashAccounts(ctx context.Context) ([]*finance.CashAccount, error) {
	var list []cashAccountGORM
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	results := make([]*finance.CashAccount, len(list))
	for i, g := range list {
		results[i] = finance.ReconstructCashAccount(
			g.ID, g.Name, g.Code, g.Description, g.Color,
			g.AllocatedAmount, g.SpentAmount, g.CurrentBalance,
			g.IsActive, g.CreatedAt, g.UpdatedAt,
		)
	}
	return results, nil
}

func (r *FinanceRepository) SyncRunningBalances(ctx context.Context) error {
	return SyncCashAccountLedger(ctx, r.db)
}
