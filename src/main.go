package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
	"gopkg.in/yaml.v2"

	"github.com/jiaming2012/receipt-bot/src/models"
	"github.com/jiaming2012/receipt-bot/src/services"
)

const ReceiptURL = "https://mercury-technologies-user-uploads-prod.s3.us-east-1.amazonaws.com/0b13f45c-aafb-11ed-af69-d7d21b3593cf/attachment/f8ed1ce2-e38a-11f0-af0b-2b6ee0e5fa1b.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIA2IUVTP5NELRWABTI%2F20251228%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20251228T142859Z&X-Amz-Expires=43200&X-Amz-Security-Token=IQoJb3JpZ2luX2VjELv%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaCXVzLWVhc3QtMSJGMEQCIAi1KkDTKCUJ0T%2FS2fe4cNwMslpKStiac6hE4P0dnu4IAiBgt%2F%2BSOlnfCgR6xNO5NNLdnra39FKgxrMjUhMelp8VOSq5BQiD%2F%2F%2F%2F%2F%2F%2F%2F%2F%2F8BEAQaDDcwNTc2MjEzMTgwMiIM7vvuPDeUpOsGyXROKo0Ft%2BdFmoSfgjaDvYPdlal2oT2uAyB0dtUjxqHnzhvRbO2bS2qwI1VR7X%2BSitR37W1XCGhnxjZSeePm5XrkCu1tZLnPmeU9YjmfNc77ZYXeZPnRGf2lN5XJsscO75XRWz6ImHsTodMi7uUtEo2KdQr2w%2BeHOIslsld2fz3XWqGEZMH60PwwKXKcjjFv6qthxovJnmFGUpZmklSAC%2BoHs5OUtu9tkMKIj5%2BovPGpuLeqpaIXEUDi4my9fDfBLrCtupN0w0sxkiVosjMfgd563bxmui0AKmQ4fQSgq6xhwVrkoASCqd8UKY2zxVJ26GeeGSvA5Sy4lwMn2WTiHmSO5T8%2FB4R2hmzJrWTDmcmMDFUNUXLYJ%2Fz%2Fxt2MIZRT4jNVc1DLF21zlBHD8oyXeuondU4gPzeRpYOvHZZeaWF7jnqzbjD8fAqdbUQaQBON1byNa43yr833a9bXaQSw%2BBKMLjXB4Fq8HRrw5zImRSi%2BMDhqXaQ69g%2BUGpDHXUeA42YSJOi8EpA6CPUOOEbQgbH%2FVv08TPVzLfRfq6yd0MpbFjXO%2FyZjMSGp3%2BRc7Ld7NK9a6HR2%2Fkwq9aWVhzp3ei2Gv6sEJGg8PL55icy4iUqDE4Gr186%2FjK5n4uNpacp2S60w3xn1nbIpmu%2FF3UJ%2FeeDM1Cgabg0PltDWr05DUuoiCxUuLrs9mSVXNABmLm7EuLzKxIu1mvE%2ByNs6ZVN5HFrcPaHAaPVox5LD7ONYcS467EGee%2FSYaTgWOQKBzaUrUetqffLg%2FNSjCVNAIvWy6y3kJRGS5ify8zodQ1XcQWVF%2BeFlsC2manC9DG9T788WW1V4s6Z0yluJOx098dR0lH01nlWVHCNY4C9cPewzT7wkdQcw%2FoPEygY6sgEp%2FUqwRkcJ9vdXFRbfernpXPrEnEITx6dQUmyhVO39PHut%2FqfNS6AGtiAbWyt4cKE1eLjPn8llPH1ryEQADmwF%2Fz9rvDOe1w11AzboWhGHFakoI0wpkGF4VYtaTBaAYxCQYteW8j%2BQ0i7GXexxx5ytEA50eoSkNvfwYXNvqB1m9L6fZkgMgbPMZwD7%2BG8dJZ7IYBe9GdzDu6zSnpdn06FTVvuPXXzRcwzYXPKIiLkx%2FGX4&X-Amz-SignedHeaders=host&response-content-disposition=inline&response-content-type=image%2Fjpeg&X-Amz-Signature=e256213801abefab74345738f90d6d75887338d18f72fc12eb2d7c148061dfe6"

var IgnoreReceiptsNamed = []string{"Expense Reimbursement"}

func run_fetch_transactions(ctx context.Context, bankApiKey string, start, end time.Time) {
	resp, err := services.FetchMercuryTransactions(ctx, bankApiKey, start, end)
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

type PurchaseItemYAML struct {
	Description string   `yaml:"description"`
	Exclusions  []string `yaml:"exclusions"`
}

type PurchaseItemGroupYAML struct {
	Name          string              `yaml:"name"`
	Tags          []string            `yaml:"tags"`
	PurchaseItems []*PurchaseItemYAML `yaml:"purchase_items"`
}

func getMissingItems[T comparable](existingItems map[string]interface{}, itemsToCheck []T, getKey func(T) string) []T {
	var missingItems []T
	for _, item := range itemsToCheck {
		key := getKey(item)
		if _, exists := existingItems[key]; !exists {
			missingItems = append(missingItems, item)
		}
	}
	return missingItems
}

func run_apply_purchase_item_groups_fixtures(ctx context.Context, client *models.BaserowClient) error {
	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		return fmt.Errorf("PROJECT_DIR environment variable not set")
	}

	yamlFilePath := fmt.Sprintf("%s/fixtures/purchase_item_groups.yaml", projectDir)
	yamlBytes, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return fmt.Errorf("run_apply_purchase_item_groups_fixtures: failed to read YAML file: %w", err)
	}

	// Parse YAML content
	var groups []PurchaseItemGroupYAML
	if err := yaml.Unmarshal(yamlBytes, &groups); err != nil {
		return fmt.Errorf("Error parsing YAML file: %v", err)
	}

	// Fetch existing tags from Baserow
	existingTagsMap := make(map[string]models.BaserowTagTable)
	tagRows, err := services.ListRows[models.BaserowTagTable](client)
	for _, tagRow := range tagRows {
		existingTagsMap[tagRow.TagName] = tagRow
	}

	// Identify and add new tags
	newTagsToAdd := []models.BaserowTagTable{}
	for _, group := range groups {
		for _, tag := range group.Tags {
			if _, exists := existingTagsMap[tag]; !exists {
				newTagsToAdd = append(newTagsToAdd, models.BaserowTagTable{
					TagName: tag,
				})
				existingTagsMap[tag] = models.BaserowTagTable{
					// todo: execute to create and get ID
					TagName: tag,
				}
			}
		}
	}

	for _, tag := range newTagsToAdd {
		if err := client.CreateRow(tag); err != nil {
			return fmt.Errorf("Failed to create tag row: %w", err)
		}
	}

	// Fetch existing purchase items from Baserow
	existingPurchaseItemsMap := make(map[string]models.BaserowPurchaseItemTable)
	purchaseItemRows, err := services.ListRows[models.BaserowPurchaseItemTable](client)
	if err != nil {
		return fmt.Errorf("Failed to list purchase item rows: %w", err)
	}

	for _, itemRow := range purchaseItemRows {
		existingPurchaseItemsMap[itemRow.Description] = itemRow
	}

	// Identify and add new purchase items
	newPurchaseItemsToAdd := []models.BaserowPurchaseItemTable{}
	for _, group := range groups {
		for _, item := range group.PurchaseItems {
			// Remove % signs from description for matching
			description := strings.ReplaceAll(item.Description, "%", "")
			if description == "" {
				continue
			}

			if _, exists := existingPurchaseItemsMap[description]; !exists {
				newPurchaseItemsToAdd = append(newPurchaseItemsToAdd, models.BaserowPurchaseItemTable{
					Description: description,
				})
				existingPurchaseItemsMap[description] = models.BaserowPurchaseItemTable{
					Description: description,
				}
			}
		}
	}

	for _, item := range newPurchaseItemsToAdd {
		if err := client.CreateRow(item); err != nil {
			return fmt.Errorf("Failed to create purchase item row: %w", err)
		}
	}

	// Fetch existing purchase item groups from Baserow
	existingPurchaseItemGroupsMap := make(map[string]interface{})
	purchaseItemGroupRows, err := services.ListRows[models.BaserowPurchaseItemGroupTable](client)
	if err != nil {
		return fmt.Errorf("Failed to list purchase item group rows: %w", err)
	}

	for _, groupRow := range purchaseItemGroupRows {
		existingPurchaseItemGroupsMap[groupRow.Name] = groupRow
	}

	// Identify and add new purchase item groups
	newPurchaseItemGroupsToAdd := []models.BaserowPurchaseItemGroupTableInsert{}
	for _, group := range groups {
		if _, exists := existingPurchaseItemGroupsMap[group.Name]; !exists {
			item := models.BaserowPurchaseItemGroupTableInsert{
				Name:          group.Name,
				PurchaseItems: []string{},
				Tags:          []string{},
			}

			missingPurchaseItems := getMissingItems[*PurchaseItemYAML](existingPurchaseItemGroupsMap, group.PurchaseItems, func(pi *PurchaseItemYAML) string {
				return strings.ReplaceAll(pi.Description, "%", "")
			})

			for _, mpi := range missingPurchaseItems {
				key := strings.ReplaceAll(mpi.Description, "%", "")
				if key == "" {
					continue
				}
				_, found := existingPurchaseItemsMap[key]
				if !found {
					return fmt.Errorf("Purchase item not found for description: %s", key)
				}

				item.PurchaseItems = append(item.PurchaseItems, key)
			}

			missingTags := getMissingItems[string](existingPurchaseItemGroupsMap, group.Tags, func(tag string) string {
				return tag
			})

			for _, mt := range missingTags {
				if mt == "" {
					continue
				}
				_, found := existingTagsMap[mt]
				if !found {
					return fmt.Errorf("Tag not found for name: %s", mt)
				}

				item.Tags = append(item.Tags, mt)
			}

			newPurchaseItemGroupsToAdd = append(newPurchaseItemGroupsToAdd, item)
			existingPurchaseItemGroupsMap[group.Name] = item
		}
	}

	for _, group := range newPurchaseItemGroupsToAdd {
		if err := client.CreateRow(group); err != nil {
			return fmt.Errorf("Failed to create purchase item group row: %w", err)
		}
	}

	return nil
}

func main() {
	// ctx := context.Background()

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

	baserowApiKey := "hMX57EJDuf8rrB68BaLpY8BPxqFS4x87"
	baserowClient := models.NewBaserowClient("https://api.baserow.io", baserowApiKey)

	// newRow := BaserowItemTable{
	// 	Name: "Test Item from Go Client",
	// }

	// if err := baserowClient.CreateRow(newRow); err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to create Baserow row: %w", err))
	// }

	// rows, err := ListRows[BaserowItemTable](baserowClient)
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to list Baserow rows: %w", err))
	// }

	// for _, row := range rows {
	// 	fmt.Printf("Row: %+v\n", row)
	// }

	// fmt.Println("Successfully created a new row in Baserow!")

	if err := run_apply_purchase_item_groups_fixtures(context.Background(), baserowClient); err != nil {
		log.Fatal(fmt.Errorf("Failed to apply purchase item groups fixtures: %w", err))
	}
}
