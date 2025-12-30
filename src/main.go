package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/genai"
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

const ReceiptURL = "https://mercury-technologies-user-uploads-prod.s3.us-east-1.amazonaws.com/0b13f45c-aafb-11ed-af69-d7d21b3593cf/attachment/f8ed1ce2-e38a-11f0-af0b-2b6ee0e5fa1b.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIA2IUVTP5NELRWABTI%2F20251228%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20251228T142859Z&X-Amz-Expires=43200&X-Amz-Security-Token=IQoJb3JpZ2luX2VjELv%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaCXVzLWVhc3QtMSJGMEQCIAi1KkDTKCUJ0T%2FS2fe4cNwMslpKStiac6hE4P0dnu4IAiBgt%2F%2BSOlnfCgR6xNO5NNLdnra39FKgxrMjUhMelp8VOSq5BQiD%2F%2F%2F%2F%2F%2F%2F%2F%2F%2F8BEAQaDDcwNTc2MjEzMTgwMiIM7vvuPDeUpOsGyXROKo0Ft%2BdFmoSfgjaDvYPdlal2oT2uAyB0dtUjxqHnzhvRbO2bS2qwI1VR7X%2BSitR37W1XCGhnxjZSeePm5XrkCu1tZLnPmeU9YjmfNc77ZYXeZPnRGf2lN5XJsscO75XRWz6ImHsTodMi7uUtEo2KdQr2w%2BeHOIslsld2fz3XWqGEZMH60PwwKXKcjjFv6qthxovJnmFGUpZmklSAC%2BoHs5OUtu9tkMKIj5%2BovPGpuLeqpaIXEUDi4my9fDfBLrCtupN0w0sxkiVosjMfgd563bxmui0AKmQ4fQSgq6xhwVrkoASCqd8UKY2zxVJ26GeeGSvA5Sy4lwMn2WTiHmSO5T8%2FB4R2hmzJrWTDmcmMDFUNUXLYJ%2Fz%2Fxt2MIZRT4jNVc1DLF21zlBHD8oyXeuondU4gPzeRpYOvHZZeaWF7jnqzbjD8fAqdbUQaQBON1byNa43yr833a9bXaQSw%2BBKMLjXB4Fq8HRrw5zImRSi%2BMDhqXaQ69g%2BUGpDHXUeA42YSJOi8EpA6CPUOOEbQgbH%2FVv08TPVzLfRfq6yd0MpbFjXO%2FyZjMSGp3%2BRc7Ld7NK9a6HR2%2Fkwq9aWVhzp3ei2Gv6sEJGg8PL55icy4iUqDE4Gr186%2FjK5n4uNpacp2S60w3xn1nbIpmu%2FF3UJ%2FeeDM1Cgabg0PltDWr05DUuoiCxUuLrs9mSVXNABmLm7EuLzKxIu1mvE%2ByNs6ZVN5HFrcPaHAaPVox5LD7ONYcS467EGee%2FSYaTgWOQKBzaUrUetqffLg%2FNSjCVNAIvWy6y3kJRGS5ify8zodQ1XcQWVF%2BeFlsC2manC9DG9T788WW1V4s6Z0yluJOx098dR0lH01nlWVHCNY4C9cPewzT7wkdQcw%2FoPEygY6sgEp%2FUqwRkcJ9vdXFRbfernpXPrEnEITx6dQUmyhVO39PHut%2FqfNS6AGtiAbWyt4cKE1eLjPn8llPH1ryEQADmwF%2Fz9rvDOe1w11AzboWhGHFakoI0wpkGF4VYtaTBaAYxCQYteW8j%2BQ0i7GXexxx5ytEA50eoSkNvfwYXNvqB1m9L6fZkgMgbPMZwD7%2BG8dJZ7IYBe9GdzDu6zSnpdn06FTVvuPXXzRcwzYXPKIiLkx%2FGX4&X-Amz-SignedHeaders=host&response-content-disposition=inline&response-content-type=image%2Fjpeg&X-Amz-Signature=e256213801abefab74345738f90d6d75887338d18f72fc12eb2d7c148061dfe6"

var IgnoreReceiptsNamed = []string{"Expense Reimbursement"}

func FetchMercuryTransactions(ctx context.Context, apiKey string, startAt, endAt time.Time) (*MercuryListAllTransactionsResponse, error) {
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

	var transactionsResponse MercuryListAllTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactionsResponse); err != nil {
		return nil, fmt.Errorf("FetchMercuryTransactions: failed to decode json response: %w", err)
	}

	if len(transactionsResponse.Transactions) >= 1000 {
		return nil, fmt.Errorf("FetchMercuryTransactions: response limit reached - implement pagination")
	}

	return &transactionsResponse, nil
}

func run_fetch_transactions(ctx context.Context, bankApiKey string, start, end time.Time) {
	resp, err := FetchMercuryTransactions(ctx, bankApiKey, start, end)
	if err != nil {
		log.Fatal(err)
	}

	for _, tx := range resp.Transactions {
		if len(tx.Attachments) > 0 {
			for _, ignoreName := range IgnoreReceiptsNamed {
				if tx.BankDescription == ignoreName {
					continue
				}
			}
			
			fmt.Printf("Transaction ID: %s has %d attachments\n", tx.ID, len(tx.Attachments))
			fmt.Printf("Desc: %s\n", tx.BankDescription)
			fmt.Printf("Amount: %.2f\n", tx.Amount)
			for _, att := range tx.Attachments {
				fmt.Printf(" - Attachment: %s (%s) URL: %s\n", att.FileName, att.AttachmentType, att.URL)
			}
			fmt.Println("-----\n")
		}
	}

	fmt.Printf("Fetched %d transactions\n", len(resp.Transactions))
}

func run_parse_receipt_with_genai(ctx context.Context, client *genai.Client) error {
	imageResp, err := http.Get(ReceiptURL)
	if err != nil {
		return fmt.Errorf("run_parse_receipt_with_genai: failed to fetch receipt image: %w", err)
	}
	defer imageResp.Body.Close()

	imageBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return fmt.Errorf("run_parse_receipt_with_genai: failed to read image bytes: %w", err)
	}

	parts := []*genai.Part{
		genai.NewPartFromBytes(imageBytes, "image/jpeg"),
		genai.NewPartFromText("Parse receipt with columns: Name, Quantity, Price, Total"),
		genai.NewPartFromText("Parse subtotal, tax, and total amounts."),
		genai.NewPartFromText("Provide the response in CSV format."),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		contents,
		nil,
	)

	if err != nil {
		return fmt.Errorf("run_parse_receipt_with_genai: failed to generate content: %w", err)
	}

	fmt.Println(result.Text())

	return nil
}

func main() {
	ctx := context.Background()

	// AiApiKey := os.Getenv("AI_API_KEY")
	// if AiApiKey == "" {
	// 	log.Fatal("AI_API_KEY environment variable is not set")
	// }

	// bankApiKey := os.Getenv("BANK_API_KEY")
	// if bankApiKey == "" {
	// 	log.Fatal("BANK_API_KEY environment variable is not set")
	// }

	// client, err := genai.NewClient(ctx, &genai.ClientConfig{
	// 	APIKey: AiApiKey,
	// })

	// if err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to initiate AI Client"))
	// }

	// if err := demo(ctx, client); err != nil {
	// 	log.Fatal(err)
	// }

	// end := time.Now()
	// start := end.AddDate(0, 0, -7) // 1 week ago

	// if err := run_parse_receipt_with_genai(ctx, client); err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to parse receipt: %w", err))
	// }
}
