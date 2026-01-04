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

func (c *BaserowClient) UpdateRow(data BaserowData, jsonStr string) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/database/rows/table/%s/%s/?user_field_names=true", c.BaseURL, data.GetTableID(), data.GetRowID())

	rawData := []byte(jsonStr)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(rawData))
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

func (c *BaserowClient) DeleteRow(data BaserowData) error {
	if !data.DeleteRowsAllowed() {
		return fmt.Errorf("deleting rows is not allowed for table ID %s", data.GetTableID())
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/database/rows/table/%s/%s/", c.BaseURL, data.GetTableID(), data.GetRowID())

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Token "+c.ApiKey)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API request failed with status: %s and error reading body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("API request failed with status: %s and body: %s", resp.Status, bodyString)
	}

	return nil
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
	ID   int    `json:"id"`
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

func (b BaserowItemTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowTagTable struct {
	ID      int    `json:"id"`
	TagName string `json:"Name"`
}

func (b *BaserowTagTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	var out BaserowQueryResponse[*BaserowTagTable]
	if err := json.NewDecoder(data).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return out, nil
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

func (b BaserowTagTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowPurchaseItemTable struct {
	ID          int    `json:"id"`
	Description string `json:"Description"`
}

func (b *BaserowPurchaseItemTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	var out BaserowQueryResponse[*BaserowPurchaseItemTable]
	if err := json.NewDecoder(data).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return out, nil
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

func (b BaserowPurchaseItemTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type LinkedItem struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Order string `json:"order"`
}

type BaserowPurchaseItemGroupTable struct {
	ID            int          `json:"id"`
	Name          string       `json:"Name"`
	PurchaseItems []LinkedItem `json:"Purchase Items"` // Assuming this is a link to multiple purchase items by their IDs
	Tags          []LinkedItem `json:"Tags"`           // Assuming this is a link to multiple tags by their IDs
}

func (b *BaserowPurchaseItemGroupTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	var out BaserowQueryResponse[*BaserowPurchaseItemGroupTable]
	if err := json.NewDecoder(data).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return out, nil
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

func (b BaserowPurchaseItemGroupTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowPurchaseItemGroupTableInsert struct {
	ID            int      `json:"id"`
	Name          string   `json:"Name"`
	PurchaseItems []string `json:"Purchase Items"`
	Tags          []string `json:"Tags"`
}

func (b *BaserowPurchaseItemGroupTableInsert) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	var out BaserowQueryResponse[*BaserowPurchaseItemGroupTableInsert]
	if err := json.NewDecoder(data).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return out, nil
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

func (b BaserowPurchaseItemGroupTableInsert) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowPurchaseEventTable struct {
	ID                int      `json:"id"`
	BankTxID          string   `json:"Bank Tx ID"`
	Date              string   `json:"Date"`
	Tax               float64  `json:"Tax"`
	Total             float64  `json:"Total"`
	TotalUnits        float64  `json:"Total Units"`
	TotalCases        float64  `json:"Total Cases"`
	Vendor            []string `json:"Vendor"`
	Note              string   `json:"Note"`
	PendingPurchaseID *int     `json:"PendingPurchase,omitempty"`
}

func (b *BaserowPurchaseEventTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	type raw struct {
		ID              int          `json:"id"`
		BankTxID        string       `json:"Bank Tx ID"`
		Date            string       `json:"Date"`
		Tax             *string      `json:"Tax"`
		Total           *string      `json:"Total"`
		TotalUnits      *string      `json:"Total Units"`
		TotalCases      *string      `json:"Total Cases"`
		Vendor          []LinkedItem `json:"Vendor"`
		Note            *string      `json:"Note"`
		PendingPurchase []LinkedItem `json:"PendingPurchase"`
	}

	var r BaserowQueryResponse[raw]
	if err := json.NewDecoder(data).Decode(&r); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	out := BaserowQueryResponse[*BaserowPurchaseEventTable]{
		Count:    r.Count,
		Next:     r.Next,
		Previous: r.Previous,
	}

	for _, r := range r.Results {
		var err error
		tax := 0.0
		if r.Tax != nil {
			tax, err = strconv.ParseFloat(*r.Tax, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Tax: %v", err)
			}
		}

		total := 0.0
		if r.Total != nil {
			total, err = strconv.ParseFloat(*r.Total, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total: %v", err)
			}
		}

		totalUnits := 0.0
		if r.TotalUnits != nil {
			totalUnits, err = strconv.ParseFloat(*r.TotalUnits, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total Units: %v", err)
			}
		}

		totalCases := 0.0
		if r.TotalCases != nil {
			totalCases, err = strconv.ParseFloat(*r.TotalCases, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total Cases: %v", err)
			}
		}

		var vendor []string
		for _, v := range r.Vendor {
			vendor = append(vendor, v.Value)
		}

		note := ""
		if r.Note != nil {
			note = *r.Note
		}

		var pendingPurchaseID *int
		if len(r.PendingPurchase) == 1 {
			pendingPurchaseID = &r.PendingPurchase[0].ID
		} else if len(r.PendingPurchase) > 1 {
			return nil, fmt.Errorf("unexpected number of linked pending purchases. Found %d, expected 0 or 1", len(r.PendingPurchase))
		}

		out.Results = append(out.Results, &BaserowPurchaseEventTable{
			ID:                r.ID,
			BankTxID:          r.BankTxID,
			Date:              r.Date,
			Note:              note,
			Vendor:            vendor,
			Tax:               tax,
			Total:             total,
			TotalUnits:        totalUnits,
			TotalCases:        totalCases,
			PendingPurchaseID: pendingPurchaseID,
		})
	}

	return out, nil
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

func (b BaserowPurchaseEventTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowPendingPurchase struct {
	ID              int     `json:"id"`
	BankTxID        string  `json:"Bank Tx ID"`
	ReceiptURL      string  `json:"Receipt URL"`
	Vendor          string  `json:"Vendor"`
	Date            *string `json:"Date"`
	ItemName        string  `json:"Item: Name"`
	ItemQuantity    int     `json:"Item: Quantity"`
	ItemIsCase      bool    `json:"Item: Is Case"`
	ItemPrice       float64 `json:"Item: Price"`
	Note            string  `json:"Note"`
	Tax             float64 `json:"Tax"`
	Total           float64 `json:"Total"`
	TotalUnits      int     `json:"Total Units"`
	TotalCases      int     `json:"Total Cases"`
	Reason          string  `json:"Reason"`
	BankTotal       float64 `json:"Bank Total"`
	PurchaseID      *int    `json:"PurchaseID"`
	PurchaseEventID *int    `json:"PurchaseEventID"`
}

func (b *BaserowPendingPurchase) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	type raw struct {
		ID            int          `json:"id"`
		BankTxID      string       `json:"Bank Tx ID"`
		Vendor        string       `json:"Vendor"`
		Date          *string      `json:"Date"`
		ItemName      *string      `json:"Item: Name"`
		ItemQuantity  *string      `json:"Item: Quantity"`
		ItemIsCase    bool         `json:"Item: Is Case"`
		ItemPrice     *string      `json:"Item: Price"`
		Note          *string      `json:"Note"`
		Tax           *string      `json:"Tax"`
		Total         *string      `json:"Total"`
		TotalUnits    *string      `json:"Total Units"`
		TotalCases    *string      `json:"Total Cases"`
		Reason        *string      `json:"Reason"`
		BankTotal     *string      `json:"Bank Total"`
		Purchase      []LinkedItem `json:"Purchase"`
		PurchaseEvent []LinkedItem `json:"PurchaseEvent"`
	}

	var r BaserowQueryResponse[raw]
	if err := json.NewDecoder(data).Decode(&r); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	out := BaserowQueryResponse[*BaserowPendingPurchase]{
		Count:    r.Count,
		Next:     r.Next,
		Previous: r.Previous,
	}

	for _, r := range r.Results {
		var err error
		tax := 0.0
		if r.Tax != nil {
			tax, err = strconv.ParseFloat(*r.Tax, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Tax: %v", err)
			}
		}

		total := 0.0
		if r.Total != nil {
			total, err = strconv.ParseFloat(*r.Total, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total: %v", err)
			}
		}

		totalUnits := 0
		if r.TotalUnits != nil {
			tu, err := strconv.ParseFloat(*r.TotalUnits, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total Units: %v", err)
			}
			totalUnits = int(tu)
		}

		totalCases := 0
		if r.TotalCases != nil {
			tc, err := strconv.ParseFloat(*r.TotalCases, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Total Cases: %v", err)
			}
			totalCases = int(tc)
		}

		bankTotal := 0.0
		if r.BankTotal != nil {
			bankTotal, err = strconv.ParseFloat(*r.BankTotal, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Bank Total: %v", err)
			}
		}

		note := ""
		if r.Note != nil {
			note = *r.Note
		}

		itemName := ""
		if r.ItemName != nil {
			itemName = *r.ItemName
		}

		itemQuantity := 0
		if r.ItemQuantity != nil {
			iq, err := strconv.ParseFloat(*r.ItemQuantity, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Item Quantity: %v", err)
			}
			itemQuantity = int(iq)
		}

		itemPrice := 0.0
		if r.ItemPrice != nil {
			itemPrice, err = strconv.ParseFloat(*r.ItemPrice, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing Item Price: %v", err)
			}
		}

		reason := ""
		if r.Reason != nil {
			reason = *r.Reason
		}

		var purchaseID *int
		if len(r.Purchase) == 1 {
			purchaseID = &r.Purchase[0].ID
		} else if len(r.Purchase) > 1 {
			return nil, fmt.Errorf("unexpected number of linked purchases. Found %d, expected 0 or 1", len(r.Purchase))
		}

		var purchaseEventID *int
		if len(r.PurchaseEvent) == 1 {
			purchaseEventID = &r.PurchaseEvent[0].ID
		} else if len(r.PurchaseEvent) > 1 {
			return nil, fmt.Errorf("unexpected number of linked purchase events. Found %d, expected 0 or 1", len(r.PurchaseEvent))
		}

		pp := &BaserowPendingPurchase{
			ID:              r.ID,
			BankTxID:        r.BankTxID,
			Date:            r.Date,
			Note:            note,
			Vendor:          r.Vendor,
			ItemName:        itemName,
			ItemQuantity:    itemQuantity,
			ItemIsCase:      r.ItemIsCase,
			ItemPrice:       itemPrice,
			Tax:             tax,
			Total:           total,
			TotalUnits:      totalUnits,
			TotalCases:      totalCases,
			Reason:          reason,
			BankTotal:       bankTotal,
			PurchaseID:      purchaseID,
			PurchaseEventID: purchaseEventID,
		}

		out.Results = append(out.Results, pp)
	}

	return out, nil
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

func (b BaserowPendingPurchase) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowVendorTable struct {
	ID       int    `json:"id"`
	Name     string `json:"Name"`
	Addrsess string `json:"Address"`
}

func (b *BaserowVendorTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	var out BaserowQueryResponse[*BaserowVendorTable]
	if err := json.NewDecoder(data).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return out, nil
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

func (b BaserowVendorTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

type BaserowData interface {
	GetTableID() string
	DeleteRowsAllowed() bool
	GetPrimaryKey() string
	GetRowID() string
	UnmarshalJSON(data io.ReadCloser) (interface{}, error) // interface{} BaserowQueryResponse[T]
}

type BaserowPurchaseTable struct {
	ID                 int      `json:"id"`
	Name               string   `json:"Name"`
	Quantity           int      `json:"Quantity"`
	IsCase             bool     `json:"Is Case"`
	Price              float64  `json:"Price"`
	PurchaseItem       []string `json:"PurchaseItem"`
	PurchaseEvent      []string `json:"PurchaseEvent"`
	PendingPurchaseIDs []int    `json:"PendingPurchases,omitempty"`
}

func (b *BaserowPurchaseTable) UnmarshalJSON(data io.ReadCloser) (interface{}, error) {
	// Alias struct with string fields
	type raw struct {
		ID               int          `json:"id"`
		Name             string       `json:"Name"`
		Quantity         string       `json:"Quantity"`
		IsCase           bool         `json:"Is Case"`
		Price            string       `json:"Price"`
		PurchaseItem     []LinkedItem `json:"PurchaseItem"`
		PurchaseEvent    []LinkedItem `json:"PurchaseEvent"`
		PendingPurchases []LinkedItem `json:"PendingPurchases"`
	}

	var r BaserowQueryResponse[raw]
	if err := json.NewDecoder(data).Decode(&r); err != nil {
		return BaserowQueryResponse[interface{}]{}, fmt.Errorf("error decoding JSON: %v", err)
	}

	out := BaserowQueryResponse[*BaserowPurchaseTable]{
		Count:    r.Count,
		Next:     r.Next,
		Previous: r.Previous,
	}

	for _, r := range r.Results {
		qty, err := strconv.Atoi(r.Quantity)
		if err != nil && r.Quantity != "" {
			return nil, fmt.Errorf("error parsing Quantity: %v", err)
		}

		price, err := strconv.ParseFloat(r.Price, 64)
		if err != nil && r.Price != "" {
			return nil, fmt.Errorf("error parsing Price: %v", err)
		}

		var purchaseItemIDs []string
		for _, pi := range r.PurchaseItem {
			purchaseItemIDs = append(purchaseItemIDs, pi.Value)
		}

		var purchaseEventIDs []string
		for _, pe := range r.PurchaseEvent {
			purchaseEventIDs = append(purchaseEventIDs, pe.Value)
		}

		var pendingPurchaseIDs []int
		for _, pp := range r.PendingPurchases {
			pendingPurchaseIDs = append(pendingPurchaseIDs, pp.ID)
		}

		out.Results = append(out.Results, &BaserowPurchaseTable{
			ID:                 r.ID,
			Name:               r.Name,
			Quantity:           qty,
			IsCase:             r.IsCase,
			Price:              price,
			PurchaseItem:       purchaseItemIDs,
			PurchaseEvent:      purchaseEventIDs,
			PendingPurchaseIDs: pendingPurchaseIDs,
		})
	}

	return out, nil
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

func (b BaserowPurchaseTable) GetRowID() string {
	return fmt.Sprintf("%d", b.ID)
}

func NewBaserowPurchaseTable(item ReceiptItem, purchaseItemID string, purchaseEventID string) *BaserowPurchaseTable {
	out := &BaserowPurchaseTable{
		Name:          cases.Title(language.English).String(item.Name),
		Quantity:      item.Quantity,
		IsCase:        item.IsCase,
		Price:         item.Price,
		PurchaseItem:  []string{purchaseItemID},
		PurchaseEvent: []string{purchaseEventID},
	}

	if item.PendingPurchase != nil {
		out.PendingPurchaseIDs = []int{item.PendingPurchase.ID}
	}

	return out
}

type BaserowQueryResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

func NewPurchaseEvent(req CreateBaserowPurchaseRequest) *BaserowPurchaseEventTable {
	var pendingPurchaseID *int
	if req.PendingPurchase != nil {
		pendingPurchaseID = &req.PendingPurchase.ID
	}

	return &BaserowPurchaseEventTable{
		BankTxID:          req.BankTransaction.ID,
		Date:              req.BankTransaction.CreatedAt,
		Tax:               req.ReceiptSummary.Tax,
		Total:             req.ReceiptSummary.Total,
		TotalUnits:        float64(req.ReceiptSummary.TotalUnits),
		TotalCases:        float64(req.ReceiptSummary.TotalCases),
		Vendor:            []string{req.ReceiptSummary.Vendor},
		Note:              req.BankTransaction.Note,
		PendingPurchaseID: pendingPurchaseID,
	}
}

func NewBaserowPurchaseItemTable(purchaseItemID string) BaserowPurchaseItemTable {
	return BaserowPurchaseItemTable{
		Description: purchaseItemID,
	}
}
