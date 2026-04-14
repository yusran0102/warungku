package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/middleware"
	"github.com/yusran0102/warungku/internal/models"
	"github.com/yusran0102/warungku/internal/services"
)

func ShowSales(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var sales []models.Sale
	database.DB.
		Preload("Customer").
		Preload("NonMemberCustomer").
		Preload("Cashier").
		Preload("Items.Variant.Product").
		Order("created_at DESC").
		Limit(100).
		Find(&sales)

	// Only active variants with stock > 0 or preorder type
	var variants []models.ProductVariant
	database.DB.
		Preload("Product").
		Where("product_variants.is_active = true").
		Joins("JOIN products ON products.id = product_variants.product_id AND products.is_active = true").
		Find(&variants)

	var members []models.User
	database.DB.Select("id, name, phone, points").
		Where("role = ?", models.RoleMember).
		Order("name ASC").Find(&members)

	var nonMembers []models.Customer
	database.DB.Order("name ASC").Find(&nonMembers)

	conversionRate := services.GetPointsConversionRate()

	return c.Render("pages/admin/sales", fiber.Map{
		"Title":          "Penjualan – Warung-Ku",
		"ActivePage":     "sales",
		"User":           claims,
		"Sales":          sales,
		"Variants":       variants,
		"Members":        members,
		"NonMembers":     nonMembers,
		"ConversionRate": conversionRate,
	}, "layouts/admin")
}

func CreateSale(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var items []services.SaleItemInput
	for i := 0; ; i++ {
		vid := c.FormValue("items[" + strconv.Itoa(i) + "][variantId]")
		if vid == "" {
			break
		}
		qty, _ := strconv.Atoi(c.FormValue("items[" + strconv.Itoa(i) + "][quantity]"))
		price, _ := strconv.ParseFloat(c.FormValue("items["+strconv.Itoa(i)+"][price]"), 64)
		if qty > 0 && price > 0 {
			items = append(items, services.SaleItemInput{VariantID: vid, Quantity: qty, Price: price})
		}
	}

	discount, _ := strconv.ParseFloat(c.FormValue("discount"), 64)
	tax, _ := strconv.ParseFloat(c.FormValue("tax"), 64)
	ongkir, _ := strconv.ParseFloat(c.FormValue("ongkir"), 64)
	pointsRedeemed, _ := strconv.Atoi(c.FormValue("pointsRedeemed"))
	paymentStatus := models.PaymentStatus(c.FormValue("paymentStatus"))
	if paymentStatus == "" {
		paymentStatus = models.PaymentStatusPaid
	}

	var customerID *string
	var nonMemberID *string
	if cid := strings.TrimSpace(c.FormValue("customerId")); cid != "" {
		customerID = &cid
	}
	if nmid := strings.TrimSpace(c.FormValue("nonMemberCustomerId")); nmid != "" {
		nonMemberID = &nmid
	}

	notes := strings.TrimSpace(c.FormValue("notes"))
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	input := services.CreateSaleInput{
		Items:               items,
		CustomerID:          customerID,
		NonMemberCustomerID: nonMemberID,
		PaymentMethod:       c.FormValue("paymentMethod"),
		PaymentStatus:       paymentStatus,
		Discount:            discount,
		Tax:                 tax,
		Ongkir:              ongkir,
		Notes:               notesPtr,
		PointsRedeemed:      pointsRedeemed,
		CashierID:           claims.UserID,
	}

	if _, err := services.CreateSale(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect("/admin/sales-customers/sales")
}

func DeleteSale(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	id := c.Params("id")

	if err := services.DeleteSale(id, claims.UserID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect("/admin/sales-customers/sales")
}

func UpdateSale(c *fiber.Ctx) error {
	id := c.Params("id")

	paymentMethod := strings.TrimSpace(c.FormValue("paymentMethod"))
	paymentStatus := models.PaymentStatus(c.FormValue("paymentStatus"))
	discount, _ := strconv.ParseFloat(c.FormValue("discount"), 64)
	tax, _ := strconv.ParseFloat(c.FormValue("tax"), 64)
	ongkir, _ := strconv.ParseFloat(c.FormValue("ongkir"), 64)
	notes := strings.TrimSpace(c.FormValue("notes"))

	var sale models.Sale
	if err := database.DB.First(&sale, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Penjualan tidak ditemukan"})
	}

	sale.PaymentMethod = paymentMethod
	sale.PaymentStatus = paymentStatus
	sale.Discount = discount
	sale.Tax = tax
	sale.Ongkir = ongkir
	if notes != "" {
		sale.Notes = &notes
	} else {
		sale.Notes = nil
	}

	// Parse updated items (es_items[N][variantId/quantity/price])
	var newItems []models.SaleItem
	var subtotal float64
	for i := 0; ; i++ {
		vid := strings.TrimSpace(c.FormValue("es_items[" + strconv.Itoa(i) + "][variantId]"))
		if vid == "" {
			break
		}
		qty, _ := strconv.Atoi(c.FormValue("es_items[" + strconv.Itoa(i) + "][quantity]"))
		price, _ := strconv.ParseFloat(c.FormValue("es_items["+strconv.Itoa(i)+"][price]"), 64)
		if qty <= 0 {
			qty = 1
		}
		itemSubtotal := price * float64(qty)
		subtotal += itemSubtotal
		newItems = append(newItems, models.SaleItem{
			SaleID:    id,
			VariantID: vid,
			Quantity:  qty,
			Price:     price,
			Subtotal:  itemSubtotal,
		})
	}

	// Only replace items if the form submitted any
	if len(newItems) > 0 {
		if err := database.DB.Where("sale_id = ?", id).Delete(&models.SaleItem{}).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus item lama"})
		}
		if err := database.DB.Create(&newItems).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan item baru"})
		}
		sale.Subtotal = subtotal
	}

	sale.Total = sale.Subtotal - discount + tax + ongkir

	if err := database.DB.Save(&sale).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui penjualan"})
	}

	return c.Redirect("/admin/sales-customers/sales")
}
