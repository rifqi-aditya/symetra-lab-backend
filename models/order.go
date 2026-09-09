package models

import (
	"time"
)

// ShopeeOrder merepresentasikan data transaksi pesanan dari Shopee Open Platform
type ShopeeOrder struct {
	OrderSN           string              `gorm:"primaryKey;size:64;not null" json:"order_sn"`
	ShopID            uint64              `gorm:"index;not null" json:"shop_id"`
	OrderStatus       string              `gorm:"size:32;index;not null" json:"order_status"` // READY_TO_SHIP, PROCESSED, SHIPPED, COMPLETED, CANCELLED
	BuyerUserID       uint64              `json:"buyer_user_id"`
	BuyerUsername     string              `gorm:"size:128" json:"buyer_username"`
	MessageToSeller   string              `gorm:"type:text" json:"message_to_seller"` // Catatan khusus/kustom pembeli
	ShipByDate        int64               `json:"ship_by_date"`                       // Unix timestamp batas akhir pengiriman (SLA)
	ShipByDateTime    *time.Time          `json:"ship_by_date_time,omitempty"`       // Format waktu deadline yang bisa dibaca
	ShippingCarrier   string              `gorm:"size:64" json:"shipping_carrier"`    // J&T, SPX, SiCepat, dll
	TrackingNumber    string              `gorm:"size:64" json:"tracking_number"`     // Nomor resi pengiriman
	TotalAmount       float64             `gorm:"type:decimal(15,2)" json:"total_amount"` // Total nominal yang dibayar pembeli
	BuyerCancelReason string              `gorm:"size:255" json:"buyer_cancel_reason,omitempty"`
	CreateTimeShopee  int64               `json:"create_time_shopee"`
	UpdateTimeShopee  int64               `json:"update_time_shopee"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`

	// Relasi
	Items  []ShopeeOrderItem  `gorm:"foreignKey:OrderSN;references:OrderSN;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Escrow *ShopeeOrderEscrow `gorm:"foreignKey:OrderSN;references:OrderSN;constraint:OnDelete:CASCADE" json:"escrow,omitempty"`
}

func (ShopeeOrder) TableName() string {
	return "shopee_orders"
}

// ShopeeOrderItem merepresentasikan rincian produk/varian dalam satu pesanan Shopee
type ShopeeOrderItem struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderSN         string    `gorm:"size:64;index;not null" json:"order_sn"`
	ItemID          uint64    `gorm:"index" json:"item_id"`
	ItemName        string    `gorm:"size:255" json:"item_name"`
	ItemSKU         string    `gorm:"size:128" json:"item_sku"`
	ModelID         uint64    `json:"model_id"`
	ModelName       string    `gorm:"size:255" json:"model_name"` // Varian: warna filamen, ukuran, dll
	ModelSKU        string    `gorm:"size:128" json:"model_sku"`
	Quantity        int       `json:"quantity"`
	OriginalPrice   float64   `gorm:"type:decimal(15,2)" json:"original_price"`
	DiscountedPrice float64   `gorm:"type:decimal(15,2)" json:"discounted_price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ShopeeOrderItem) TableName() string {
	return "shopee_order_items"
}

// ShopeeOrderEscrow merepresentasikan transparansi finansial dan rincian potongan admin Shopee
type ShopeeOrderEscrow struct {
	OrderSN                  string    `gorm:"primaryKey;size:64;not null" json:"order_sn"`
	EscrowAmount             float64   `gorm:"type:decimal(15,2)" json:"escrow_amount"`               // Nominal bersih masuk saldo seller
	SellingPrice             float64   `gorm:"type:decimal(15,2)" json:"selling_price"`                // Total harga kotor produk
	CommissionFee            float64   `gorm:"type:decimal(15,2)" json:"commission_fee"`              // Biaya admin kategori Shopee
	CommissionRuleName       string    `gorm:"size:255" json:"commission_rule_name,omitempty"`       // Nama aturan komisi kategori
	CommissionPercentage     float64   `gorm:"type:decimal(5,2)" json:"commission_percentage"`        // % potongan komisi kategori
	ServiceFee               float64   `gorm:"type:decimal(15,2)" json:"service_fee"`                 // Biaya Gratis Ongkir XTRA / Cashback XTRA
	ServiceRuleName          string    `gorm:"size:255" json:"service_rule_name,omitempty"`          // Nama program layanan XTRA
	ServicePercentage        float64   `gorm:"type:decimal(5,2)" json:"service_percentage"`           // % potongan layanan XTRA
	SellerTransactionFee     float64   `gorm:"type:decimal(15,2)" json:"seller_transaction_fee"`      // Biaya penanganan transaksi pembayaran (~4%)
	SellerOrderProcessingFee float64   `gorm:"type:decimal(15,2)" json:"seller_order_processing_fee"` // Biaya pemrosesan per pesanan (Rp 1.000 / 1.250)
	SellerVoucherDiscount    float64   `gorm:"type:decimal(15,2)" json:"seller_voucher_discount"`     // Diskon voucher ditanggung seller
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (ShopeeOrderEscrow) TableName() string {
	return "shopee_order_escrows"
}
