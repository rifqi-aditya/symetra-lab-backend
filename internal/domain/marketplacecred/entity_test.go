package marketplacecred_test

import (
	"testing"

	"symetra-lab-backend-v2/internal/domain/marketplacecred"
)

func TestMarketplaceCredential_Entity(t *testing.T) {
	t.Run("should create valid credential and mask partner key", func(t *testing.T) {
		cred, err := marketplacecred.NewMarketplaceCredential(
			marketplacecred.PlatformShopee,
			"2044497",
			"shpk676b516461596a455257534a557973446969466f44455268686f6d727756",
			"https://api.symetra.com/callback",
			true,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cred.Platform() != marketplacecred.PlatformShopee {
			t.Fatalf("expected platform SHOPEE, got %s", cred.Platform())
		}
		if cred.PartnerID() != "2044497" {
			t.Fatalf("expected partner ID 2044497, got %s", cred.PartnerID())
		}
		if !cred.IsActive() {
			t.Fatalf("new credential should be active")
		}

		masked := cred.MaskedPartnerKey()
		if !maskedMatches(masked, "shpk...7756") {
			t.Fatalf("expected masked key format shpk...7756, got: %s", masked)
		}

		cred.SetPushKey("4d68734e79635058547376725a6b505765555279544867476d57456f54774455")
		if cred.PushKey() != "4d68734e79635058547376725a6b505765555279544867476d57456f54774455" {
			t.Fatalf("expected push key to be set, got: %s", cred.PushKey())
		}
		if !maskedMatches(cred.MaskedPushKey(), "4d68...4455") {
			t.Fatalf("expected masked push key 4d68...4455, got: %s", cred.MaskedPushKey())
		}
	})

	t.Run("should validate required fields", func(t *testing.T) {
		_, err := marketplacecred.NewMarketplaceCredential("", "123", "secret", "", true)
		if err != marketplacecred.ErrPlatformRequired {
			t.Fatalf("expected ErrPlatformRequired, got %v", err)
		}

		_, err = marketplacecred.NewMarketplaceCredential("SHOPEE", "", "secret", "", true)
		if err != marketplacecred.ErrPartnerIDRequired {
			t.Fatalf("expected ErrPartnerIDRequired, got %v", err)
		}

		_, err = marketplacecred.NewMarketplaceCredential("SHOPEE", "123", "", "", true)
		if err != marketplacecred.ErrPartnerKeyRequired {
			t.Fatalf("expected ErrPartnerKeyRequired, got %v", err)
		}
	})
}

func maskedMatches(masked, expected string) bool {
	return masked == expected
}
