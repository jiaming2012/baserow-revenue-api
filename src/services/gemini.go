package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"google.golang.org/genai"

	"github.com/jiaming2012/receipt-bot/src/models"
)

func parseJSONBody(jsonStr string) ([]models.ReceiptItem, models.ReceiptSummary, error) {
	// Regex to extract JSON code blocks
	re := regexp.MustCompile("```json\\s*([\\s\\S]*?)\\s*```")
	matches := re.FindAllStringSubmatch(jsonStr, -1)

	if len(matches) != 1 {
		return nil, models.ReceiptSummary{}, fmt.Errorf("unexpected number of JSON code blocks found")
	}

	// Parse items array
	var result models.ReceiptJSON
	if err := json.Unmarshal([]byte(matches[0][1]), &result); err != nil {
		return nil, models.ReceiptSummary{}, fmt.Errorf("error unmarshaling receipt JSON: %v", err)
	}

	var items []models.ReceiptItem
	for _, it := range result.Items {
		item, err := it.ToReceiptItem()
		if err != nil {
			return nil, models.ReceiptSummary{}, fmt.Errorf("error converting items JSON to ReceiptItems: %v", err)
		}
		items = append(items, item)
	}

	if len(result.Items) == 0 {
		return nil, models.ReceiptSummary{}, fmt.Errorf("failed to parse items from JSON")
	}

	summary, err := result.Summary.ToReceiptSummary()
	if err != nil {
		return nil, models.ReceiptSummary{}, fmt.Errorf("error converting summary JSON to ReceiptSummary: %v", err)
	}

	if summary.Total <= 0.0 {
		return nil, models.ReceiptSummary{}, fmt.Errorf("failed to parse summary from JSON")
	}

	return items, summary, nil
}

func ParseReceipt(ctx context.Context, client *genai.Client, receiptURL string) ([]models.ReceiptItem, models.ReceiptSummary, error) {
	imageResp, err := http.Get(receiptURL)
	if err != nil {
		return nil, models.ReceiptSummary{}, fmt.Errorf("run_parse_receipt_with_genai: failed to fetch receipt image: %w", err)
	}
	defer imageResp.Body.Close()

	imageBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return nil, models.ReceiptSummary{}, fmt.Errorf("run_parse_receipt_with_genai: failed to read image bytes: %w", err)
	}

	parts := []*genai.Part{
		genai.NewPartFromBytes(imageBytes, "image/jpeg"),
		genai.NewPartFromText("Parse items[] with fields: name,quantity(int),price(float),total(float),is_case(bool)"),
		genai.NewPartFromText("Parse summary with fields: vendor,total_units(int),total_cases(int),tax(float),total(float)"),
		genai.NewPartFromText("Response in JSON format"),
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
		return nil, models.ReceiptSummary{}, fmt.Errorf("ParseReceipt: failed to generate content: %w", err)
	}

	jsonResp := result.Text()

	items, summary, err := parseJSONBody(jsonResp)
	if err != nil {
		return nil, models.ReceiptSummary{}, fmt.Errorf("ParseReceipt: failed to parse JSON body: %w", err)
	}

	totalItems := 0
	for _, item := range items {
		totalItems += item.Quantity
	}

	if summary.TotalUnits+summary.TotalCases != totalItems {
		_, _, err := parseJSONBody(jsonResp)
		if err != nil {
			return nil, models.ReceiptSummary{}, fmt.Errorf("ParseReceipt: failed to parse JSON body: %w", err)
		}
		return nil, models.ReceiptSummary{}, fmt.Errorf("ParseReceipt: mismatch between summary total units/cases and number of items parsed")
	}

	return items, summary, nil
}
