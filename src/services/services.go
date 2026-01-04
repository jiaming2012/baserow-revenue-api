package services

import (
	"fmt"
	"math"

	"github.com/jiaming2012/receipt-bot/src/models"
)

func ValidateReceiptData(items []models.ReceiptItem, summary models.ReceiptSummary, mercuryTx *models.MercuryTransaction) error {
	// Validate summary total against transaction amount
	if summary.Total != -mercuryTx.Amount {
		return fmt.Errorf("receipt total %.2f does not match transaction amount %.2f for tx ID %s", summary.Total, -mercuryTx.Amount, mercuryTx.ID)
	}

	// Validate items total against summary total minus tax
	itemsTotal := 0.0
	for _, item := range items {
		itemsTotal += item.Price * float64(item.Quantity)
	}

	summaryTotal := summary.Total - summary.Tax
	if math.Abs(summaryTotal-itemsTotal) > 0.01 {
		return fmt.Errorf("ParseReceipt: mismatch between summary total (%.2f) and sum of item totals (%.2f)", summaryTotal, itemsTotal)
	}

	// Validate total units and cases
	totalItems := 0
	for _, item := range items {
		totalItems += item.Quantity
	}

	if summary.TotalUnits+summary.TotalCases != totalItems {
		return fmt.Errorf("ParseReceipt: mismatch between summary total units/cases (%d) and number of items parsed (%d)", summary.TotalUnits+summary.TotalCases, totalItems)
	}

	return nil
}

func GroupPendingPurchasesByBankTxID(pendingPurchases []*models.BaserowPendingPurchase) (map[string][]*models.BaserowPendingPurchase, error) {
	grouped := make(map[string][]*models.BaserowPendingPurchase)
	for _, pp := range pendingPurchases {
		if pp.BankTxID == "" {
			continue
		}

		if _, exists := grouped[pp.BankTxID]; !exists {
			grouped[pp.BankTxID] = []*models.BaserowPendingPurchase{}
		}

		grouped[pp.BankTxID] = append(grouped[pp.BankTxID], pp)
	}

	return grouped, nil
}

func GroupPurchasesByBankTxID(pendingPurchases []*models.BaserowPurchaseTable, purchaseEventsMap map[string]*models.BaserowPurchaseEventTable) (map[string][]*models.BaserowPurchaseTable, error) {
	grouped := make(map[string][]*models.BaserowPurchaseTable)
	for _, pp := range pendingPurchases {
		if len(pp.PurchaseEvent) != 1 {
			return nil, fmt.Errorf("GroupPurchasesByBankTxID: expected exactly one linked purchase event for purchase ID %d, got %d", pp.ID, len(pp.PurchaseEvent))
		}

		purchaseEvent, found := purchaseEventsMap[pp.PurchaseEvent[0]]
		if !found {
			return nil, fmt.Errorf("GroupPurchasesByBankTxID: purchase event ID %s not found for purchase ID %d", pp.PurchaseEvent[0], pp.ID)
		}

		if purchaseEvent.BankTxID == "" {
			return nil, fmt.Errorf("GroupPurchasesByBankTxID: empty BankTxID for purchase event ID %s", purchaseEvent.ID)
		}

		if _, exists := grouped[purchaseEvent.BankTxID]; !exists {
			grouped[purchaseEvent.BankTxID] = []*models.BaserowPurchaseTable{}
		}

		grouped[purchaseEvent.BankTxID] = append(grouped[purchaseEvent.BankTxID], pp)
	}

	return grouped, nil
}
