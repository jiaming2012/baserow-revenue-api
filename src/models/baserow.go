package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	BaserowItemTableID              = "786113"
	BaserowTagTableID               = "786134"
	BaserowPurchaseItemTableID      = "786129"
	BaserowPurchaseItemGroupTableID = "786135"
	BaserowVendorTableID            = "786130"
	BaserowPurchaseEventTableID     = "786138"
	BaserowPendingPurchasesTableID  = "788804"
	BaserowPurchaseTableID          = "786116"
)

type BaserowClient struct {
	ApiKey  string
	BaseURL string
}

func (c *BaserowClient) CreateRow(data BaserowData) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/database/rows/table/%s/?user_field_names=true", c.BaseURL, data.GetTableID())

	rawData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error marshalling data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawData))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Token "+c.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API request failed with status: %s and error reading body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("API request failed with status: %s and body: %s", resp.Status, bodyString)
	}

	return nil
}

func NewBaserowClient(baseURL, apiKey string) *BaserowClient {
	return &BaserowClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
	}
}

type BaserowItemTable struct {
	Name string `json:"Name"`
}

func (b BaserowItemTable) GetTableID() string {
	return BaserowItemTableID
}

func (b BaserowItemTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowItemTable) GetPrimaryKey() string {
	return b.Name
}

type BaserowTagTable struct {
	ID      int    `json:"id"`
	TagName string `json:"Name"`
}

func (b BaserowTagTable) GetTableID() string {
	return BaserowTagTableID
}

func (b BaserowTagTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowTagTable) GetPrimaryKey() string {
	return b.TagName
}

type BaserowPurchaseItemTable struct {
	ID          int    `json:"id"`
	Description string `json:"Description"`
}

func (b BaserowPurchaseItemTable) GetTableID() string {
	return BaserowPurchaseItemTableID
}

func (b BaserowPurchaseItemTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseItemTable) GetPrimaryKey() string {
	return b.Description
}

type LinkedItem struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Order string `json:"order"`
}

type BaserowPurchaseItemGroupTable struct {
	Name          string       `json:"Name"`
	PurchaseItems []LinkedItem `json:"Purchase Items"` // Assuming this is a link to multiple purchase items by their IDs
	Tags          []LinkedItem `json:"Tags"`           // Assuming this is a link to multiple tags by their IDs
}

func (b BaserowPurchaseItemGroupTable) GetTableID() string {
	return BaserowPurchaseItemGroupTableID
}

func (b BaserowPurchaseItemGroupTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseItemGroupTable) GetPrimaryKey() string {
	return b.Name
}

type BaserowPurchaseItemGroupTableInsert struct {
	Name          string   `json:"Name"`
	PurchaseItems []string `json:"Purchase Items"`
	Tags          []string `json:"Tags"`
}

func (b BaserowPurchaseItemGroupTableInsert) GetTableID() string {
	return BaserowPurchaseItemGroupTableID
}

func (b BaserowPurchaseItemGroupTableInsert) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseItemGroupTableInsert) GetPrimaryKey() string {
	return b.Name
}

type BaserowPurchaseEventTableDTO struct {
	ID         int          `json:"id"`
	BankTxID   string       `json:"Bank Tx ID"`
	Date       string       `json:"Date"`
	Tax        *string      `json:"Tax"`
	Total      *string      `json:"Total"`
	TotalUnits *string      `json:"Total Units"`
	TotalCases *string      `json:"Total Cases"`
	Vendor     []LinkedItem `json:"Vendor"`
	Note       *string      `json:"Note"`
}

func (b BaserowPurchaseEventTableDTO) GetTableID() string {
	return BaserowPurchaseEventTableID
}

func (b BaserowPurchaseEventTableDTO) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseEventTableDTO) GetPrimaryKey() string {
	return b.BankTxID
}

func (b BaserowPurchaseEventTableDTO) ToBaserowPurchaseEventTable() (BaserowPurchaseEventTable, error) {
	var err error
	tax := 0.0
	if b.Tax != nil {
		tax, err = strconv.ParseFloat(*b.Tax, 64)
		if err != nil {
			return BaserowPurchaseEventTable{}, fmt.Errorf("error parsing Tax: %v", err)
		}
	}

	total := 0.0
	if b.Total != nil {
		total, err = strconv.ParseFloat(*b.Total, 64)
		if err != nil {
			return BaserowPurchaseEventTable{}, fmt.Errorf("error parsing Total: %v", err)
		}
	}

	totalUnits := 0.0
	if b.TotalUnits != nil {
		totalUnits, err = strconv.ParseFloat(*b.TotalUnits, 64)
		if err != nil {
			return BaserowPurchaseEventTable{}, fmt.Errorf("error parsing Total Units: %v", err)
		}
	}

	totalCases := 0.0
	if b.TotalCases != nil {
		totalCases, err = strconv.ParseFloat(*b.TotalCases, 64)
		if err != nil {
			return BaserowPurchaseEventTable{}, fmt.Errorf("error parsing Total Cases: %v", err)
		}
	}

	var vendor []string
	for _, v := range b.Vendor {
		vendor = append(vendor, v.Value)
	}

	note := ""
	if b.Note != nil {
		note = *b.Note
	}

	return BaserowPurchaseEventTable{
		ID:         b.ID,
		BankTxID:   b.BankTxID,
		Date:       b.Date,
		Note:       note,
		Vendor:     vendor,
		Tax:        tax,
		Total:      total,
		TotalUnits: totalUnits,
		TotalCases: totalCases,
	}, nil
}

type BaserowPurchaseEventTable struct {
	ID         int      `json:"id"`
	BankTxID   string   `json:"Bank Tx ID"`
	Date       string   `json:"Date"`
	Tax        float64  `json:"Tax"`
	Total      float64  `json:"Total"`
	TotalUnits float64  `json:"Total Units"`
	TotalCases float64  `json:"Total Cases"`
	Vendor     []string `json:"Vendor"`
	Note       string   `json:"Note"`
}

func (b BaserowPurchaseEventTable) GetTableID() string {
	return BaserowPurchaseEventTableID
}

func (b BaserowPurchaseEventTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseEventTable) GetPrimaryKey() string {
	return b.BankTxID
}

type BaserowPendingPurchase struct {
	BankTxID          string   `json:"Bank Tx ID"`
	Vendor            []string `json:"Vendor"`
	Date              string   `json:"Date"`
	ItemQuantity      int      `json:"Item: Quantity"`
	ItemIsCase        bool     `json:"Item: Is Case"`
	ItemPurchaseItem  []string `json:"Item: Purchase Item"`
	ItemPurchaseEvent []string `json:"Item: Purchase Event"`
	ItemPrice         float64  `json:"Item: Price"`
	Note              string   `json:"Note"`
	Tax               float64  `json:"Tax"`
	Total             float64  `json:"Total"`
	TotalUnits        float64  `json:"Total Units"`
	TotalCases        float64  `json:"Total Cases"`
}

func (b BaserowPendingPurchase) GetTableID() string {
	return BaserowPendingPurchasesTableID
}

func (b BaserowPendingPurchase) DeleteRowsAllowed() bool {
	return true
}

func (b BaserowPendingPurchase) GetPrimaryKey() string {
	return b.BankTxID
}

func NewBaserowPendingPurchases(req CreateBaserowPurchaseRequest) []BaserowPendingPurchase {
	// Header
	out := []BaserowPendingPurchase{
		{
			BankTxID:   req.BankTransaction.ID,
			Vendor:     []string{req.ReceiptSummary.Vendor},
			Date:       req.BankTransaction.CreatedAt,
			Note:       req.BankTransaction.Note,
			Tax:        req.ReceiptSummary.Tax,
			Total:      req.ReceiptSummary.Total,
			TotalUnits: float64(req.ReceiptSummary.TotalUnits),
			TotalCases: float64(req.ReceiptSummary.TotalCases),
		},
	}

	// Items
	for _, item := range req.ReceiptItems {
		out = append(out, BaserowPendingPurchase{
			ItemQuantity:      item.Quantity,
			ItemIsCase:        item.IsCase,
			ItemPurchaseItem:  []string{item.Name},
			ItemPurchaseEvent: []string{req.BankTransaction.ID},
			ItemPrice:         item.Price,
		})
	}

	return out
}

type BaserowData interface {
	GetTableID() string
	DeleteRowsAllowed() bool
	GetPrimaryKey() string
}

type BaserowVendorTable struct {
	Name     string `json:"Name"`
	Addrsess string `json:"Address"`
}

func (b BaserowVendorTable) GetTableID() string {
	return BaserowVendorTableID
}

func (b BaserowVendorTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowVendorTable) GetPrimaryKey() string {
	return b.Name
}

type BaserowPurchaseTable struct {
	Name          string   `json:"Name"`
	Quantity      int      `json:"Quantity"`
	IsCase        bool     `json:"Is Case"`
	Price         float64  `json:"Price"`
	PurchaseItem  []string `json:"PurchaseItem"`
	PurchaseEvent []string `json:"PurchaseEvent"`
}

func (b BaserowPurchaseTable) GetTableID() string {
	return BaserowPurchaseTableID
}

func (b BaserowPurchaseTable) DeleteRowsAllowed() bool {
	return false
}

func (b BaserowPurchaseTable) GetPrimaryKey() string {
	panic("GetPrimaryKey not implemented for BaserowPurchaseTable")
}

func NewBaserowPurchaseTable(item ReceiptItem, purchaseItemID string, purchaseEventID string) BaserowPurchaseTable {
	return BaserowPurchaseTable{
		Name:          cases.Title(language.English).String(item.Name),
		Quantity:      item.Quantity,
		IsCase:        item.IsCase,
		Price:         item.Price,
		PurchaseItem:  []string{purchaseItemID},
		PurchaseEvent: []string{purchaseEventID},
	}
}

type BaserowQueryResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

func NewPurchaseEvent(req CreateBaserowPurchaseRequest) *BaserowPurchaseEventTable {
	return &BaserowPurchaseEventTable{
		BankTxID:   req.BankTransaction.ID,
		Date:       req.BankTransaction.CreatedAt,
		Tax:        req.ReceiptSummary.Tax,
		Total:      req.ReceiptSummary.Total,
		TotalUnits: float64(req.ReceiptSummary.TotalUnits),
		TotalCases: float64(req.ReceiptSummary.TotalCases),
		Vendor:     []string{req.ReceiptSummary.Vendor},
		Note:       req.BankTransaction.Note,
	}
}

func NewBaserowPurchaseItemTable(purchaseItemID string) BaserowPurchaseItemTable {
	return BaserowPurchaseItemTable{
		Description: purchaseItemID,
	}
}

// func NewPurchaseItems(req CreateBaserowPurchaseRequest) []BaserowPurchaseItemTable {
// 	var items []BaserowPurchaseItemTable
// 	for _, ri := range req.ReceiptItems {
// 		items = append(items, BaserowPurchaseItemTable{
// 			Description: ri.Name,
// 		})
// 	}
// 	return items
// }
