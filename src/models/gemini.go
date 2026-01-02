package models

import (
	"fmt"
	"strconv"
	"strings"
)

type Receipt struct {
	Items   []ReceiptItem  `json:"items"`
	Summary ReceiptSummary `json:"summary"`
}

type ReceiptJSON struct {
	Items   []ReceiptItemsJSON `json:"items"`
	Summary ReceiptSummaryJSON `json:"summary"`
}

type ReceiptItemsJSON struct {
	Name     string      `json:"Name"`
	IsCase   bool        `json:"IsCase"`
	Quantity interface{} `json:"Quantity"`
	Price    interface{} `json:"Price"`
}

func (r ReceiptItemsJSON) ToReceiptItem() (ReceiptItem, error) {
	receiptItem := ReceiptItem{
		Name:   r.Name,
		IsCase: r.IsCase,
	}

	var quantity int
	switch qty := r.Quantity.(type) {
	case float64:
		quantity = int(qty)
	case string:
		qtyInt, err := strconv.Atoi(qty)
		if err != nil {
			return ReceiptItem{}, fmt.Errorf("error parsing quantity: %v", err)
		}
		quantity = qtyInt
	default:
		return ReceiptItem{}, fmt.Errorf("unexpected type for quantity")
	}

	// Default quantity to 1 if not provided or invalid
	if quantity <= 0 {
		quantity = 1
	}

	var price float64
	switch prc := r.Price.(type) {
	case float64:
		price = prc
	case string:
		priceStr := strings.ReplaceAll(prc, "$", "")
		priceFloat, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return ReceiptItem{}, fmt.Errorf("error parsing price: %v", err)
		}
		price = priceFloat
	default:
		return ReceiptItem{}, fmt.Errorf("unexpected type for price")
	}

	receiptItem.Quantity = quantity
	receiptItem.Price = price

	return receiptItem, nil
}

type ReceiptItem struct {
	Quantity int     `json:"Quantity"`
	Price    float64 `json:"Price"`
	IsCase   bool    `json:"IsCase"`
	Name     string  `json:"Name"`
}

type ReceiptSummaryJSON struct {
	Vendor     string      `json:"vendor"`
	Tax        interface{} `json:"tax"`
	Total      interface{} `json:"total"`
	TotalUnits interface{} `json:"total_units"`
	TotalCases interface{} `json:"total_cases"`
}

type ReceiptSummary struct {
	Vendor     string  `json:"vendor"`
	Tax        float64 `json:"Tax"`
	Total      float64 `json:"Total"`
	TotalUnits int     `json:"total_units"`
	TotalCases int     `json:"total_cases"`
}

func (r ReceiptSummaryJSON) ToReceiptSummary() (ReceiptSummary, error) {
	receiptSummary := ReceiptSummary{
		Vendor: r.Vendor,
	}

	switch tax := r.Tax.(type) {
	case float64:
		receiptSummary.Tax = tax
	case string:
		taxStr := strings.ReplaceAll(tax, "$", "")
		taxFloat, err := strconv.ParseFloat(taxStr, 64)
		if err != nil {
			return ReceiptSummary{}, fmt.Errorf("error parsing tax: %v", err)
		}
		receiptSummary.Tax = taxFloat
	default:
		return ReceiptSummary{}, fmt.Errorf("unexpected type for tax")
	}

	switch total := r.Total.(type) {
	case float64:
		receiptSummary.Total = total
	case string:
		totalStr := strings.ReplaceAll(total, "$", "")
		totalFloat, err := strconv.ParseFloat(totalStr, 64)
		if err != nil {
			return ReceiptSummary{}, fmt.Errorf("error parsing total: %v", err)
		}
		receiptSummary.Total = totalFloat
	default:
		return ReceiptSummary{}, fmt.Errorf("unexpected type for total")
	}

	switch totalUnits := r.TotalUnits.(type) {
	case int:
		receiptSummary.TotalUnits = totalUnits
	case float64:
		receiptSummary.TotalUnits = int(totalUnits)
	case string:
		totalUnitsInt, err := strconv.Atoi(totalUnits)
		if err != nil {
			return ReceiptSummary{}, fmt.Errorf("error parsing total units: %v", err)
		}
		receiptSummary.TotalUnits = totalUnitsInt
	default:
		return ReceiptSummary{}, fmt.Errorf("unexpected type for total units")
	}

	switch totalCases := r.TotalCases.(type) {
	case int:
		receiptSummary.TotalCases = totalCases
	case float64:
		receiptSummary.TotalCases = int(totalCases)
	case string:
		totalCasesInt, err := strconv.Atoi(totalCases)
		if err != nil {
			return ReceiptSummary{}, fmt.Errorf("error parsing total cases: %v", err)
		}
		receiptSummary.TotalCases = totalCasesInt
	default:
		return ReceiptSummary{}, fmt.Errorf("unexpected type for total cases")
	}

	return receiptSummary, nil
}
