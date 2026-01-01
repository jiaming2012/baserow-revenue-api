package models

import "fmt"

const (
	BaserowItemTableID              = "786113"
	BaserowTagTableID               = "786134"
	BaserowPurchaseItemTableID      = "786129"
	BaserowPurchaseItemGroupTableID = "786135"
)

type MercuryPagination struct {
	NextPage     string `json:"nextPage"`
	PreviousPage string `json:"previousPage"`
}

type MercuryTransactionAttachment struct {
	AttachmentType string `json:"attachmentType"`
	FileName       string `json:"fileName"`
	URL            string `json:"url"`
}

type MercuryTransaction struct {
	ID              string                          `json:"id"`
	Amount          float64                         `json:"amount"`
	BankDescription string                          `json:"bankDescription"`
	Attachments     []*MercuryTransactionAttachment `json:"attachments"`
	CreatedAt       string                          `json:"createdAt"`
	Category        string                          `json:"mercuryCategory"`
	Note            string                          `json:"note"`
}

func (t MercuryTransaction) String() string {
	return fmt.Sprintf("(%s) %s - %s - %.2f", t.ID, t.CreatedAt, t.BankDescription, t.Amount)
}

type MercuryListAllTransactionsResponse struct {
	Page         MercuryPagination     `json:"page"`
	Transactions []*MercuryTransaction `json:"transactions"`
}
