package api

import (
	"net/http"

	"go-backend/configs"
	"go-backend/internals/api/handlers"
	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/ports"
	"go-backend/internals/infrastructure/excel"
	"go-backend/internals/infrastructure/pdf"
	"go-backend/internals/infrastructure/scheduler"
	"go-backend/internals/infrastructure/sso"
	"go-backend/internals/repository"
	"go-backend/internals/services"
	"go-backend/internals/websocket"
	"go-backend/pkg/cache"
	"go-backend/pkg/utils"

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

	// Live Runtime Profiling via net/http/pprof
	RegisterPprof(r)

	// 1. Infrastructure Adapters & Transaction Manager
	txManager := repository.NewTransactionManager(db)
	ssoProvider := sso.NewSSOProvider(&cfg.SSO)
	passwordHasher := utils.NewPasswordHasher()
	memCache := cache.GetCache()
	pdfGenerator := pdf.NewPDFGenerator()
	excelProcessor := excel.NewExcelProcessor("Templates")

	// 2. Repositories
	userRepo := repository.NewUserRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	handoverRepo := repository.NewHandoverRepository(db)
	disposalRepo := repository.NewDisposalRepository(db)
	problemRepo := repository.NewProblemReportRepository(db)
	assetsLogRepo := repository.NewAssetsLogRepository(db)
	deptRepo := repository.NewDepartmentRepository(db)
	itemDataRepo := repository.NewItemMasterDataRepository(db)
	autoSyncRepo := repository.NewAutoSyncRepository(db)

	typeRepo := repository.NewGenericRepository[domains.TYPE](db, "Type", []string{"Type", "TypeName", "VN"})
	deviceRepo := repository.NewGenericRepository[domains.DEVICECONFIG](db, "Device", []string{"Device", "Detail1", "Detail2", "Detail3"})
	softwareRepo := repository.NewGenericRepository[domains.SOFTWARE](db, "ID", []string{"UnitCode", "Software", "Version"})
	techSpecRepo := repository.NewGenericRepository[domains.KEYTECHSPEC](db, "Name", []string{"Name", "DescriptionTechSpecs"})
	stateDataRepo := repository.NewGenericRepository[domains.STATEMASTERDATA](db, "ID", []string{"ID", "State", "Description"})
	titleEmpRepo := repository.NewGenericRepository[domains.EmployeeTitle](db, "EmployeeCode", []string{"EmployeeCode", "Title"})
	permissionRepo := repository.NewGenericRepository[domains.PERMISSION](db, "UserID", []string{"UserID"})
	accessoryRepo := repository.NewGenericRepository[domains.ACCESSORY](db, "ID", []string{"UnitCode", "Description", "Brand", "Serial"})
	serviceRepo := repository.NewGenericRepository[domains.SERVICE](db, "MainAssets", []string{"MainAssets", "SubAssets", "ServiceLevel"})

	// 3. Use Cases / Services
	authService := services.NewAuthService(userRepo, ssoProvider, passwordHasher, memCache, cfg)
	userService := services.NewUserInfoService(userRepo, txManager, passwordHasher, memCache)
	assetService := services.NewAssetService(assetRepo, userRepo, authService, txManager, excelProcessor)
	handoverService := services.NewHandoverService(handoverRepo, assetRepo, userRepo, txManager, pdfGenerator)
	disposalService := services.NewDisposalService(disposalRepo, assetRepo, userRepo, authService, txManager, excelProcessor)
	problemService := services.NewProblemReportService(problemRepo, assetRepo, userRepo, pdfGenerator)
	assetsLogService := services.NewAssetsLogService(assetsLogRepo, userRepo, excelProcessor)
	deptService := services.NewDepartmentService(deptRepo, userRepo)
	typeService := services.NewTypeService(typeRepo, userRepo)
	permissionService := services.NewPermissionService(permissionRepo, userRepo, memCache)
	itemDataService := services.NewItemMasterDataService(itemDataRepo, userRepo)
	syncService := services.NewAutoSyncService(autoSyncRepo)

	deviceService := services.NewGenericService[domains.DEVICECONFIG](deviceRepo, userRepo, "DEVICECONFIG")
	softwareService := services.NewGenericService[domains.SOFTWARE](softwareRepo, userRepo, "SOFTWARE")
	techSpecService := services.NewGenericService[domains.KEYTECHSPEC](techSpecRepo, userRepo, "KEYTECHSPEC")
	stateDataService := services.NewGenericService[domains.STATEMASTERDATA](stateDataRepo, userRepo, "STATEMASTERDATA")
	titleEmpService := services.NewGenericService[domains.EmployeeTitle](titleEmpRepo, userRepo, "EmployeeTitle")
	accessoryService := services.NewGenericService[domains.ACCESSORY](accessoryRepo, userRepo, "ACCESSORY")
	serviceEntityService := services.NewGenericService[domains.SERVICE](serviceRepo, userRepo, "SERVICE")

	if cfg.AutoSync.Enabled {
		cronSched := scheduler.NewCronScheduler(syncService)
		cronSched.Start(cfg.AutoSync.CronSchedule)
	}

	// 4. Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, authService)
	assetHandler := handlers.NewAssetHandler(assetService)
	handoverHandler := handlers.NewHandoverHandler(handoverService)
	disposalHandler := handlers.NewDisposalHandler(disposalService)
	problemHandler := handlers.NewProblemReportHandler(problemService)
	assetsLogHandler := handlers.NewAssetsLogHandler(assetsLogService)
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
			users.DELETE("/:id", userHandler.Delete)
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
			assets.GET("/get-all", assetHandler.GetAll)
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
			handovers.GET("/get-all", handoverHandler.GetAll)
			handovers.GET("/export/:id", handoverHandler.Export)
			handovers.GET("/:id", handoverHandler.GetById)
			handovers.POST("", handoverHandler.Insert)
			handovers.PUT("/:id", handoverHandler.Update)
			handovers.DELETE("/:id", handoverHandler.Delete)
		}
	}

	// 5. Disposals Controller
	for _, dispPath := range []string{"/disposals", "/Disposals"} {
		disposals := api.Group(dispPath)
		{
			disposals.POST("/transfer-wfd", disposalHandler.TransferToWFD)
			disposals.POST("/confirm-dispose", disposalHandler.ConfirmDispose)
			disposals.GET("/batches", disposalHandler.GetBatches)
			disposals.GET("/batch-details/:idDisposal", disposalHandler.GetBatchDetails)
			disposals.POST("/check-unitcodes", disposalHandler.CheckUnitCodes)
			disposals.POST("/import-unitcodes", disposalHandler.ImportUnitCodes)
			disposals.POST("/search", disposalHandler.Search)
			disposals.GET("/get-all", disposalHandler.GetAll)
			disposals.GET("/:id", disposalHandler.GetById)
			disposals.POST("", disposalHandler.Insert)
			disposals.PUT("/:id", disposalHandler.Update)
			disposals.DELETE("/:id", disposalHandler.Delete)
		}
	}

	// 6. Problem Reports Controller
	for _, prbPath := range []string{"/problemreports", "/ProblemReports", "/problem-reports"} {
		prb := api.Group(prbPath)
		{
			prb.GET("/export/:id", problemHandler.Export)
			prb.POST("/search", problemHandler.Search)
			prb.GET("/get-all", problemHandler.GetAll)
			prb.GET("/:id", problemHandler.GetById)
			prb.POST("", problemHandler.Insert)
			prb.PUT("/:id", problemHandler.Update)
			prb.DELETE("/:id", problemHandler.Delete)
		}
	}

	// 7. AssetsLog Controller
	for _, logPath := range []string{"/assetslog", "/AssetsLog", "/assets-log"} {
		logs := api.Group(logPath)
		{
			logs.GET("/export", assetsLogHandler.Export)
			logs.POST("/search", assetsLogHandler.Search)
			logs.GET("/get-all", assetsLogHandler.GetAll)
			logs.GET("/:id", assetsLogHandler.GetById)
			logs.POST("", assetsLogHandler.Insert)
			logs.PUT("/:id", assetsLogHandler.Update)
			logs.DELETE("/:id", assetsLogHandler.Delete)
		}
	}

	// 8. Generic & Specialized Master Data Controllers
	registerGenericCrud(api, []string{"/departments", "/Departments"}, deptService)
	registerGenericCrud(api, []string{"/types", "/Types"}, typeService)
	registerGenericCrud(api, []string{"/deviceconfigs", "/DeviceConfigs", "/device-configs"}, deviceService)
	registerGenericCrud(api, []string{"/softwares", "/Softwares"}, softwareService)
	registerGenericCrud(api, []string{"/techspecs", "/TechSpecs", "/tech-specs", "/keytechspecs", "/KeyTechSpecs"}, techSpecService)
	registerGenericCrud(api, []string{"/statemasterdatas", "/StateMasterDatas", "/state-master-datas"}, stateDataService)
	registerGenericCrud(api, []string{"/itemmasterdatas", "/ItemMasterDatas", "/item-master-datas"}, itemDataService)
	registerGenericCrud(api, []string{"/titleemployees", "/TitleEmployees", "/title-employees"}, titleEmpService)
	registerGenericCrud(api, []string{"/permissions", "/Permissions"}, permissionService)
	registerGenericCrud(api, []string{"/accessories", "/Accessories"}, accessoryService)
	registerGenericCrud(api, []string{"/services", "/Services"}, serviceEntityService)

	// 9. AutoSync Controller
	for _, syncPath := range []string{"/autosync", "/AutoSync", "/auto-sync"} {
		syncRoutes := api.Group(syncPath)
		{
			syncRoutes.POST("/sync-all", syncHandler.SyncAll)
			syncRoutes.POST("/sync-ad", syncHandler.SyncAD)
			syncRoutes.POST("/sync-patchmanager", syncHandler.SyncPatchManager)
			syncRoutes.POST("/sync-trendmicro", syncHandler.SyncTrendMicro)
			syncRoutes.POST("/sync-medetail", syncHandler.SyncMEDetail)
			syncRoutes.GET("/ad", syncHandler.GetADAssets)
			syncRoutes.GET("/tm", syncHandler.GetTMAssets)
			syncRoutes.GET("/me", syncHandler.GetMEAssets)
			syncRoutes.GET("/medetail", syncHandler.GetMEDetails)
		}
	}

	return r
}

func registerGenericCrud[T any](rg *gin.RouterGroup, paths []string, useCase ports.GenericUseCase[T]) {
	h := handlers.NewGenericHandler[T](useCase)
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
