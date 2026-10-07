package shopee

// ShopeeOrderListResponse merepresentasikan respon dari /api/v2/order/get_order_list
type ShopeeOrderListResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN     string `json:"order_sn"`
			OrderStatus string `json:"order_status"`
		} `json:"order_list"`
	} `json:"response"`
}

// ShopeeOrderDetailResponse merepresentasikan respon dari /api/v2/order/get_order_detail
type ShopeeOrderDetailResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		OrderList []ShopeeOrderDetailItem `json:"order_list"`
	} `json:"response"`
}

type ShopeeOrderDetailItem struct {
	OrderSN           string           `json:"order_sn"`
	OrderStatus       string           `json:"order_status"`
	BuyerUserID       uint64           `json:"buyer_user_id"`
	BuyerUsername     string           `json:"buyer_username"`
	MessageToSeller   string           `json:"message_to_seller"`
	ShipByDate              int64            `json:"ship_by_date"`
	ShippingCarrier         string           `json:"shipping_carrier"`
	CheckoutShippingCarrier string           `json:"checkout_shipping_carrier"`
	PackageList             []ShopeePackage  `json:"package_list"`
	TotalAmount             float64          `json:"total_amount"`
	BuyerCancelReason       string           `json:"buyer_cancel_reason"`
	CreateTime              int64            `json:"create_time"`
	UpdateTime              int64            `json:"update_time"`
	ItemList                []ShopeeItemInfo `json:"item_list"`
}

type ShopeePackage struct {
	PackageNumber   string `json:"package_number"`
	LogisticsStatus string `json:"logistics_status"`
	ShippingCarrier string `json:"shipping_carrier"`
}

type ShopeeItemInfo struct {
	ItemID                 uint64  `json:"item_id"`
	ItemName               string  `json:"item_name"`
	ItemSKU                string  `json:"item_sku"`
	ModelID                uint64  `json:"model_id"`
	ModelName              string  `json:"model_name"`
	ModelSKU               string  `json:"model_sku"`
	ModelQuantityPurchased int     `json:"model_quantity_purchased"`
	ModelOriginalPrice     float64 `json:"model_original_price"`
	ModelDiscountedPrice   float64 `json:"model_discounted_price"`
}
