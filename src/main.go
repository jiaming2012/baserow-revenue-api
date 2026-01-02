package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/genai"
	"gopkg.in/yaml.v2"

	"github.com/jiaming2012/receipt-bot/src/models"
	"github.com/jiaming2012/receipt-bot/src/services"
)

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
	ctx := context.Background()

	AiApiKey := os.Getenv("AI_API_KEY")
	if AiApiKey == "" {
		log.Fatal("AI_API_KEY environment variable is not set")
	}

	bankApiKey := os.Getenv("BANK_API_KEY")
	if bankApiKey == "" {
		log.Fatal("BANK_API_KEY environment variable is not set")
	}

	aiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: AiApiKey,
	})

	if err != nil {
		log.Fatal(fmt.Errorf("Failed to initiate AI Client"))
	}

	// if err := demo(ctx, client); err != nil {
	// 	log.Fatal(err)
	// }

	end := time.Now()
	start := end.AddDate(0, 0, -7) // 1 week ago

	// if err := run_parse_receipt_with_genai(ctx, client); err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to parse receipt: %w", err))
	// }

	baserowApiKey := os.Getenv("BASEROW_API_KEY")
	if baserowApiKey == "" {
		log.Fatal("BASEROW_API_KEY environment variable is not set")
	}

	baserowClient := models.NewBaserowClient("https://api.baserow.io", baserowApiKey)

	// fetch existing purchase events
	purchaseEventsDTOs, err := services.ListRows[models.BaserowPurchaseEventTableDTO](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list purchase events: %w", err))
	}

	var purchaseEvents []models.BaserowPurchaseEventTable
	for _, dto := range purchaseEventsDTOs {
		ev, err := dto.ToBaserowPurchaseEventTable()
		if err != nil {
			log.Fatal(fmt.Errorf("Failed to convert purchase event DTO: %w", err))
		}
		purchaseEvents = append(purchaseEvents, ev)
	}

	existingPurchaseEventsMap := make(map[string]interface{})
	for _, pe := range purchaseEvents {
		existingPurchaseEventsMap[pe.BankTxID] = pe
	}

	var newPurchaseRequests []models.CreateBaserowPurchaseRequest
	validTx, invalidTx, err := services.FetchReceipts(context.Background(), bankApiKey, start, end)
	if len(validTx) > 0 {
		missingPurchaseEvents := getMissingItems(existingPurchaseEventsMap, validTx, func(tx *models.MercuryTransaction) string {
			return tx.ID
		})

		for _, ev := range missingPurchaseEvents {
			if len(ev.Attachments) > 0 {
				for _, att := range ev.Attachments {
					items, summary, err := services.ParseReceipt(ctx, aiClient, att.URL)
					if err != nil {
						log.Fatalf("Failed to parse receipt: %v, from %+v", err, ev)
					}

					mercuryTx := ev

					req, err := models.NewBaserowPurchaseRequest(summary, items, mercuryTx)
					if err != nil {
						log.Fatalf("Failed to create Baserow purchase request: %v", err)
					}

					newPurchaseRequests = append(newPurchaseRequests, req)
				}
			}

			// temp: for testing
			if len(newPurchaseRequests) > 0 {
				break
			}
		}
	}

	// fetch existing vendors
	exisitingVendors, err := services.ListRows[models.BaserowVendorTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list existing vendors: %w", err))
	}

	existingVendorsMap := make(map[string]models.BaserowVendorTable)
	for _, v := range exisitingVendors {
		existingVendorsMap[v.Name] = v
	}

	// fetch existing purchase items
	existingPurchaseItems, err := services.ListRows[models.BaserowPurchaseItemTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list existing purchase items: %w", err))
	}

	existingPurchaseItemMap := make(map[string]models.BaserowPurchaseItemTable)
	for _, pi := range existingPurchaseItems {
		existingPurchaseItemMap[pi.Description] = pi
	}

	// fetch existing purchases
	// existingPurchases, err := services.ListRows[models.BaserowPurchaseEventTable](baserowClient)
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to list existing purchases: %w", err))
	// }

	// existingPurchasesMap := make(map[string]models.BaserowPurchaseEventTable)
	// for _, p := range existingPurchases {
	// 	existingPurchasesMap[p.BankTxID] = p
	// }

	// process new purchase requests
	for _, pr := range newPurchaseRequests {
		vendorPk, isNew := services.DerivePurchaseItem(pr.ReceiptSummary.Vendor, existingVendorsMap)
		if isNew {
			newVendor := &models.BaserowVendorTable{
				Name: vendorPk,
			}

			if err := baserowClient.CreateRow(newVendor); err != nil {
				log.Fatal(fmt.Errorf("Failed to create new vendor: %w", err))
			}

			// update vendor map
			existingVendorsMap[vendorPk] = *newVendor
		}

		// update receipt summary vendor to use primary key
		pr.ReceiptSummary.Vendor = vendorPk

		// process purchase event
		var purchaseEventID string
		if _, exists := existingPurchaseEventsMap[pr.BankTransaction.ID]; !exists {
			purchaseEvent := models.NewPurchaseEvent(pr)

			if err := baserowClient.CreateRow(purchaseEvent); err != nil {
				log.Fatal(fmt.Errorf("Failed to create purchase event: %w", err))
			}

			// update purchase events map
			existingPurchaseEventsMap[pr.BankTransaction.ID] = *purchaseEvent

			purchaseEventID = purchaseEvent.BankTxID
		} else {
			purchaseEventID = existingPurchaseEventsMap[pr.BankTransaction.ID].(models.BaserowPurchaseEventTable).BankTxID
		}

		for _, item := range pr.ReceiptItems {
			purchaseItemID, isNew := services.DerivePurchaseItem(item.Name, existingPurchaseItemMap)
			if isNew {
				purchaseItem := models.NewBaserowPurchaseItemTable(purchaseItemID)

				if err := baserowClient.CreateRow(purchaseItem); err != nil {
					log.Fatal(fmt.Errorf("Failed to create purchase item: %w", err))
				}

				// update purchase item map
				existingPurchaseItemMap[purchaseItem.Description] = purchaseItem
			}

			purchase := models.NewBaserowPurchaseTable(item, purchaseItemID, purchaseEventID)

			if err := baserowClient.CreateRow(purchase); err != nil {
				log.Fatal(fmt.Errorf("Failed to create purchase: %w", err))
			}
		}
	}

	if err != nil {
		if len(invalidTx) > 0 {
			log.Errorf("Next invalid transactions: %+v", invalidTx)
		}

		log.Fatal(fmt.Errorf("Failed to fetch receipts: %w", err))
	}

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

	// if err := run_apply_purchase_item_groups_fixtures(context.Background(), baserowClient); err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to apply purchase item groups fixtures: %w", err))
	// }
}
