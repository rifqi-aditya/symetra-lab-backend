package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ShopeeEscrowResponse merepresentasikan respon dari /api/v2/payment/get_escrow_detail
type ShopeeEscrowResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		OrderSN     string            `json:"order_sn"`
		OrderIncome ShopeeOrderIncome `json:"order_income"`
	} `json:"response"`
}

type ShopeeOrderIncome struct {
	EscrowAmount             float64                  `json:"escrow_amount"`               // Uang bersih masuk saldo seller
	BuyerTotalAmount         float64                  `json:"buyer_total_amount"`          // Yang dibayar pembeli
	OrderSellingPrice        float64                  `json:"order_selling_price"`         // Total harga jual pesanan
	SellingPrice             float64                  `json:"selling_price"`               // Total harga barang
	SellerOrderProcessingFee float64                  `json:"seller_order_processing_fee"` // Biaya pesanan tetap (Rp 1.000 / 1.250)
	CommissionFee            float64                  `json:"commission_fee"`              // Biaya admin komisi kategori
	ServiceFee               float64                  `json:"service_fee"`                 // Biaya Gratis Ongkir/Cashback XTRA
	SellerTransactionFee     float64                  `json:"seller_transaction_fee"`      // Biaya penanganan transaksi (~4%)
	VoucherFromSeller        float64                  `json:"voucher_from_seller"`         // Potongan voucher ditanggung toko
	SellerDiscount           float64                  `json:"seller_discount"`             // Diskon langsung dari toko
	NetCommissionFeeInfo     []ShopeeCommissionRule   `json:"net_commission_fee_info"`
	NetServiceFeeInfo        []ShopeeServiceRule      `json:"net_service_fee_info"`
}

type ShopeeCommissionRule struct {
	RuleID          int64   `json:"rule_id"`
	FeeAmount       float64 `json:"fee_amount"`
	RuleDisplayName string  `json:"rule_display_name"`
}

type ShopeeServiceRule struct {
	RuleID          int64   `json:"rule_id"`
	FeeAmount       float64 `json:"fee_amount"`
	RuleDisplayName string  `json:"rule_display_name"`
	Category        string  `json:"category"`
}

// GetEscrowDetail mengambil rincian penghasilan bersih seller & transparansi potongan biaya Shopee
func (c *Client) GetEscrowDetail(accessToken string, shopID uint64, orderSN string) (*ShopeeEscrowResponse, error) {
	path := "/api/v2/payment/get_escrow_detail"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)
	params.Set("order_sn", orderSN)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat HTTP request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons Shopee: %w", err)
	}

	var result ShopeeEscrowResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON escrow detail: %w, raw: %s", err, string(bodyBytes))
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s (req_id: %s)", result.Error, result.Message, result.RequestID)
	}

	return &result, nil
}
