package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jiaming2012/receipt-bot/src/models"
)

func FetchMercuryTransactions(ctx context.Context, apiKey string, startAt, endAt time.Time) (*models.MercuryListAllTransactionsResponse, error) {
	url := "https://api.mercury.com/api/v1/transactions"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("FetchMercuryTransactions: failed to create requrest with context: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Accept", "application/json;charset=utf-8")

	q := req.URL.Query()
	q.Add("start", startAt.Format("2006-01-02"))
	q.Add("end", endAt.Format("2006-01-02"))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FetchMercuryTransactions: failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FetchMercuryTransactions: received non-200 response code: %d", resp.StatusCode)
	}

	var transactionsResponse models.MercuryListAllTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactionsResponse); err != nil {
		return nil, fmt.Errorf("FetchMercuryTransactions: failed to decode json response: %w", err)
	}

	if len(transactionsResponse.Transactions) >= 1000 {
		return nil, fmt.Errorf("FetchMercuryTransactions: response limit reached - implement pagination")
	}

	return &transactionsResponse, nil
}

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
