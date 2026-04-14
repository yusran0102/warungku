package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/middleware"
	"github.com/yusran0102/warungku/internal/models"
)

// ShowCashflow renders GET /admin/finance/cashflow
func ShowCashflow(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var cashflows []models.Cashflow
	database.DB.
		Preload("CreatedBy").
		Order("date DESC, created_at DESC").
		Limit(200).
		Find(&cashflows)

	// Summary totals
	var totalIncome, totalExpense float64
	for _, cf := range cashflows {
		if cf.Type == models.TransactionTypeIncome {
			totalIncome += cf.Amount
		} else {
			totalExpense += cf.Amount
		}
	}

	return c.Render("pages/admin/cashflow", fiber.Map{
		"Title":        "Arus Kas – Warung-Ku",
		"User":         claims,
		"ActivePage":   "cashflow",
		"Cashflows":    cashflows,
		"TotalIncome":  totalIncome,
		"TotalExpense": totalExpense,
		"NetCashflow":  totalIncome - totalExpense,
	}, "layouts/admin")
}

// CreateCashflow handles POST /admin/finance/cashflow
func CreateCashflow(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	if amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Jumlah harus lebih dari 0"})
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	cf := models.Cashflow{
		Type:        models.TransactionType(c.FormValue("type")),
		Category:    c.FormValue("category"),
		Amount:      amount,
		Description: c.FormValue("description"),
		Date:        date,
		CreatedByID: claims.UserID,
	}

	if err := database.DB.Create(&cf).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mencatat transaksi"})
	}

	return c.Redirect("/admin/finance/cashflow")
}

// UpdateCashflow handles POST /admin/finance/cashflow/:id/edit
func UpdateCashflow(c *fiber.Ctx) error {
	id := c.Params("id")

	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	if amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Jumlah harus lebih dari 0"})
	}

	date, err := time.Parse("2006-01-02", c.FormValue("date"))
	if err != nil {
		date = time.Now()
	}

	updates := map[string]interface{}{
		"type":        c.FormValue("type"),
		"category":    c.FormValue("category"),
		"amount":      amount,
		"description": c.FormValue("description"),
		"date":        date,
	}

	database.DB.Model(&models.Cashflow{}).Where("id = ?", id).Updates(updates)
	return c.Redirect("/admin/finance/cashflow")
}

// DeleteCashflow handles POST /admin/finance/cashflow/:id/delete
func DeleteCashflow(c *fiber.Ctx) error {
	id := c.Params("id")
	database.DB.Delete(&models.Cashflow{}, "id = ?", id)
	return c.Redirect("/admin/finance/cashflow")
}
