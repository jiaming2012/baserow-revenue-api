package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jiaming2012/receipt-bot/src/models"
)

func listRows[T models.BaserowData](url string, c *models.BaserowClient) (models.BaserowQueryResponse[T], error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.BaserowQueryResponse[T]{}, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Token "+c.ApiKey)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return models.BaserowQueryResponse[T]{}, fmt.Errorf("Error making API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return models.BaserowQueryResponse[T]{}, fmt.Errorf("API request failed with status: %s and error reading body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)

		return models.BaserowQueryResponse[T]{}, fmt.Errorf("API request failed with status: %s and body: %s", resp.Status, bodyString)
	}

	var response models.BaserowQueryResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.BaserowQueryResponse[T]{}, fmt.Errorf("Error decoding response: %v", err)
	}

	return response, nil
}

func ListRows[T models.BaserowData](c *models.BaserowClient) ([]T, error) {
	var instance T
	url := fmt.Sprintf("%s/api/database/rows/table/%s/?user_field_names=true", c.BaseURL, instance.GetTableID())

	var results []T
	var count int
	for {
		resp, err := listRows[T](url, c)
		if err != nil {
			return nil, fmt.Errorf("Error listing rows: %v, url: %s", err, url)
		}

		results = append(results, resp.Results...)
		count = resp.Count

		if resp.Next == "" {
			break
		} else {
			url = resp.Next
		}
	}

	if len(results) != count {
		return nil, fmt.Errorf("Mismatch in expected count and results length. Expected: %d, Got: %d", count, len(results))
	}

	return results, nil
}

func CreatePurchaseEventWithItems(ctx context.Context, client *models.BaserowClient, req models.CreateBaserowPurchaseRequest) error {
	purchaseEvent := &models.BaserowPurchaseEventTable{
		BankTxID:  req.BankTransaction.ID,
		Date:      req.BankTransaction.CreatedAt,
		Tax:       req.ReceiptSummary.Tax,
		Total:     req.ReceiptSummary.Total,
		CardLast4: "0000",
	}

	if err := client.CreateRow(purchaseEvent); err != nil {
		return fmt.Errorf("CreatePurchaseEventWithItems: failed to create purchase event row: %w", err)
	}

	return nil
}
