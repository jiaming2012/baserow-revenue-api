package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
}

type MercuryListAllTransactionsResponse struct {
	Page         MercuryPagination     `json:"page"`
	Transactions []*MercuryTransaction `json:"transactions"`
}

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

type BaserowTagTable struct {
	ID      int    `json:"id"`
	TagName string `json:"Name"`
}

func (b BaserowTagTable) GetTableID() string {
	return BaserowTagTableID
}

type BaserowPurchaseItemTable struct {
	ID          int    `json:"id"`
	Description string `json:"Description"`
}

func (b BaserowPurchaseItemTable) GetTableID() string {
	return BaserowPurchaseItemTableID
}

type NestedItem struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Order string `json:"order"`
}

type BaserowPurchaseItemGroupTable struct {
	Name          string       `json:"Name"`
	PurchaseItems []NestedItem `json:"Purchase Items"` // Assuming this is a link to multiple purchase items by their IDs
	Tags          []NestedItem `json:"Tags"`           // Assuming this is a link to multiple tags by their IDs
}

func (b BaserowPurchaseItemGroupTable) GetTableID() string {
	return BaserowPurchaseItemGroupTableID
}

type BaserowPurchaseItemGroupTableInsert struct {
	Name          string   `json:"Name"`
	PurchaseItems []string `json:"Purchase Items"`
	Tags          []string `json:"Tags"`
}

func (b BaserowPurchaseItemGroupTableInsert) GetTableID() string {
	return BaserowPurchaseItemGroupTableID
}

type BaserowData interface {
	GetTableID() string
}

type BaserowQueryResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}
