package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jiaming2012/receipt-bot/src/models"
)

var IgnoreReceiptsNamed = []string{"Expense Reimbursement"}

func FetchReceipts(ctx context.Context, bankApiKey string, start, end time.Time) (validTransactions []*models.MercuryTransaction, invalidTransactions []*models.MercuryTransaction, err error) {
	resp, err := fetchMercuryTransactions(ctx, bankApiKey, start, end)
	if err != nil {
		err = fmt.Errorf("FetchReceipts: failed to fetch mercury transactions: %w", err)
		return
	}

outer_loop:
	for _, tx := range resp.Transactions {
		for _, ignoreName := range IgnoreReceiptsNamed {
			if tx.BankDescription == ignoreName {
				continue outer_loop
			}
		}

		if len(tx.Attachments) > 0 || tx.Note != "" {
			validTransactions = append(validTransactions, tx)
		} else {
			invalidTransactions = append(invalidTransactions, tx)
		}
	}

	if len(invalidTransactions) > 0 {
		err = fmt.Errorf("FetchReceipts: found %d transactions without attachments or notes", len(invalidTransactions))
	}

	return
}

func fetchMercuryTransactions(ctx context.Context, apiKey string, startAt, endAt time.Time) (*models.MercuryListAllTransactionsResponse, error) {
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
