package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/ysrn87/warung-ku/internal/database"
	"github.com/ysrn87/warung-ku/internal/handlers"
	"github.com/ysrn87/warung-ku/internal/middleware"
	"github.com/ysrn87/warung-ku/internal/models"
	"github.com/ysrn87/warung-ku/internal/utils"
)

func main() {
	// ── CLI flags ──────────────────────────────────────────────────────────
	// Usage:
	//   go run main.go                  → start server (runs pending migrations)
	//   go run main.go -migrate up      → apply migrations only
	//   go run main.go -migrate down    → rollback 1 migration
	//   go run main.go -migrate down 3  → rollback 3 migrations
	//   go run main.go -migrate version → print current version
	migrateCmd := flag.String("migrate", "", "Run a migration command: up | down | version")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	database.Connect()

	// Handle -migrate flag (CLI mode — exits after running)
	if *migrateCmd != "" {
		switch *migrateCmd {
		case "up":
			database.RunMigrations()
		case "down":
			steps := 1
			if args := flag.Args(); len(args) > 0 {
				if n, err := strconv.Atoi(args[0]); err == nil {
					steps = n
				}
			}
			database.MigrateDown(steps)
		case "version":
			database.MigrateVersion()
		default:
			fmt.Fprintf(os.Stderr, "unknown migrate command %q. Use: up | down | version\n", *migrateCmd)
			os.Exit(1)
		}
		return // exit after migration command
	}

	// ── Normal server start ────────────────────────────────────────────────
	// Run pending migrations automatically on every start
	database.RunMigrations()

	engine := html.New("./templates", ".html")
	engine.Reload(false)

	engine.AddFunc("deref", func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	})
	engine.AddFunc("add", func(a, b int) int { return a + b })
	engine.AddFunc("formatRupiah", utils.FormatRupiah)
	engine.AddFunc("safe", func(s string) template.HTML { return template.HTML(s) })
	engine.AddFunc("initial", func(s string) string {
		r := []rune(s)
		if len(r) == 0 {
			return "?"
		}
		return string(r[:1])
	})
	engine.AddFunc("str", func(v interface{}) string {
		switch t := v.(type) {
		case string:
			return t
		case fmt.Stringer:
			return t.String()
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/base",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("❌ RENDER ERROR on %s: %v", c.Path(), err)
			return c.Status(500).SendString("Template error on " + c.Path() + ":\n" + err.Error())
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Static("/static", "./static")

	// ── Public ─────────────────────────────────────────────────────────────
	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/login") })
	app.Get("/login", handlers.ShowLogin)
	app.Post("/login", handlers.Login)
	app.Get("/register", handlers.ShowRegister)
	app.Post("/register", handlers.Register)
	app.Post("/logout", handlers.Logout)

	// ── Admin + Manager ────────────────────────────────────────────────────
	admin := app.Group("/admin",
		middleware.RequireRole(models.RoleAdministrator, models.RoleManager),
	)
	admin.Get("/", handlers.AdminDashboard)
	admin.Get("", handlers.AdminDashboard)

	admin.Get("/inventory/products", handlers.ShowProducts)
	admin.Post("/inventory/products", handlers.CreateProduct)
	admin.Post("/inventory/products/variant", handlers.CreateVariant)
	admin.Post("/inventory/products/:id/edit", handlers.EditProduct)
	admin.Post("/inventory/products/:id/delete", handlers.DeleteProduct)

	admin.Get("/inventory/stock", handlers.ShowStock)
	admin.Post("/inventory/stock/adjust", handlers.AdjustStock)
	admin.Get("/inventory/stock/:variantId/movements", handlers.ShowStockMovements)

	admin.Get("/sales-customers/sales", handlers.ShowSales)
	admin.Post("/sales-customers/sales", handlers.CreateSale)
	admin.Post("/sales-customers/sales/:id/delete", handlers.DeleteSale)
	admin.Post("/sales-customers/sales/:id/update", handlers.UpdateSale)

	admin.Get("/sales-customers/recap", handlers.ShowRecap)

	admin.Get("/sales-customers/customers", handlers.ShowCustomers)
	admin.Post("/sales-customers/customers/member", handlers.CreateMember)
	admin.Post("/sales-customers/customers/member/:id/delete", handlers.DeleteMember)
	admin.Post("/sales-customers/customers/member/:id/update", handlers.UpdateMember)
	admin.Post("/sales-customers/customers/non-member", handlers.CreateNonMember)
	admin.Post("/sales-customers/customers/non-member/:id/delete", handlers.DeleteNonMember)
	admin.Post("/sales-customers/customers/non-member/:id/update", handlers.UpdateNonMember)

	admin.Get("/finance/cashflow", handlers.ShowCashflow)
	admin.Post("/finance/cashflow", handlers.CreateCashflow)
	admin.Post("/finance/cashflow/:id/edit", handlers.UpdateCashflow)
	admin.Post("/finance/cashflow/:id/delete", handlers.DeleteCashflow)
	admin.Get("/finance/reports", handlers.ShowReports)

	admin.Get("/settings/points", handlers.ShowSettings)
	admin.Post("/settings/points", handlers.UpdateSettings)
	admin.Get("/settings/profile", handlers.ShowProfile)
	admin.Post("/settings/profile", handlers.UpdateProfile)
	admin.Post("/settings/profile/password", handlers.UpdatePassword)
	admin.Post("/settings/profile/managers", handlers.CreateManager)
	admin.Post("/settings/profile/managers/:id/delete", handlers.DeleteManager)

	// ── Member ─────────────────────────────────────────────────────────────
	member := app.Group("/member",
		middleware.RequireRole(models.RoleMember, models.RoleAdministrator, models.RoleManager),
	)
	member.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Member area – coming soon")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Server running at http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
