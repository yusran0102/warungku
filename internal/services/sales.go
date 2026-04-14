package services

import (
	"errors"
	"fmt"

	"github.com/ysrn87/warung-ku/internal/database"
	"github.com/ysrn87/warung-ku/internal/models"
	"github.com/ysrn87/warung-ku/internal/utils"
	"gorm.io/gorm"
)

type SaleItemInput struct {
	VariantID string
	Quantity  int
	Price     float64
}

type CreateSaleInput struct {
	Items               []SaleItemInput
	CustomerID          *string
	NonMemberCustomerID *string
	PaymentMethod       string
	PaymentStatus       models.PaymentStatus
	Discount            float64
	Tax                 float64
	Ongkir              float64
	Notes               *string
	PointsRedeemed      int
	CashierID           string
}

// CreateSale mirrors createSaleAction in sales.ts
func CreateSale(input CreateSaleInput) (*models.Sale, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("no items in sale")
	}
	if input.CustomerID != nil && input.NonMemberCustomerID != nil {
		return nil, errors.New("cannot assign both member and non-member customer")
	}
	if input.PointsRedeemed > 0 && input.NonMemberCustomerID != nil {
		return nil, errors.New("points redemption is only available for members")
	}

	// Validate available points
	if input.PointsRedeemed > 0 && input.CustomerID != nil {
		available, err := GetAvailablePoints(*input.CustomerID)
		if err != nil {
			return nil, err
		}
		if input.PointsRedeemed > available {
			return nil, fmt.Errorf("insufficient points. Available: %d", available)
		}
	}

	if input.PaymentStatus == "" {
		input.PaymentStatus = models.PaymentStatusPaid
	}

	// Validate stock + calculate points earned
	pointsEarned := 0
	for _, item := range input.Items {
		var variant models.ProductVariant
		if err := database.DB.Preload("Product").First(&variant, "id = ?", item.VariantID).Error; err != nil {
			return nil, errors.New("product variant not found")
		}
		if variant.Product.Type != models.ProductTypePreorder && variant.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s", variant.Name)
		}
		if input.CustomerID != nil && input.PointsRedeemed == 0 && input.PaymentStatus == models.PaymentStatusPaid {
			pointsEarned += variant.Points * item.Quantity
		}
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range input.Items {
		subtotal += item.Price * float64(item.Quantity)
	}
	conversionRate := GetPointsConversionRate()
	pointDiscount := float64(input.PointsRedeemed) * float64(conversionRate)
	totalDiscount := input.Discount + pointDiscount

	if totalDiscount > subtotal {
		return nil, fmt.Errorf("total discount (%.0f) cannot exceed subtotal (%.0f)", totalDiscount, subtotal)
	}

	total := subtotal - input.Discount - pointDiscount + input.Tax + input.Ongkir
	if total < 0 {
		return nil, errors.New("total payment cannot be negative")
	}

	var sale models.Sale

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Create sale
		earnedForSale := 0
		if input.CustomerID != nil && input.PointsRedeemed == 0 && input.PaymentStatus == models.PaymentStatusPaid {
			earnedForSale = pointsEarned
		}
		redeemedForSale := 0
		if input.CustomerID != nil {
			redeemedForSale = input.PointsRedeemed
		}

		sale = models.Sale{
			SaleNumber:          utils.GenerateSaleNumber(),
			CashierID:           input.CashierID,
			CustomerID:          input.CustomerID,
			NonMemberCustomerID: input.NonMemberCustomerID,
			Subtotal:            subtotal,
			Discount:            input.Discount,
			Tax:                 input.Tax,
			Ongkir:              input.Ongkir,
			Total:               total,
			PaymentMethod:       input.PaymentMethod,
			PaymentStatus:       input.PaymentStatus,
			Notes:               input.Notes,
			PointsEarned:        earnedForSale,
			PointsRedeemed:      redeemedForSale,
		}
		if err := tx.Create(&sale).Error; err != nil {
			return err
		}

		// Create sale items + update stock
		for _, item := range input.Items {
			saleItem := models.SaleItem{
				SaleID:    sale.ID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
				Price:     item.Price,
				Subtotal:  item.Price * float64(item.Quantity),
			}
			if err := tx.Create(&saleItem).Error; err != nil {
				return err
			}

			var variant models.ProductVariant
			tx.Preload("Product").First(&variant, "id = ?", item.VariantID)
			if variant.Product.Type == models.ProductTypePreorder {
				continue
			}

			if err := tx.Model(&models.ProductVariant{}).Where("id = ?", item.VariantID).
				Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.StockMovement{
				VariantID: item.VariantID,
				Quantity:  -item.Quantity,
				Type:      "OUT",
				Notes:     strPtr("PENJUALAN " + sale.SaleNumber),
			}).Error; err != nil {
				return err
			}
		}

		// Points handling
		if input.CustomerID != nil {
			if input.PointsRedeemed > 0 {
				if err := tx.Model(&models.User{}).Where("id = ?", *input.CustomerID).
					Update("points", gorm.Expr("points - ?", input.PointsRedeemed)).Error; err != nil {
					return err
				}
				tx.Create(&models.PointHistory{
					UserID:      *input.CustomerID,
					Points:      -input.PointsRedeemed,
					Type:        "REDEEMED",
					Description: "Penukaran poin " + sale.SaleNumber,
				})
			} else if pointsEarned > 0 {
				if err := tx.Model(&models.User{}).Where("id = ?", *input.CustomerID).
					Update("points", gorm.Expr("points + ?", pointsEarned)).Error; err != nil {
					return err
				}
				expiry := utils.PointsExpiryDate(sale.CreatedAt)
				tx.Create(&models.PointHistory{
					UserID:      *input.CustomerID,
					Points:      pointsEarned,
					Type:        "EARNED",
					Description: "Poin pembelian " + sale.SaleNumber,
					ExpiresAt:   &expiry,
				})
			}
		}

		// Auto cashflow entry
		customerName := "Pelanggan umum"
		if input.CustomerID != nil {
			var u models.User
			tx.Select("name").First(&u, "id = ?", *input.CustomerID)
			customerName = u.Name
		}
		tx.Create(&models.Cashflow{
			Type:        models.TransactionTypeIncome,
			Category:    "Penjualan",
			Amount:      total,
			Description: fmt.Sprintf("Sale %s - %s", sale.SaleNumber, customerName),
			CreatedByID: input.CashierID,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// DeleteSale mirrors deleteSaleAction - restores stock, reverses points, adds cashflow entry
func DeleteSale(saleID string, deletedByID string) error {
	var sale models.Sale
	if err := database.DB.Preload("Items").First(&sale, "id = ?", saleID).Error; err != nil {
		return errors.New("sale not found")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Restore stock
		for _, item := range sale.Items {
			var variant models.ProductVariant
			tx.Preload("Product").First(&variant, "id = ?", item.VariantID)
			if variant.Product.Type == models.ProductTypePreorder {
				continue
			}
			tx.Model(&models.ProductVariant{}).Where("id = ?", item.VariantID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity))
			tx.Create(&models.StockMovement{
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
				Type:      "IN",
				Notes:     strPtr("Stok dikembalikan dari penghapusan " + sale.SaleNumber),
			})
		}

		// Reverse points
		if sale.CustomerID != nil {
			if sale.PointsEarned > 0 {
				tx.Model(&models.User{}).Where("id = ?", *sale.CustomerID).
					Update("points", gorm.Expr("points - ?", sale.PointsEarned))
				tx.Create(&models.PointHistory{
					UserID:      *sale.CustomerID,
					Points:      -sale.PointsEarned,
					Type:        "ADJUSTED",
					Description: "Poin dikembalikan dari penghapusan " + sale.SaleNumber,
				})
			}
			if sale.PointsRedeemed > 0 {
				tx.Model(&models.User{}).Where("id = ?", *sale.CustomerID).
					Update("points", gorm.Expr("points + ?", sale.PointsRedeemed))
				tx.Create(&models.PointHistory{
					UserID:      *sale.CustomerID,
					Points:      sale.PointsRedeemed,
					Type:        "ADJUSTED",
					Description: "Poin dikembalikan dari penghapusan " + sale.SaleNumber,
				})
			}
		}

		// Cashflow reversal
		tx.Create(&models.Cashflow{
			Type:        models.TransactionTypeExpense,
			Category:    "Penghapusan Penjualan",
			Amount:      sale.Total,
			Description: "Penghapusan " + sale.SaleNumber,
			CreatedByID: deletedByID,
		})

		return tx.Delete(&sale).Error
	})
}

func strPtr(s string) *string { return &s }
