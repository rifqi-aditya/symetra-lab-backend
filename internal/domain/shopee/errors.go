package shopee

import "errors"

var (
	ErrShopNotFound      = errors.New("shopee shop not found")
	ErrOrderNotFound     = errors.New("shopee order not found")
	ErrPartnerNotConfig  = errors.New("shopee partner credentials not configured")
	ErrInvalidCallback   = errors.New("invalid shopee callback parameters")
)
