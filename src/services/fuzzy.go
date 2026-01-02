package services

import (
	"strings"

	"github.com/antzucaro/matchr"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/jiaming2012/receipt-bot/src/models"
)

func DerivePurchaseItem[T models.BaserowData](purchaseName string, existing map[string]T) (purchaseItem string, isNew bool) {
	purchaseNameLower := strings.ToLower(purchaseName)

	// exact match
	if item, exists := existing[purchaseNameLower]; exists {
		return item.GetPrimaryKey(), false
	}

	// fuzzy match
	var bestMatch string
	var highestScore float64 = 0.0
	for name := range existing {
		nameLower := strings.ToLower(name)

		score := matchr.JaroWinkler(purchaseNameLower, nameLower, true)
		if score > highestScore {
			highestScore = score
			bestMatch = name
		}
	}

	if highestScore >= 0.85 {
		return existing[bestMatch].GetPrimaryKey(), false
	}

	return cases.Title(language.English).String(purchaseName), true
}
