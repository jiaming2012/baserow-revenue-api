package models

import "fmt"

type CreateBaserowPurchaseRequest struct {
	ReceiptSummary  ReceiptSummary          `json:"receipt_summary"`
	ReceiptItems    []ReceiptItem           `json:"receipt_items"`
	BankTransaction *MercuryTransaction     `json:"bank_transaction"`
	PendingPurchase *BaserowPendingPurchase `json:"pending_purchase"`
}

func NewCreateBaserowPurchaseRequest(summary ReceiptSummary, items []ReceiptItem, tx *MercuryTransaction, pendingPurchase *BaserowPendingPurchase) (CreateBaserowPurchaseRequest, error) {
	return CreateBaserowPurchaseRequest{
		ReceiptSummary:  summary,
		ReceiptItems:    items,
		BankTransaction: tx,
		PendingPurchase: pendingPurchase,
	}, nil
}

func NewCreateBaserowPurchaseRequestFromPendingPurchases(pendingPurchases []*BaserowPendingPurchase) (CreateBaserowPurchaseRequest, error) {
	if len(pendingPurchases) < 2 {
		return CreateBaserowPurchaseRequest{}, fmt.Errorf("not enough pending purchases to create CreateBaserowPurchaseRequest. expected header + at least 1 item, got %d", len(pendingPurchases))
	}

	header := pendingPurchases[0]

	if len(header.Vendor) == 0 {
		return CreateBaserowPurchaseRequest{}, fmt.Errorf("pending purchase missing vendor for bank tx ID %s", header.BankTxID)
	}

	summary := ReceiptSummary{
		Vendor:     header.Vendor,
		Total:      header.Total,
		Tax:        header.Tax,
		TotalUnits: header.TotalUnits,
		TotalCases: header.TotalCases,
	}

	if header.Date == nil {
		return CreateBaserowPurchaseRequest{}, fmt.Errorf("pending purchase missing date for bank tx ID %s", header.BankTxID)
	}

	tx := &MercuryTransaction{
		ID:        header.BankTxID,
		Amount:    header.BankTotal,
		CreatedAt: *header.Date,
		Note:      header.Note,
		Attachments: []*MercuryTransactionAttachment{
			{
				URL: header.ReceiptURL,
			},
		},
	}

	var items []ReceiptItem
	for _, item := range pendingPurchases[1:] {
		items = append(items, ReceiptItem{
			Name:            item.ItemName,
			Quantity:        item.ItemQuantity,
			Price:           item.ItemPrice,
			IsCase:          item.ItemIsCase,
			PendingPurchase: item,
		})
	}

	return NewCreateBaserowPurchaseRequest(summary, items, tx, header)
}

func NewBaserowPendingPurchases(summary ReceiptSummary, items []ReceiptItem, tx *MercuryTransaction, err error) ([]*BaserowPendingPurchase, error) {
	receiptURL := ""
	if len(tx.Attachments) == 1 {
		receiptURL = tx.Attachments[0].URL
	} else if len(tx.Attachments) > 1 {
		return nil, fmt.Errorf("NewBaserowPendingPurchases: expected 0 or 1 attachment for tx ID %s, got %d", tx.ID, len(tx.Attachments))
	}

	reason := err.Error()

	out := []*BaserowPendingPurchase{
		// Header
		{
			BankTxID:   tx.ID,
			BankTotal:  tx.Amount,
			Vendor:     summary.Vendor,
			Date:       &tx.CreatedAt,
			Note:       tx.Note,
			Tax:        summary.Tax,
			Total:      summary.Total,
			TotalUnits: summary.TotalUnits,
			TotalCases: summary.TotalCases,
			ReceiptURL: receiptURL,
			Reason:     reason,
		},
	}

	// Items
	for _, item := range items {
		pendingPurchase := &BaserowPendingPurchase{
			BankTxID:     tx.ID,
			ItemName:     item.Name,
			ItemQuantity: item.Quantity,
			ItemPrice:    item.Price,
			ItemIsCase:   item.IsCase,
			Total:        item.Price * float64(item.Quantity),
		}
		out = append(out, pendingPurchase)
	}

	return out, nil
}
