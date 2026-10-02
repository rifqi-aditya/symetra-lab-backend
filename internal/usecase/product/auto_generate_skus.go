package product

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/product"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type AutoGenerateSKUsUseCase struct {
	productRepo product.Repository
}

func NewAutoGenerateSKUsUseCase(pRepo product.Repository) *AutoGenerateSKUsUseCase {
	return &AutoGenerateSKUsUseCase{productRepo: pRepo}
}

func (uc *AutoGenerateSKUsUseCase) Execute(ctx context.Context, userID uuid.UUID) (int, error) {
	products, err := uc.productRepo.FindAllWithoutSKU(ctx, userID)
	if err != nil {
		return 0, err
	}

	updatedCount := 0
	for _, p := range products {
		catCode := getCategoryCode(p.Category())
		nameCode := getProductNameCode(p.Name())
		fullSKU := fmt.Sprintf("%s-%s", catCode, nameCode)

		existing, _ := uc.productRepo.FindBySKU(ctx, userID, fullSKU)
		if existing != nil && existing.ID() != p.ID() {
			fullSKU = fmt.Sprintf("%s-%s", fullSKU, p.ID().String()[:4])
		}

		p.SetSKU(&fullSKU)
		if err := uc.productRepo.Update(ctx, p); err == nil {
			updatedCount++
		}
	}

	return updatedCount, nil
}

func getCategoryCode(cat string) string {
	c := strings.ToUpper(strings.TrimSpace(cat))
	switch {
	case strings.Contains(c, "BOX") || strings.Contains(c, "FUNCTIONAL"):
		return "BOX"
	case strings.Contains(c, "CLICKER") || strings.Contains(c, "FIDGET"):
		return "CLK"
	case strings.Contains(c, "KEYCHAIN"):
		return "KEY"
	case strings.Contains(c, "FLEXI") || strings.Contains(c, "ARTICULAT"):
		return "TOY"
	case strings.Contains(c, "DIORAMA") || strings.Contains(c, "MINIATURE"):
		return "DEC"
	case strings.Contains(c, "ORGANIZER") || strings.Contains(c, "SKADIS"):
		return "ORG"
	default:
		clean := cleanSKUPart(c)
		if len(clean) >= 3 {
			return clean[:3]
		}
		return "PRD"
	}
}

func getProductNameCode(name string) string {
	cleaned := cleanSKUPart(name)
	words := strings.Split(cleaned, "-")
	var codeParts []string
	for _, w := range words {
		if w == "DAN" || w == "DENGAN" || w == "FOR" || w == "THE" || w == "ITEM" {
			continue
		}
		if len(w) > 4 {
			codeParts = append(codeParts, w[:4])
		} else {
			codeParts = append(codeParts, w)
		}
		if len(strings.Join(codeParts, "")) >= 6 {
			break
		}
	}
	res := strings.Join(codeParts, "")
	if len(res) > 8 {
		return res[:8]
	}
	if res == "" {
		return "ITEM"
	}
	return res
}

func cleanSKUPart(input string) string {
	cleaned := strings.TrimSpace(input)
	cleaned = strings.ToUpper(cleaned)
	cleaned = nonAlphanumericRegex.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	return cleaned
}