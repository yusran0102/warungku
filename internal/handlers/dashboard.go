package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warung-ku/internal/database"
	"github.com/yusran0102/warung-ku/internal/middleware"
	"github.com/yusran0102/warung-ku/internal/models"
)

func AdminDashboard(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var totalVariants, totalSales, totalMembers, lowStockCount int64
	database.DB.Model(&models.ProductVariant{}).Count(&totalVariants)
	database.DB.Model(&models.Sale{}).Count(&totalSales)
	database.DB.Model(&models.User{}).Where("role = ?", models.RoleMember).Count(&totalMembers)
	database.DB.Model(&models.ProductVariant{}).Where("stock <= low_stock AND is_active = true").Count(&lowStockCount)

	var totalRevenue float64
	database.DB.Model(&models.Sale{}).Select("COALESCE(SUM(total),0)").Scan(&totalRevenue)

	var recentSales []models.Sale
	database.DB.Preload("Customer").Preload("NonMemberCustomer").
		Order("created_at DESC").Limit(5).Find(&recentSales)

	var lowStockItems []models.ProductVariant
	database.DB.Preload("Product").Where("stock <= low_stock AND is_active = true").Limit(5).Find(&lowStockItems)

	return c.Render("pages/admin/dashboard", fiber.Map{
		"Title":         "Beranda – Warung-Ku",
		"ActivePage":    "dashboard",
		"User":          claims,
		"TotalVariants": totalVariants,
		"TotalSales":    totalSales,
		"TotalMembers":  totalMembers,
		"LowStockCount": lowStockCount,
		"TotalRevenue":  totalRevenue,
		"RecentSales":   recentSales,
		"LowStockItems": lowStockItems,
	}, "layouts/admin")
}
