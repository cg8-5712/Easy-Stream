package main

import (
	"log"

	"easy-stream/internal/config"
	"easy-stream/internal/handler"
	"easy-stream/internal/middleware"
	"easy-stream/internal/repository"
	"easy-stream/internal/service"
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

	// 初始化 Repository
	streamRepo := repository.NewStreamRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 初始化 Service
	streamSvc := service.NewStreamService(streamRepo, cfg.ZLMediaKit)
	authSvc := service.NewAuthService(userRepo, cfg.JWT)

	// 初始化 Handler
	streamHandler := handler.NewStreamHandler(streamSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	hookHandler := handler.NewHookHandler(streamSvc)
	systemHandler := handler.NewSystemHandler()

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
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", middleware.Auth(cfg.JWT.Secret), authHandler.Profile)
		}

		// 推流管理接口 (需要认证)
		streams := api.Group("/streams")
		streams.Use(middleware.Auth(cfg.JWT.Secret))
		{
			streams.GET("", streamHandler.List)
			streams.POST("", streamHandler.Create)
			streams.GET("/:key", streamHandler.Get)
			streams.PUT("/:key", streamHandler.Update)
			streams.DELETE("/:key", streamHandler.Delete)
			streams.POST("/:key/kick", streamHandler.Kick)
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
		}
	}

	// 启动服务
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
