package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/middleware"
	"github.com/yusran0102/warungku/internal/models"
	"github.com/yusran0102/warungku/internal/services"
)

func ShowStock(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var variants []models.ProductVariant
	database.DB.
		Preload("Product").
		Where("is_active = ?", true).
		Order("updated_at DESC").
		Find(&variants)

	// Bug fix: compute stat card values here, not in template
	var lowStockCount int
	var totalStock int
	for _, v := range variants {
		totalStock += v.Stock
		if v.Stock <= v.LowStock {
			lowStockCount++
		}
	}

	return c.Render("pages/admin/stock", fiber.Map{
		"Title":         "Stok – warungku",
		"User":          claims,
		"ActivePage":    "stock",
		"Variants":      variants,
		"LowStockCount": lowStockCount, // was missing → blank card
		"TotalStock":    totalStock,    // was missing → blank card
	}, "layouts/admin")
}

func AdjustStock(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	qty, _ := strconv.Atoi(c.FormValue("quantity"))
	input := services.StockAdjustInput{
		VariantID: c.FormValue("variantId"),
		Type:      c.FormValue("type"),
		Quantity:  qty,
		Notes:     c.FormValue("reason"),
		UserID:    claims.UserID,
	}

	if err := services.AdjustStock(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect("/admin/inventory/stock")
}

func ShowStockMovements(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	variantID := c.Params("variantId")

	var variant models.ProductVariant
	database.DB.Preload("Product").First(&variant, "id = ?", variantID)

	var movements []models.StockMovement
	database.DB.
		Where("variant_id = ?", variantID).
		Order("created_at DESC").
		Limit(50).
		Find(&movements)

	return c.Render("pages/admin/stock_movements", fiber.Map{
		"Title":     "Riwayat Stok – warungku",
		"User":      claims,
		"Variant":   variant,
		"Movements": movements,
	}, "layouts/admin")
}
