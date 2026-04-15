package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warung-ku/internal/database"
	"github.com/yusran0102/warung-ku/internal/middleware"
	"github.com/yusran0102/warung-ku/internal/models"
)

type RecapRow struct {
	VariantID         string
	ProductName       string
	VariantName       string
	SKU               string
	TotalQty          int
	TotalRevenue      float64
	AvgPrice          float64
	TotalTransactions int
}

// InventoryRow adds a pre-computed StockValue so templates don't need mul
type InventoryRow struct {
	models.ProductVariant
	StockValue float64
}

func ShowRecap(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var items []models.SaleItem
	database.DB.Preload("Variant.Product").Find(&items)

	rowMap := map[string]*RecapRow{}
	saleIDs := map[string]bool{}
	for _, item := range items {
		v := item.Variant
		r, ok := rowMap[v.ID]
		if !ok {
			r = &RecapRow{VariantID: v.ID, ProductName: v.Product.Name, VariantName: v.Name, SKU: v.SKU}
			rowMap[v.ID] = r
		}
		r.TotalQty += item.Quantity
		r.TotalRevenue += item.Subtotal
		r.TotalTransactions++
		saleIDs[item.SaleID] = true
	}

	rows := []*RecapRow{}
	var totalRevenue float64
	var totalQty int
	for _, r := range rowMap {
		if r.TotalQty > 0 {
			r.AvgPrice = r.TotalRevenue / float64(r.TotalQty)
		}
		rows = append(rows, r)
		totalRevenue += r.TotalRevenue
		totalQty += r.TotalQty
	}

	return c.Render("pages/admin/recap", fiber.Map{
		"Title":             "Rekap Penjualan – Warung-Ku",
		"ActivePage":        "recap",
		"User":              claims,
		"Rows":              rows,
		"TotalRevenue":      totalRevenue,
		"TotalQty":          totalQty,
		"TotalTransactions": len(saleIDs),
	}, "layouts/admin")
}

func ShowReports(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var totalIncome, totalExpense, totalSalesRevenue float64
	database.DB.Model(&models.Cashflow{}).Where("type = ?", "INCOME").Select("COALESCE(SUM(amount),0)").Scan(&totalIncome)
	database.DB.Model(&models.Cashflow{}).Where("type = ?", "EXPENSE").Select("COALESCE(SUM(amount),0)").Scan(&totalExpense)
	database.DB.Model(&models.Sale{}).Select("COALESCE(SUM(total),0)").Scan(&totalSalesRevenue)

	var totalSalesCount, totalProducts int64
	database.DB.Model(&models.Sale{}).Count(&totalSalesCount)
	database.DB.Model(&models.ProductVariant{}).Count(&totalProducts)

	// Pre-compute StockValue per row so template doesn't need mul
	var variants []models.ProductVariant
	database.DB.Preload("Product").Order("stock ASC").Limit(50).Find(&variants)

	var inventoryValue float64
	inventoryRows := make([]InventoryRow, 0, len(variants))
	for _, v := range variants {
		sv := v.Cost * float64(v.Stock)
		inventoryValue += sv
		inventoryRows = append(inventoryRows, InventoryRow{ProductVariant: v, StockValue: sv})
	}

	var recentSales []models.Sale
	database.DB.Preload("Customer").Preload("Items.Variant.Product").
		Order("created_at DESC").Limit(10).Find(&recentSales)

	return c.Render("pages/admin/reports", fiber.Map{
		"Title":             "Laporan – Warung-Ku",
		"ActivePage":        "reports",
		"User":              claims,
		"TotalIncome":       totalIncome,
		"TotalExpense":      totalExpense,
		"TotalSalesRevenue": totalSalesRevenue,
		"NetProfit":         totalIncome - totalExpense,
		"TotalSalesCount":   totalSalesCount,
		"TotalProducts":     totalProducts,
		"InventoryValue":    inventoryValue,
		"Inventory":         inventoryRows,
		"RecentSales":       recentSales,
	}, "layouts/admin")
}
