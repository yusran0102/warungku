package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/middleware"
	"github.com/yusran0102/warungku/internal/models"
)

// ShowProducts renders GET /admin/inventory/products
func ShowProducts(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	search := c.Query("search", "")
	status := c.Query("status", "all")

	query := database.DB.Preload("Variants").Preload("CreatedBy")
	switch status {
	case "active":
		query = query.Where("is_active = true")
	case "inactive":
		query = query.Where("is_active = false")
	}
	if search != "" {
		query = query.Where("name ILIKE ? OR sku ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var products []models.Product
	query.Order("name ASC").Find(&products)

	return c.Render("pages/admin/products", fiber.Map{
		"Title":      "Produk – Warung-Ku",
		"ActivePage": "products",
		"Products":   products,
		"User":       claims,
		"Search":     search,
		"Status":     status,
	}, "layouts/admin")
}

// CreateProduct handles POST /admin/inventory/products
func CreateProduct(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	name := strings.TrimSpace(c.FormValue("name"))
	sku := strings.TrimSpace(c.FormValue("sku"))
	description := strings.TrimSpace(c.FormValue("description"))
	productType := models.ProductType(c.FormValue("type"))

	if name == "" || sku == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and SKU are required"})
	}

	product := models.Product{
		Name:        name,
		SKU:         sku,
		CreatedByID: claims.UserID,
		Type:        productType,
	}
	if description != "" {
		product.Description = &description
	}

	if err := database.DB.Create(&product).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return c.Status(400).JSON(fiber.Map{"error": "SKU already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create product"})
	}

	return c.Redirect("/admin/inventory/products")
}

// UpdateProduct handles PUT /admin/inventory/products/:id
func UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	claims := middleware.GetClaims(c)

	var product models.Product
	if err := database.DB.First(&product, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
	}

	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	updatedByID := claims.UserID

	updates := map[string]interface{}{
		"name":          name,
		"description":   description,
		"updated_by_id": updatedByID,
	}

	if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update product"})
	}

	return c.Redirect("/admin/inventory/products")
}

// DeleteProduct handles DELETE /admin/inventory/products/:id (soft delete)
func DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := database.DB.Model(&models.Product{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete product"})
	}

	return c.Redirect("/admin/inventory/products")
}

// CreateVariant handles POST /admin/inventory/products/variant
func CreateVariant(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	productID := c.FormValue("productId")
	name := c.FormValue("name")
	sku := c.FormValue("sku")
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
	cost, _ := strconv.ParseFloat(c.FormValue("cost"), 64)
	stock, _ := strconv.Atoi(c.FormValue("stock"))
	lowStock, _ := strconv.Atoi(c.FormValue("lowStock"))
	points, _ := strconv.Atoi(c.FormValue("points"))

	if productID == "" || name == "" || sku == "" {
		return c.Status(400).JSON(fiber.Map{"error": "All fields are required"})
	}

	variant := models.ProductVariant{
		ProductID: productID, Name: name, SKU: sku,
		Price: price, Cost: cost, Stock: stock, LowStock: lowStock, Points: points,
	}
	if err := database.DB.Create(&variant).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create variant"})
	}

	// Initial stock movement + cashflow
	var product models.Product
	database.DB.First(&product, "id = ?", productID)
	if product.Type != models.ProductTypePreorder && stock > 0 {
		database.DB.Create(&models.StockMovement{VariantID: variant.ID, Quantity: stock, Type: "IN", Notes: strPtr("Stok Awal")})
		database.DB.Create(&models.Cashflow{
			Type: models.TransactionTypeExpense, Category: "Pembelian Inventaris",
			Amount:      cost * float64(stock),
			Description: "Penambahan Stok Awal: " + name + " (" + sku + ")",
			CreatedByID: claims.UserID,
		})
	}
	return c.Redirect("/admin/inventory/products")
}

// EditProduct handles POST /admin/inventory/products/:id/edit
func EditProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	claims := middleware.GetClaims(c)
	name := strings.TrimSpace(c.FormValue("name"))
	sku := strings.TrimSpace(c.FormValue("sku"))
	description := strings.TrimSpace(c.FormValue("description"))
	productType := models.ProductType(c.FormValue("type"))

	if name == "" || sku == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and SKU required"})
	}
	var existing models.Product
	if database.DB.Where("sku = ? AND id != ?", sku, id).First(&existing).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "SKU already exists"})
	}
	updates := map[string]interface{}{"name": name, "sku": sku, "type": productType, "updated_by_id": claims.UserID}
	if description != "" {
		updates["description"] = description
	} else {
		updates["description"] = nil
	}
	database.DB.Model(&models.Product{}).Where("id = ?", id).Updates(updates)
	return c.Redirect("/admin/inventory/products")
}

// UpdateVariant handles POST /admin/inventory/products/variant/:id/update
func UpdateVariant(c *fiber.Ctx) error {
	id := c.Params("id")

	var variant models.ProductVariant
	if err := database.DB.First(&variant, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Varian tidak ditemukan"})
	}

	name := strings.TrimSpace(c.FormValue("name"))
	sku := strings.TrimSpace(c.FormValue("sku"))
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
	cost, _ := strconv.ParseFloat(c.FormValue("cost"), 64)
	lowStock, _ := strconv.Atoi(c.FormValue("lowStock"))
	points, _ := strconv.Atoi(c.FormValue("points"))
	isActive := c.FormValue("isActive") == "true"

	if name == "" || sku == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama dan SKU wajib diisi"})
	}

	// Check for duplicate SKU (excluding this variant)
	var existing models.ProductVariant
	if database.DB.Where("sku = ? AND id != ?", sku, id).First(&existing).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "SKU sudah digunakan"})
	}

	updates := map[string]interface{}{
		"name":      name,
		"sku":       sku,
		"price":     price,
		"cost":      cost,
		"low_stock": lowStock,
		"points":    points,
		"is_active": isActive,
	}
	if err := database.DB.Model(&models.ProductVariant{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui varian"})
	}

	return c.Redirect("/admin/inventory/products")
}

// DeleteVariant handles POST /admin/inventory/products/variant/:id/delete
func DeleteVariant(c *fiber.Ctx) error {
	id := c.Params("id")

	// Block deletion if variant is referenced in any sale
	var count int64
	database.DB.Model(&models.SaleItem{}).Where("variant_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Tidak bisa menghapus varian yang sudah memiliki data penjualan"})
	}

	// Block if referenced in stock movements
	database.DB.Model(&models.StockMovement{}).Where("variant_id = ?", id).Count(&count)
	if count > 0 {
		// Soft-delete only — deactivate instead
		database.DB.Model(&models.ProductVariant{}).Where("id = ?", id).Update("is_active", false)
		return c.Redirect("/admin/inventory/products")
	}

	if err := database.DB.Delete(&models.ProductVariant{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus varian"})
	}

	return c.Redirect("/admin/inventory/products")
}
