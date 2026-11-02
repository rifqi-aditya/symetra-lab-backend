package marketplacecred

import "errors"

var (
	ErrCredentialNotFound = errors.New("marketplace credential not found")
	ErrPlatformRequired   = errors.New("platform name is required")
	ErrPartnerIDRequired  = errors.New("partner ID / app key is required")
	ErrPartnerKeyRequired = errors.New("partner key / app secret is required")
)

