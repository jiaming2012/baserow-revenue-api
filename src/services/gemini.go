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

	// Decode the JSON content
	var result models.ReceiptJSON

	if len(matches) == 1 {
		if err := json.Unmarshal([]byte(matches[0][1]), &result); err != nil {
			return nil, models.ReceiptSummary{}, fmt.Errorf("error unmarshaling receipt JSON: %v", err)
		}
	} else if len(matches) == 2 {
		bFoundItems, bFoundSummary := false, false
		for _, match := range matches {
			if !bFoundItems {
				if err := json.Unmarshal([]byte(match[1]), &result.Items); err == nil {
					bFoundItems = true
					continue
				}
			}

			if !bFoundSummary {
				if err := json.Unmarshal([]byte(match[1]), &result.Summary); err == nil {
					bFoundSummary = true
					continue
				}
			}
		}

		if !bFoundItems {
			return nil, models.ReceiptSummary{}, fmt.Errorf("failed to parse items from JSON")
		}

		if !bFoundSummary {
			return nil, models.ReceiptSummary{}, fmt.Errorf("failed to parse summary from JSON")
		}
	} else {
		return nil, models.ReceiptSummary{}, fmt.Errorf("unexpected number of JSON code blocks found")
	}

	// Parse items array
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

	// Parse summary
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
		// "gemini-2.5-pro",
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

	return items, summary, nil
}

// func ManuallyParseReceipt(oldItems []models.ReceiptItem, oldSummary models.ReceiptSummary) ([]models.ReceiptItem, models.ReceiptSummary, error) {
// 	var newItems []models.ReceiptItem
// 	for i, old := range oldItems {
// 		newItem := models.ReceiptItem{}

// 		fmt.Printf("item[%d].Name = %s (y/n)?", i+1, old.Name)
// 		var resp string
// 		fmt.Scanln(&resp)
// 		if strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes" {
// 			newItem.Name = old.Name
// 		} else {
// 			fmt.Printf("Enter correct name: ")
// 			fmt.Scanln(&newItem.Name)
// 		}

// 		fmt.Printf("item[%d].Price = %.2f (y/n)?", i+1, old.Price)
// 		fmt.Scanln(&resp)
// 		if strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes" {
// 			newItem.Price = old.Price
// 		} else {
// 			fmt.Printf("Enter correct price: ")
// 			fmt.Scanln(&newItem.Price)
// 		}

// 		fmt.Printf("item[%d].Quantity = %d (y/n)?", i+1, old.Quantity)
// 		fmt.Scanln(&resp)
// 		if strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes" {
// 			newItem.Quantity = old.Quantity
// 		} else {
// 			fmt.Printf("Enter correct quantity: ")
// 			fmt.Scanln(&newItem.Quantity)
// 		}

// 		fmt.Printf("item[%d].IsCase = %t (y/n)?", i+1, old.IsCase)
// 		fmt.Scanln(&resp)
// 		if strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes" {
// 			newItem.IsCase = old.IsCase
// 		} else {
// 			fmt.Printf("Enter correct IsCase (true/false): ")
// 			fmt.Scanln(&newItem.IsCase)
// 	}

// 	newItems = append(newItems, newItem)
