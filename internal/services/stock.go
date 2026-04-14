package services

import (
	"errors"
	"fmt"

	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/models"
	"gorm.io/gorm"
)

type StockAdjustInput struct {
	VariantID string
	Type      string // IN | OUT | ADJUSTMENT
	Quantity  int
	Notes     string
	UserID    string
}

// AdjustStock mirrors adjustStockAction in stock.ts
func AdjustStock(input StockAdjustInput) error {
	var variant models.ProductVariant
	if err := database.DB.Preload("Product").First(&variant, "id = ?", input.VariantID).Error; err != nil {
		return errors.New("product variant not found")
	}

	isPreorder := variant.Product.Type == models.ProductTypePreorder

	newStock := variant.Stock
	if !isPreorder {
		switch input.Type {
		case "IN":
			newStock += input.Quantity
		case "OUT":
			newStock -= input.Quantity
			if newStock < 0 {
				return errors.New("insufficient stock")
			}
		case "ADJUSTMENT":
			newStock = input.Quantity
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if !isPreorder {
			if err := tx.Model(&models.ProductVariant{}).Where("id = ?", input.VariantID).
				Update("stock", newStock).Error; err != nil {
				return err
			}
			var notes *string
			if input.Notes != "" {
				notes = &input.Notes
			}
			if err := tx.Create(&models.StockMovement{
				VariantID: input.VariantID,
				Quantity:  input.Quantity,
				Type:      input.Type,
				Notes:     notes,
			}).Error; err != nil {
				return err
			}
		}

		// Auto-create cashflow for stock IN (purchasing inventory)
		if !isPreorder && input.Type == "IN" && input.Quantity > 0 {
			totalCost := variant.Cost * float64(input.Quantity)
			desc := fmt.Sprintf("Pembelian %d unit %s - %s", input.Quantity, variant.Product.Name, variant.Name)
			if input.Notes != "" {
				desc += " (" + input.Notes + ")"
			}
			tx.Create(&models.Cashflow{
				Type:        models.TransactionTypeExpense,
				Category:    "Pembelian Inventaris",
				Amount:      totalCost,
				Description: desc,
				CreatedByID: input.UserID,
			})
		}

		return nil
	})
}
