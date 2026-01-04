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
	existingTagsMap := make(map[string]*models.BaserowTagTable)
	tagRows, err := models.ListRows[*models.BaserowTagTable](client)
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
				existingTagsMap[tag] = &models.BaserowTagTable{
					TagName: tag,
				}
			}
		}
	}

	for _, tag := range newTagsToAdd {
		if err := client.CreateRow(&tag); err != nil {
			return fmt.Errorf("Failed to create tag row: %w", err)
		}
	}

	// Fetch existing purchase items from Baserow
	existingPurchaseItemsMap := make(map[string]*models.BaserowPurchaseItemTable)
	purchaseItemRows, err := models.ListRows[*models.BaserowPurchaseItemTable](client)
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
				existingPurchaseItemsMap[description] = &models.BaserowPurchaseItemTable{
					Description: description,
				}
			}
		}
	}

	for _, item := range newPurchaseItemsToAdd {
		if err := client.CreateRow(&item); err != nil {
			return fmt.Errorf("Failed to create purchase item row: %w", err)
		}
	}

	// Fetch existing purchase item groups from Baserow
	existingPurchaseItemGroupsMap := make(map[string]interface{})
	purchaseItemGroupRows, err := models.ListRows[*models.BaserowPurchaseItemGroupTable](client)
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
		if err := client.CreateRow(&group); err != nil {
			return fmt.Errorf("Failed to create purchase item group row: %w", err)
		}
	}

	return nil
}

func DeriveProcessedPendingPurchase(pendingPurchases []*models.BaserowPendingPurchase, purchases []*models.BaserowPurchaseTable, purchaseEvents map[string]*models.BaserowPurchaseEventTable) ([]*models.BaserowPendingPurchase, error) {
	groupedPendingPurchases, err := services.GroupPendingPurchasesByBankTxID(pendingPurchases)
	if err != nil {
		return nil, fmt.Errorf("DeriveProcessedPendingPurchase: failed to group pending purchases: %w", err)
	}

	var out []*models.BaserowPendingPurchase
	for _, ppGroup := range groupedPendingPurchases {
		for _, pp := range ppGroup {
			if pp.PurchaseID != nil || pp.PurchaseEventID != nil {
				out = append(out, pp)
			}
		}
	}

	return out, nil
}

func remove_processed_pending_purchases(ctx context.Context, baserowClient *models.BaserowClient) error {
	pendingPurchasesToDelete, err := models.ListRows[*models.BaserowPendingPurchase](baserowClient)
	if err != nil {
		return fmt.Errorf("Failed to list pending purchases for deletion: %w", err)
	}

	currentPurchaseEvents, err := models.ListRows[*models.BaserowPurchaseEventTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list purchase events: %w", err))
	}

	currentPurchaseEventsMap := make(map[string]*models.BaserowPurchaseEventTable)
	for _, pe := range currentPurchaseEvents {
		currentPurchaseEventsMap[pe.BankTxID] = pe
	}

	currentPurchases, err := models.ListRows[*models.BaserowPurchaseTable](baserowClient)
	if err != nil {
		return fmt.Errorf("Failed to list current purchases for deletion: %w", err)
	}

	processedPendingPurchases, err := DeriveProcessedPendingPurchase(pendingPurchasesToDelete, currentPurchases, currentPurchaseEventsMap)
	if err != nil {
		return fmt.Errorf("Failed to derive processed pending purchases for deletion: %w", err)
	}

	for i := len(processedPendingPurchases) - 1; i >= 0; i-- {
		pp := processedPendingPurchases[i]
		if err := baserowClient.DeleteRow(pp); err != nil {
			return fmt.Errorf("Failed to delete processed pending purchase ID %d: %w", pp.ID, err)
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
	start := end.AddDate(0, 0, -14) // 1 week ago

	// if err := run_parse_receipt_with_genai(ctx, client); err != nil {
	// 	log.Fatal(fmt.Errorf("Failed to parse receipt: %w", err))
	// }

	baserowApiKey := os.Getenv("BASEROW_API_KEY")
	if baserowApiKey == "" {
		log.Fatal("BASEROW_API_KEY environment variable is not set")
	}

	baserowClient := models.NewBaserowClient("https://api.baserow.io", baserowApiKey)

	// fetch existing purchase events
	purchaseEvents, err := models.ListRows[*models.BaserowPurchaseEventTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list purchase events: %w", err))
	}

	existingPurchaseEventsMap := make(map[string]interface{})
	parsedReceiptsMap := make(map[string]interface{})
	pendingPurchaseIDToPurchaseEventMap := make(map[int]*models.BaserowPurchaseEventTable)
	for _, pe := range purchaseEvents {
		existingPurchaseEventsMap[pe.BankTxID] = pe
		parsedReceiptsMap[pe.BankTxID] = true
		if pe.PendingPurchaseID != nil {
			pendingPurchaseIDToPurchaseEventMap[*pe.PendingPurchaseID] = pe
		}
	}

	// fetch existing purchases
	purchases, err := models.ListRows[*models.BaserowPurchaseTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list purchases: %w", err))
	}

	pendingPurchaseIDToPurchaseMap := make(map[int]*models.BaserowPurchaseTable)
	for _, p := range purchases {
		for _, ppID := range p.PendingPurchaseIDs {
			pendingPurchaseIDToPurchaseMap[ppID] = p
		}
	}

	// add pending purchase events to parsed receipts
	pendingPurchases, err := models.ListRows[*models.BaserowPendingPurchase](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list pending purchases: %w", err))
	}

	// remove processed pending purchases from pendingPurchases
	for i := len(pendingPurchases) - 1; i >= 0; i-- {
		pp := pendingPurchases[i]
		if _, exists := pendingPurchaseIDToPurchaseMap[pp.ID]; exists {
			pendingPurchases = append(pendingPurchases[:i], pendingPurchases[i+1:]...)
		} else if _, exists := pendingPurchaseIDToPurchaseEventMap[pp.ID]; exists {
			pendingPurchases = append(pendingPurchases[:i], pendingPurchases[i+1:]...)
		}
	}

	groupedPendingPurchases, err := services.GroupPendingPurchasesByBankTxID(pendingPurchases)
	if err != nil {
		log.Fatalf("Failed to group pending purchases by bank tx ID: %v", err)
	}

	for bankTxID := range groupedPendingPurchases {
		parsedReceiptsMap[bankTxID] = true
	}

	var newPurchaseRequests []models.CreateBaserowPurchaseRequest

	// validate and add pending purchase requests
	for _, pp := range groupedPendingPurchases {
		purchaseReq, err := models.NewCreateBaserowPurchaseRequestFromPendingPurchases(pp)
		if err != nil {
			log.Fatalf("Invalid pending purchase for bank tx ID %s: %v", pp[0].BankTxID, err)
		}

		if err := services.ValidateReceiptData(purchaseReq.ReceiptItems, purchaseReq.ReceiptSummary, purchaseReq.BankTransaction); err != nil {
			for _, item := range pp {
				if len(item.Reason) > 0 {
					if !strings.Contains(item.Reason, err.Error()) {
						if err := baserowClient.UpdateRow(item, fmt.Sprintf(`{"Reason": "%s"}`, err.Error())); err != nil {
							log.Fatalf("Failed to update pending purchase reason: %v", err)
						}
					}
				}
			}
			log.Errorf("Invalid receipt data in pending purchase for bank tx ID %s: %v", pp[0].BankTxID, err)
			continue
		}

		newPurchaseRequests = append(newPurchaseRequests, purchaseReq)
	}

	// fetch new receipts from bank
	validTx, invalidTx, err := services.FetchReceipts(context.Background(), bankApiKey, start, end)
	if len(validTx) > 0 {
		missingPurchaseEvents := getMissingItems(parsedReceiptsMap, validTx, func(tx *models.MercuryTransaction) string {
			return tx.ID
		})

		for _, mercuryTx := range missingPurchaseEvents {
			if len(mercuryTx.Attachments) > 1 {
				log.Fatalf("Expected 0 or 1 attachment for tx ID %s, got %d", mercuryTx.ID, len(mercuryTx.Attachments))
			} else if len(mercuryTx.Attachments) == 1 {
				attachment := mercuryTx.Attachments[0]

				items, summary, err := services.ParseReceipt(ctx, aiClient, attachment.URL)
				if err != nil {
					log.Fatalf("Failed to parse receipt: %v, from %+v", err, mercuryTx)
				}

				if err := services.ValidateReceiptData(items, summary, mercuryTx); err != nil {
					log.Errorf("Invalid receipt data: %v, from %+v", err, mercuryTx)

					log.Info("Storing invalid transaction for review in PendingPurchases table")
					pendingPurchases, err := models.NewBaserowPendingPurchases(summary, items, mercuryTx, err)
					if err != nil {
						log.Fatalf("Failed to create Baserow pending purchases: %v", err)
					}

					for _, pp := range pendingPurchases {
						if err := baserowClient.CreateRow(pp); err != nil {
							log.Fatalf("Failed to create pending purchase row: %v", err)
						}
					}

					continue
				}

				req, err := models.NewCreateBaserowPurchaseRequest(summary, items, mercuryTx, nil)
				if err != nil {
					log.Fatalf("Failed to create Baserow purchase request: %v", err)
				}

				newPurchaseRequests = append(newPurchaseRequests, req)
			}
		}
	}

	// fetch existing vendors
	exisitingVendors, err := models.ListRows[*models.BaserowVendorTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list existing vendors: %w", err))
	}

	existingVendorsMap := make(map[string]*models.BaserowVendorTable)
	for _, v := range exisitingVendors {
		existingVendorsMap[v.Name] = v
	}

	// fetch existing purchase items
	existingPurchaseItems, err := models.ListRows[*models.BaserowPurchaseItemTable](baserowClient)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to list existing purchase items: %w", err))
	}

	existingPurchaseItemMap := make(map[string]*models.BaserowPurchaseItemTable)
	for _, pi := range existingPurchaseItems {
		existingPurchaseItemMap[pi.Description] = pi
	}

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
			existingVendorsMap[vendorPk] = newVendor
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
			purchaseEventID = existingPurchaseEventsMap[pr.BankTransaction.ID].(*models.BaserowPurchaseEventTable).BankTxID
		}

		for _, item := range pr.ReceiptItems {
			purchaseItemID, isNew := services.DerivePurchaseItem(item.Name, existingPurchaseItemMap)
			if isNew {
				purchaseItem := models.NewBaserowPurchaseItemTable(purchaseItemID)

				if err := baserowClient.CreateRow(&purchaseItem); err != nil {
					log.Fatal(fmt.Errorf("Failed to create purchase item: %w", err))
				}

				// update purchase item map
				existingPurchaseItemMap[purchaseItem.Description] = &purchaseItem
			}

			purchase := models.NewBaserowPurchaseTable(item, purchaseItemID, purchaseEventID)

			if err := baserowClient.CreateRow(purchase); err != nil {
				log.Fatal(fmt.Errorf("Failed to create purchase: %w", err))
			}
		}
	}

	// remove processed pending purchases
	if err := remove_processed_pending_purchases(ctx, baserowClient); err != nil {
		log.Fatal(fmt.Errorf("Failed to remove processed pending purchases: %w", err))
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
