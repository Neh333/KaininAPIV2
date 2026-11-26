package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/neh333/KaininAPIV2/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SquareService struct {
	db *gorm.DB
}

func NewSquareService(db *gorm.DB) *SquareService {
	return &SquareService{db: db}
}

type SquareWebhookEvent struct {
	Type           string                `json:"type"`
	RelatedObjects []SquareCatalogObject `json:"related_objects"`
	Data           struct {
		Object struct {
			InventoryCounts []SquareInventoryCount `json:"inventory_counts"`
			RelatedObjects  []SquareCatalogObject  `json:"related_objects"`
		} `json:"object"`
	} `json:"data"`
}

type SquareCatalogObject struct {
	Type              string                  `json:"type"`
	ID                string                  `json:"id"`
	ItemData          SquareItemData          `json:"item_data"`
	ItemVariationData SquareItemVariationData `json:"item_variation_data"`
}

type SquareItemData struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Variations  []SquareCatalogObject `json:"variations"`
	ItemOptions []SquareItemOption    `json:"item_options"`
}

type SquareItemOption struct {
	ItemOptionID string `json:"item_option_id"`
	Name         string `json:"name"`
}

type SquareItemOptionValue struct {
	ItemOptionID      string `json:"item_option_id"`
	ItemOptionValueID string `json:"item_option_value_id"`
	Name              string `json:"name"`
}

type SquareItemVariationData struct {
	Name              string                  `json:"name"`
	Type              string                  `json:"type"`
	SKU               string                  `json:"sku"`
	PriceMoney        SquareMoney             `json:"price_money"`
	AvailableQuantity json.RawMessage         `json:"available_quantity"`
	AllocatedQuantity json.RawMessage         `json:"allocated_quantity"`
	ItemOptionValues  []SquareItemOptionValue `json:"item_option_values"`
}

type SquareMoney struct {
	Amount int64 `json:"amount"`
}

type SquareInventoryCount struct {
	CatalogObjectID string `json:"catalog_object_id"`
	Quantity        string `json:"quantity"`
	State           string `json:"state"`
	IsDeleted       bool   `json:"is_deleted"`
}

func (s *SquareService) ProcessWebhook(ctx context.Context, event SquareWebhookEvent) error {
	if !strings.EqualFold(event.Type, "inventory.count.updated") {
		return nil
	}

	var item *SquareCatalogObject
	for _, related := range event.RelatedObjects {
		if strings.EqualFold(related.Type, "ITEM") {
			item = &related
			break
		}
	}
	if item == nil {
		for _, related := range event.Data.Object.RelatedObjects {
			if strings.EqualFold(related.Type, "ITEM") {
				item = &related
				break
			}
		}
	}
	if item == nil {
		logrus.WithField("event_type", event.Type).Info("Square webhook has no related ITEM to sync")
		return nil
	}

	return s.syncCatalogItem(ctx, *item, event.Data.Object.InventoryCounts)
}

func (s *SquareService) syncCatalogItem(ctx context.Context, item SquareCatalogObject, counts []SquareInventoryCount) error {
	if item.ID == "" || item.ItemData.Name == "" {
		return errors.New("Square ITEM is missing its ID or name")
	}

	countsByVariationID := make(map[string]SquareInventoryCount, len(counts))
	for _, count := range counts {
		if count.CatalogObjectID != "" {
			countsByVariationID[count.CatalogObjectID] = count
		}
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		product, err := findOrCreateProduct(tx, item)
		if err != nil {
			return err
		}

		optionTypes := make(map[string]model.OptionType)
		optionValues := make(map[string]model.OptionValue)
		for _, variation := range item.ItemData.Variations {
			variationData := variation.ItemVariationData
			optionTypeName, optionValueName := optionNames(item.ItemData.ItemOptions, variation, variationData)
			productOptionID := int64(0)
			if optionTypeName != "" && optionValueName != "" {
				optionType, err := findOrCreateOptionType(tx, optionTypes, optionTypeName)
				if err != nil {
					return err
				}
				optionValue, err := findOrCreateOptionValue(tx, optionValues, optionType, optionValueName)
				if err != nil {
					return err
				}
				productOption, err := findOrCreateProductOption(tx, product.ProductID, optionValue.OptionValueID)
				if err != nil {
					return err
				}
				productOptionID = productOption.ProductOptionID
			}

			available, hasAvailable := parseQuantity(variationData.AvailableQuantity)
			allocated, hasAllocated := parseQuantity(variationData.AllocatedQuantity)
			if count, ok := countsByVariationID[variation.ID]; ok {
				switch {
				case count.IsDeleted:
					available, hasAvailable = 0, true
				case count.Quantity != "":
					available, hasAvailable = parseQuantity([]byte(fmt.Sprintf("%q", count.Quantity)))
				}
			}
			if err := upsertProductVariation(tx, product, variation, productOptionID, available, hasAvailable, allocated, hasAllocated); err != nil {
				return err
			}
		}

		return nil
	})
}

func findOrCreateProduct(tx *gorm.DB, item SquareCatalogObject) (model.Product, error) {
	var product model.Product
	err := tx.Where("SquareID = ?", item.ID).First(&product).Error
	switch {
	case err == nil:
		price := product.Price
		if len(item.ItemData.Variations) > 0 {
			price = float64(item.ItemData.Variations[0].ItemVariationData.PriceMoney.Amount) / 100
		}
		if err := tx.Model(&product).Updates(map[string]any{
			"Name":        item.ItemData.Name,
			"Description": item.ItemData.Description,
			"Price":       price,
		}).Error; err != nil {
			return model.Product{}, err
		}
		product.Name = item.ItemData.Name
		product.Description = item.ItemData.Description
		product.Price = price
		return product, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return model.Product{}, err
	}

	price := float64(0)
	if len(item.ItemData.Variations) > 0 {
		price = float64(item.ItemData.Variations[0].ItemVariationData.PriceMoney.Amount) / 100
	}
	product = model.Product{
		SquareID:    item.ID,
		Name:        item.ItemData.Name,
		Description: item.ItemData.Description,
		Price:       price,
	}
	if err := tx.Create(&product).Error; err != nil {
		return model.Product{}, err
	}
	return product, nil
}

func findOrCreateOptionType(tx *gorm.DB, cache map[string]model.OptionType, name string) (model.OptionType, error) {
	cacheKey := strings.ToLower(strings.TrimSpace(name))
	if optionType, ok := cache[cacheKey]; ok {
		return optionType, nil
	}

	var optionType model.OptionType
	err := tx.Where("LOWER(OptionTypeName) = ?", cacheKey).First(&optionType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		optionType = model.OptionType{OptionTypeName: strings.TrimSpace(name)}
		err = tx.Create(&optionType).Error
	}
	if err != nil {
		return model.OptionType{}, err
	}
	cache[cacheKey] = optionType
	return optionType, nil
}

func findOrCreateOptionValue(tx *gorm.DB, cache map[string]model.OptionValue, optionType model.OptionType, name string) (model.OptionValue, error) {
	cacheKey := fmt.Sprintf("%d:%s", optionType.OptionTypeID, strings.ToLower(strings.TrimSpace(name)))
	if optionValue, ok := cache[cacheKey]; ok {
		return optionValue, nil
	}

	var optionValue model.OptionValue
	err := tx.Where(
		"OptionTypeID = ? AND LOWER(OptionValueName) = ?",
		optionType.OptionTypeID,
		strings.ToLower(strings.TrimSpace(name)),
	).First(&optionValue).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		optionValue = model.OptionValue{
			OptionTypeID:    optionType.OptionTypeID,
			OptionValueName: strings.TrimSpace(name),
		}
		err = tx.Create(&optionValue).Error
	}
	if err != nil {
		return model.OptionValue{}, err
	}
	cache[cacheKey] = optionValue
	return optionValue, nil
}

func findOrCreateProductOption(tx *gorm.DB, productID, optionValueID int64) (model.ProductOption, error) {
	var productOption model.ProductOption
	err := tx.Where("ProductID = ? AND OptionValueID = ?", productID, optionValueID).First(&productOption).Error
	if err == nil {
		return productOption, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ProductOption{}, err
	}

	productOption = model.ProductOption{ProductID: productID, OptionValueID: optionValueID}
	if err := tx.Create(&productOption).Error; err != nil {
		return model.ProductOption{}, err
	}
	return productOption, nil
}

func upsertProductVariation(
	tx *gorm.DB,
	product model.Product,
	variation SquareCatalogObject,
	productOptionID int64,
	available int64,
	hasAvailable bool,
	allocated int64,
	hasAllocated bool,
) error {
	if variation.ID == "" {
		return errors.New("Square ITEM_VARIATION is missing its ID")
	}

	data := variation.ItemVariationData
	fields := map[string]any{
		"ProductID":       product.ProductID,
		"SKU":             data.SKU,
		"ProductOptionID": productOptionID,
		"Price":           float64(data.PriceMoney.Amount) / 100,
	}
	var existing model.ProductVariation
	err := tx.Where("SquareID = ?", variation.ID).First(&existing).Error
	if err == nil {
		if !hasAvailable {
			available = existing.Available
		}
		if !hasAllocated {
			allocated = existing.Allocated
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	fields["Available"] = available
	fields["Allocated"] = allocated
	fields["InStock"] = available > allocated
	if err == nil {
		return tx.Model(&existing).Updates(fields).Error
	}

	created := model.ProductVariation{
		ProductID:       product.ProductID,
		SKU:             data.SKU,
		Available:       available,
		Allocated:       allocated,
		InStock:         available > allocated,
		SquareID:        variation.ID,
		ProductOptionID: productOptionID,
		Price:           float64(data.PriceMoney.Amount) / 100,
	}
	return tx.Create(&created).Error
}

func optionNames(options []SquareItemOption, variation SquareCatalogObject, data SquareItemVariationData) (string, string) {
	if len(data.ItemOptionValues) > 0 {
		value := data.ItemOptionValues[0]
		typeName := ""
		for _, option := range options {
			if option.ItemOptionID == value.ItemOptionID {
				typeName = strings.TrimSpace(option.Name)
				break
			}
		}
		if typeName == "" {
			typeName = strings.TrimSpace(value.ItemOptionID)
		}
		valueName := strings.TrimSpace(value.Name)
		if valueName == "" {
			valueName = strings.TrimSpace(data.Name)
		}
		if typeName != "" && valueName != "" {
			return typeName, valueName
		}
	}

	typeName := strings.TrimSpace(data.Type)
	if typeName == "" {
		typeName = strings.TrimSpace(variation.Type)
	}
	if typeName == "" {
		typeName = "Variant"
	}
	return typeName, strings.TrimSpace(data.Name)
}

func parseQuantity(raw json.RawMessage) (int64, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	var quantity int64
	if err := json.Unmarshal(raw, &quantity); err == nil {
		return quantity, true
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		var parsed int64
		if _, err := fmt.Sscan(value, &parsed); err == nil {
			return parsed, true
		}
	}
	return 0, false
}
