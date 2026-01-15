package main

import (
	"log"
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/handler"
	"easy-stream/internal/middleware"
	"easy-stream/internal/repository"
	"easy-stream/internal/service"
	"easy-stream/internal/storage"
	"easy-stream/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level)

	// 初始化数据库
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化 Redis
	rdb, err := repository.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Debug 模式下插入种子数据
	if cfg.Server.Mode == "debug" {
		if err := repository.SeedData(db); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		}
	}

	// 初始化 Repository
	streamRepo := repository.NewStreamRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 初始化 Service
	streamSvc := service.NewStreamService(streamRepo, rdb, cfg.ZLMediaKit)
	authSvc := service.NewAuthService(userRepo, rdb, cfg.JWT)

	// 初始化存储管理器
	var storageManager *storage.Manager
	if len(cfg.Storage.Targets) > 0 {
		storageManager, err = storage.NewManager(cfg.Storage)
		if err != nil {
			log.Printf("Warning: Failed to init storage manager: %v", err)
		}
	}

	// 初始化 Handler
	streamHandler := handler.NewStreamHandler(streamSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	hookHandler := handler.NewHookHandler(streamSvc, storageManager)
	systemHandler := handler.NewSystemHandler()

	// 启动定时任务：检查超时直播
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := streamSvc.CheckExpiredStreams(); err != nil {
				log.Printf("Failed to check expired streams: %v", err)
			}
		}
	}()

	// 设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 中间件
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// 路由
	api := r.Group("/api/v1")
	{
		// 认证接口
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", middleware.Auth(cfg.JWT.Secret), authHandler.Profile)
		}

		// 推流管理接口
		streams := api.Group("/streams")
		{
			// 游客可访问（公开直播列表）
			streams.GET("", streamHandler.List)
			streams.GET("/:key", streamHandler.Get)
			streams.POST("/:key/verify", streamHandler.VerifyPassword)

			// 管理员接口（需要认证）
			admin := streams.Group("")
			admin.Use(middleware.Auth(cfg.JWT.Secret))
			{
				admin.POST("", streamHandler.Create)
				admin.GET("/id/:id", streamHandler.GetByID) // 通过 ID 获取
				admin.PUT("/:key", streamHandler.Update)
				admin.DELETE("/:key", streamHandler.Delete)
				admin.POST("/:key/kick", streamHandler.Kick)
			}
		}

		// 系统接口
		system := api.Group("/system")
		{
			system.GET("/health", systemHandler.Health)
			system.GET("/stats", middleware.Auth(cfg.JWT.Secret), systemHandler.Stats)
		}

		// ZLMediaKit Hook 接口
		hooks := api.Group("/hooks")
		{
			hooks.POST("/on_publish", hookHandler.OnPublish)
			hooks.POST("/on_unpublish", hookHandler.OnUnpublish)
			hooks.POST("/on_flow_report", hookHandler.OnFlowReport)
			hooks.POST("/on_stream_none_reader", hookHandler.OnStreamNoneReader)
			hooks.POST("/on_play", hookHandler.OnPlay)
			hooks.POST("/on_player_disconnect", hookHandler.OnPlayerDisconnect)
			hooks.POST("/on_record_mp4", hookHandler.OnRecordMP4)
		}
	}

	// 启动服务
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
