package routes

import (
	"net/http"

	"symetra-lab-backend/config"
	"symetra-lab-backend/handlers"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter mengatur semua rute HTTP dan middleware Gin
func SetupRouter(cfg *config.Config, db *gorm.DB, shopeeClient *shopee.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Middleware CORS agar backend bisa dipanggil dari frontend web
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Endpoint health check
	r.GET("/health", func(c *gin.Context) {
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected / not configured"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"database":    dbStatus,
			"environment": cfg.AppEnv,
		})
	})

	// Dashboard UI untuk screenshot client & status manajemen toko
	renderDashboard := func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `
		<!DOCTYPE html>
		<html lang="id">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Symetra Lab - E-Commerce Store Management</title>
			<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
			<style>
				* { margin: 0; padding: 0; box-sizing: border-box; font-family: 'Inter', -apple-system, sans-serif; }
				body { background: #0f172a; color: #f8fafc; min-height: 100vh; display: flex; }
				.sidebar { width: 260px; background: #1e293b; border-right: 1px solid #334155; padding: 1.5rem; display: flex; flex-direction: column; gap: 2rem; }
				.brand { display: flex; align-items: center; gap: 0.75rem; font-size: 1.25rem; font-weight: 700; color: #ee4d2d; }
				.brand-logo { width: 36px; height: 36px; background: #ee4d2d; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 800; }
				.nav-list { list-style: none; display: flex; flex-direction: column; gap: 0.5rem; }
				.nav-item { padding: 0.75rem 1rem; border-radius: 8px; color: #94a3b8; font-size: 0.9rem; font-weight: 500; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 0.75rem; }
				.nav-item.active, .nav-item:hover { background: #334155; color: #f8fafc; }
				.main-content { flex: 1; padding: 2rem 3rem; overflow-y: auto; }
				.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; }
				.header h1 { font-size: 1.75rem; font-weight: 700; }
				.header p { color: #94a3b8; font-size: 0.9rem; margin-top: 0.25rem; }
				.btn-connect { background: #ee4d2d; color: white; border: none; padding: 0.65rem 1.25rem; border-radius: 8px; font-weight: 600; cursor: pointer; font-size: 0.9rem; text-decoration: none; display: inline-flex; align-items: center; gap: 0.5rem; transition: background 0.2s; }
				.btn-connect:hover { background: #d73211; }
				.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1.25rem; margin-bottom: 2rem; }
				.stat-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 1.25rem; }
				.stat-title { color: #94a3b8; font-size: 0.85rem; font-weight: 500; }
				.stat-value { font-size: 1.75rem; font-weight: 700; margin: 0.5rem 0; color: #f8fafc; }
				.stat-desc { font-size: 0.8rem; color: #10b981; font-weight: 500; }
				.content-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 1.5rem; }
				.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; }
				.badge { background: #064e3b; color: #34d399; font-size: 0.75rem; font-weight: 600; padding: 0.25rem 0.65rem; border-radius: 9999px; }
				.table { width: 100%; border-collapse: collapse; text-align: left; }
				.table th { padding: 0.75rem 1rem; color: #94a3b8; font-weight: 600; font-size: 0.8rem; text-transform: uppercase; border-bottom: 1px solid #334155; }
				.table td { padding: 1rem; border-bottom: 1px solid #334155; font-size: 0.9rem; }
			</style>
		</head>
		<body>
			<div class="sidebar">
				<div class="brand">
					<div class="brand-logo">S</div>
					<span>Symetra Lab</span>
				</div>
				<ul class="nav-list">
					<li class="nav-item active">📊 Dashboard</li>
					<li class="nav-item">📦 Orders & Fulfillment</li>
					<li class="nav-item">🛍️ Products & Inventory</li>
					<li class="nav-item">🏪 Connected Stores</li>
					<li class="nav-item">⚙️ Settings</li>
				</ul>
			</div>
			<div class="main-content">
				<div class="header">
					<div>
						<h1>Store Management Console</h1>
						<p>Internal ERP & Order Processing System for 3D Printing E-Commerce</p>
					</div>
					<a href="/api/v1/shopee/auth-url?redirect=true" class="btn-connect">
						+ Connect Shopee Shop
					</a>
				</div>
				<div class="stats-grid">
					<div class="stat-card">
						<div class="stat-title">Marketplace Channel</div>
						<div class="stat-value">Shopee ID</div>
						<div class="stat-desc">● Open Platform v2 Active</div>
					</div>
					<div class="stat-card">
						<div class="stat-title">Database Status</div>
						<div class="stat-value">Supabase PG</div>
						<div class="stat-desc">● Cloud Pooler Connected</div>
					</div>
					<div class="stat-card">
						<div class="stat-title">Order Sync</div>
						<div class="stat-value">Real-Time</div>
						<div class="stat-desc">● Automated Webhook Ready</div>
					</div>
					<div class="stat-card">
						<div class="stat-title">System Status</div>
						<div class="stat-value">Live (Vercel)</div>
						<div class="stat-desc">● High Availability</div>
					</div>
				</div>
				<div class="content-card">
					<div class="card-header">
						<h2>Channel Integration Overview</h2>
						<span class="badge">Active & Ready</span>
					</div>
					<table class="table">
						<thead>
							<tr>
								<th>Platform</th>
								<th>Channel Type</th>
								<th>Status</th>
								<th>API Version</th>
								<th>Action</th>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td><strong>Shopee Indonesia</strong></td>
								<td>E-Commerce Store</td>
								<td><span class="badge">Ready for Authorization</span></td>
								<td>Open API v2.0</td>
								<td><a href="/api/v1/shopee/auth-url?redirect=true" style="color: #38bdf8; text-decoration: none; font-weight: 500;">Authorize Now &rarr;</a></td>
							</tr>
						</tbody>
					</table>
				</div>
			</div>
		</body>
		</html>
		`)
	}

	// Route dashboard
	r.GET("/", renderDashboard)
	r.GET("/dashboard", renderDashboard)

	// Group route Shopee
	shopeeHandler := handlers.NewShopeeHandler(db, shopeeClient)
	v1 := r.Group("/api/v1")
	{
		shopeeRoutes := v1.Group("/shopee")
		{
			shopeeRoutes.GET("/auth-url", shopeeHandler.GetAuthURL)
			shopeeRoutes.GET("/callback", shopeeHandler.HandleCallback)
			shopeeRoutes.GET("/shops", shopeeHandler.GetShops)
			shopeeRoutes.POST("/shops/:shop_id/refresh", shopeeHandler.RefreshToken)
		}
	}

	return r
}
