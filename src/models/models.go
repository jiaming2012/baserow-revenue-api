package models

import "fmt"

type CreateBaserowPurchaseRequest struct {
	ReceiptSummary  ReceiptSummary      `json:"receipt_summary"`
	ReceiptItems    []ReceiptItem       `json:"receipt_items"`
	BankTransaction *MercuryTransaction `json:"bank_transaction"`
}

func NewBaserowPurchaseRequest(summary ReceiptSummary, items []ReceiptItem, tx *MercuryTransaction) (CreateBaserowPurchaseRequest, error) {
	if summary.Total != -tx.Amount {
		return CreateBaserowPurchaseRequest{}, fmt.Errorf("receipt total %.2f does not match transaction amount %.2f for tx ID %s", summary.Total, -tx.Amount, tx.ID)
	}

	return CreateBaserowPurchaseRequest{
		ReceiptSummary:  summary,
		ReceiptItems:    items,
		BankTransaction: tx,
	}, nil
}
