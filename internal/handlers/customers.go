package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/middleware"
	"github.com/yusran0102/warungku/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func ShowCustomers(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var members []models.User
	database.DB.
		Where("role = ?", models.RoleMember).
		Order("created_at DESC").
		Find(&members)

	var nonMembers []models.Customer
	database.DB.Order("created_at DESC").Find(&nonMembers)

	return c.Render("pages/admin/customers", fiber.Map{
		"Title":      "Pelanggan – Warung-Ku",
		"ActivePage": "customers",
		"User":       claims,
		"Members":    members,
		"NonMembers": nonMembers,
	}, "layouts/admin")
}

func CreateMember(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	phone := normalizePhone(c.FormValue("phone"))
	email := normalizeEmail(c.FormValue("email"))
	address := strings.TrimSpace(c.FormValue("address"))
	password := c.FormValue("password")
	birthday := c.FormValue("birthday")

	if name == "" || phone == "" || password == "" || address == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama, telepon, password, dan alamat wajib diisi"})
	}
	if len(password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password minimal 6 karakter"})
	}

	var existing models.User
	if database.DB.Where("phone = ?", phone).First(&existing).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Nomor telepon sudah terdaftar"})
	}
	if email != "" {
		if database.DB.Where("email = ?", email).First(&existing).Error == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Email sudah terdaftar"})
		}
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Name:     name,
		Phone:    phone,
		Password: string(hashed),
		Role:     models.RoleMember,
		Points:   0,
	}
	if address != "" {
		user.Address = &address
	}
	if email != "" {
		user.Email = &email
	}
	if birthday != "" {
		t, err := time.Parse("2006-01-02", birthday)
		if err == nil {
			user.Birthday = &t
		}
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat member"})
	}

	return c.Redirect("/admin/sales-customers/customers")
}

func DeleteMember(c *fiber.Ctx) error {
	id := c.Params("id")
	var count int64
	database.DB.Model(&models.Sale{}).Where("customer_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Tidak bisa menghapus pelanggan yang memiliki data penjualan"})
	}
	database.DB.Where("id = ? AND role = ?", id, models.RoleMember).Delete(&models.User{})
	return c.Redirect("/admin/sales-customers/customers")
}

func CreateNonMember(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	phone := normalizePhone(c.FormValue("phone"))
	address := strings.TrimSpace(c.FormValue("address"))

	if name == "" || phone == "" || address == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama, telepon, dan alamat wajib diisi"})
	}

	var existingUser models.User
	var existingCustomer models.Customer
	if database.DB.Where("phone = ?", phone).First(&existingUser).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Nomor telepon sudah terdaftar sebagai member"})
	}
	if database.DB.Where("phone = ?", phone).First(&existingCustomer).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Nomor telepon sudah terdaftar sebagai pelanggan"})
	}

	customer := models.Customer{Name: name, Phone: phone, Address: &address}
	if err := database.DB.Create(&customer).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat pelanggan"})
	}
	return c.Redirect("/admin/sales-customers/customers")
}

func DeleteNonMember(c *fiber.Ctx) error {
	id := c.Params("id")
	var count int64
	database.DB.Model(&models.Sale{}).Where("non_member_customer_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Tidak bisa menghapus pelanggan yang memiliki data penjualan"})
	}
	database.DB.Delete(&models.Customer{}, "id = ?", id)
	return c.Redirect("/admin/sales-customers/customers")
}

func UpdateMember(c *fiber.Ctx) error {
	id := c.Params("id")
	name := strings.TrimSpace(c.FormValue("name"))
	phone := normalizePhone(c.FormValue("phone"))
	email := normalizeEmail(c.FormValue("email"))
	address := strings.TrimSpace(c.FormValue("address"))
	birthday := c.FormValue("birthday")

	if name == "" || phone == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama dan telepon wajib diisi"})
	}

	updates := map[string]interface{}{
		"name":  name,
		"phone": phone,
	}
	if email != "" {
		updates["email"] = &email
	} else {
		updates["email"] = nil
	}
	if address != "" {
		updates["address"] = &address
	}
	if birthday != "" {
		t, err := time.Parse("2006-01-02", birthday)
		if err == nil {
			updates["birthday"] = &t
		}
	} else {
		updates["birthday"] = nil
	}

	database.DB.Model(&models.User{}).Where("id = ? AND role = ?", id, models.RoleMember).Updates(updates)
	return c.Redirect("/admin/sales-customers/customers")
}

func UpdateNonMember(c *fiber.Ctx) error {
	id := c.Params("id")
	name := strings.TrimSpace(c.FormValue("name"))
	phone := normalizePhone(c.FormValue("phone"))
	address := strings.TrimSpace(c.FormValue("address"))

	if name == "" || phone == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama dan telepon wajib diisi"})
	}

	updates := map[string]interface{}{
		"name":  name,
		"phone": phone,
	}
	if address != "" {
		updates["address"] = &address
	}

	database.DB.Model(&models.Customer{}).Where("id = ?", id).Updates(updates)
	return c.Redirect("/admin/sales-customers/customers")
}
