package api

import (
	"net/http"

	"go-backend/configs"
	"go-backend/internals/api/handlers"
	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/repository"
	"go-backend/internals/services"
	"go-backend/internals/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *configs.Config, hub *websocket.Hub) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.CORSMiddleware())

	// Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	handoverRepo := repository.NewHandoverRepository(db)
	userGenericRepo := repository.NewGenericRepository[domains.USERINFO](db, "UserID", []string{"UserID", "UserName", "Description", "Area", "Role"})
	handoverGenericRepo := repository.NewGenericRepository[domains.HANDOVER](db, "ID", []string{"ID", "UnitCode", "UserCreated", "UserTransfer"})

	// Initialize Services
	authService := services.NewAuthService(userRepo, cfg)
	assetService := services.NewAssetService(assetRepo, userRepo, authService)
	userService := services.NewUserInfoService(userRepo, db)
	handoverService := services.NewHandoverService(handoverRepo, userRepo, db)
	masterService := services.NewMasterService(db)
	syncService := services.NewAutoSyncService(db)

	if cfg.AutoSync.Enabled {
		syncService.StartScheduler(cfg.AutoSync.CronSchedule)
	}

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, authService, userGenericRepo)
	assetHandler := handlers.NewAssetHandler(assetService)
	handoverHandler := handlers.NewHandoverHandler(handoverService, handoverGenericRepo)
	syncHandler := handlers.NewAutoSyncHandler(syncService)

	// Public Health Check
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Service ready")
	})

	// WebSocket Endpoint
	r.GET("/ws", websocket.HandleWebSocket(hub))

	// API Group
	api := r.Group("/api")

	// 1. Auths Controller (support both /auths and /Auths)
	for _, authPath := range []string{"/auths", "/Auths"} {
		auths := api.Group(authPath)
		{
			auths.GET("", authHandler.GetByAuthCode)
			auths.POST("/login", authHandler.Login)
			auths.POST("/refresh", authHandler.Refresh)
		}
	}

	// Attach Token Validation Middleware for remaining API routes
	api.Use(middlewares.TokenValidationMiddleware(cfg.JWT.SecretKey))

	// 2. Users Controller
	for _, userPath := range []string{"/users", "/Users", "/userinfos", "/UserInfos"} {
		users := api.Group(userPath)
		{
			users.POST("/login", userHandler.Login)
			users.POST("/change-password", userHandler.ChangePassword)
			users.GET("/:id/history", userHandler.GetHistory)
			users.GET("/:id", userHandler.GetById)
			users.POST("", userHandler.Insert)
			users.PUT("/:id", userHandler.Update)
			users.POST("/search", userHandler.Search)
			users.GET("/get-all", userHandler.GetAll)
			users.POST("/logout", userHandler.Logout)
		}
	}

	// 3. Assets Controller
	for _, assetPath := range []string{"/assets", "/Assets"} {
		assets := api.Group(assetPath)
		{
			assets.POST("/search-grid", assetHandler.SearchGrid)
			assets.POST("/search", assetHandler.SearchGrid)
			assets.GET("/:id", assetHandler.GetById)
			assets.POST("", assetHandler.Insert)
			assets.PUT("/:id", assetHandler.Update)
			assets.DELETE("/:id", assetHandler.Delete)
			assets.POST("/export", assetHandler.Export)
			assets.POST("/import-file", assetHandler.ImportFile)
			assets.GET("/download-template", assetHandler.DownloadTemplate)
			assets.GET("/dashboard-metrics", assetHandler.GetDashboardMetrics)
			assets.GET("/generate-unit-code", assetHandler.GenerateUnitCode)
			assets.POST("/suggestions", assetHandler.GetSuggestions)
			assets.GET("/unassigned-sub-assets", assetHandler.GetUnassignedSubAssets)
			assets.POST("/unassigned-sub-assets", assetHandler.GetUnassignedSubAssets)
			assets.PUT("/:id/main-asset", assetHandler.UpdateMainAsset)
		}
	}

	// 4. Handovers Controller
	for _, handoverPath := range []string{"/handovers", "/Handovers"} {
		handovers := api.Group(handoverPath)
		{
			handovers.POST("/search", handoverHandler.Search)
			handovers.GET("/export/:id", handoverHandler.Export)
			handovers.POST("", handoverHandler.Insert)
			handovers.PUT("/:id", handoverHandler.Update)
			handovers.DELETE("/:id", handoverHandler.Delete)
		}
	}

	// 5. Generic Master Data Controllers
	registerGenericCrud(api, []string{"/departments", "/Departments"}, masterService.DeptRepo)
	registerGenericCrud(api, []string{"/types", "/Types"}, masterService.TypeRepo)
	registerGenericCrud(api, []string{"/deviceconfigs", "/DeviceConfigs", "/device-configs"}, masterService.DeviceRepo)
	registerGenericCrud(api, []string{"/disposals", "/Disposals"}, masterService.DisposalRepo)
	registerGenericCrud(api, []string{"/softwares", "/Softwares"}, masterService.SoftwareRepo)
	registerGenericCrud(api, []string{"/techspecs", "/TechSpecs", "/tech-specs"}, masterService.TechSpecRepo)
	registerGenericCrud(api, []string{"/statemasterdatas", "/StateMasterDatas", "/state-master-datas"}, masterService.StateDataRepo)
	registerGenericCrud(api, []string{"/itemmasterdatas", "/ItemMasterDatas", "/item-master-datas"}, masterService.ItemDataRepo)
	registerGenericCrud(api, []string{"/titleemployees", "/TitleEmployees", "/title-employees"}, masterService.TitleEmpRepo)
	registerGenericCrud(api, []string{"/problemreports", "/ProblemReports", "/problem-reports"}, masterService.ProblemRepo)
	registerGenericCrud(api, []string{"/assetslog", "/AssetsLog", "/assets-log"}, masterService.AssetsLogRepo)
	registerGenericCrud(api, []string{"/permissions", "/Permissions"}, masterService.PermissionRepo)

	// 6. AutoSync Controller
	for _, syncPath := range []string{"/autosync", "/AutoSync", "/auto-sync"} {
		syncRoutes := api.Group(syncPath)
		{
			syncRoutes.POST("/sync-all", syncHandler.SyncAll)
			syncRoutes.POST("/sync-ad", syncHandler.SyncAD)
			syncRoutes.POST("/sync-patchmanager", syncHandler.SyncPatchManager)
			syncRoutes.POST("/sync-trendmicro", syncHandler.SyncTrendMicro)
		}
	}

	return r
}

func registerGenericCrud[T any](rg *gin.RouterGroup, paths []string, repo *repository.GenericRepository[T]) {
	h := handlers.NewGenericHandler[T](repo)
	for _, path := range paths {
		g := rg.Group(path)
		{
			g.POST("/search", h.Search)
			g.GET("/get-all", h.GetAll)
			g.GET("/:id", h.GetById)
			g.POST("", h.Insert)
			g.PUT("/:id", h.Update)
			g.DELETE("/:id", h.Delete)
		}
	}
}
