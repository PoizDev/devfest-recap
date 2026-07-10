package main

import (
	"context"
	"dfrecap/controllers"
	"dfrecap/db"
	"dfrecap/middlewares"
	"dfrecap/services"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"dfrecap/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var conf *config.Config

func initialize() {
	var err error
	conf, err = config.LoadConfig("conf.yaml")
	if err != nil {
		panic(err)
	}
	config.InitLogger(conf.Server.ActiveLevel)
	db.Connect(conf.Database, conf.Server.EnvPath)
	db.ConnectRedis(conf.Redis)
	db.RunMigrations()
	db.SeedAdminUser()
	controllers.SetJWTConfig(conf.JWT)
	controllers.SetRecapMode(conf.Server.RecapMode)
	config.Log.Sync()
}

func main() {
	initialize()

	services.StartCronJobs()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://recap.devfestbursa.com", "https://recap.www.devfestbursa.com", "https://devfestbursa.com", "https://www.devfestbursa.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           600,
	}))
	r.Use(middlewares.RateLimitMiddleware(conf.RateLimit))

	authRoutes := r.Group("api/auth")
	{
		authRoutes.POST("register", controllers.Register)
		authRoutes.POST("login", controllers.Login)
		authRoutes.POST("logout", controllers.Logout)
	}

	r.GET("api/status", controllers.GetStatus)
	r.GET("api/qrcode", middlewares.GenerateQRCode)

	api := r.Group("api")
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("scan", controllers.Scan)
		api.GET("recap", controllers.GetRecap)
		api.GET("leaderboard", controllers.GetLeaderboard)
	}

	admin := r.Group("api/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	{
		admin.POST("sessions", controllers.CreateSession)
		admin.GET("sessions", controllers.ListSessions)
		admin.GET("sessions/:id/qrcode", controllers.GetSessionQRCode)
		admin.DELETE("sessions/:id", controllers.DeleteSession)
		admin.GET("export-recaps", controllers.ExportRecaps)
		admin.POST("sync-from-devtv", controllers.SyncFromDevTV)

		admin.GET("analytics/overview", controllers.GetAnalyticsOverview)
		admin.GET("analytics/tags", controllers.GetAnalyticsTags)
		admin.GET("analytics/rooms", controllers.GetAnalyticsRooms)
	}

	srv := &http.Server{
		Addr:    conf.Server.Port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Log.Fatal("Sunucu başlatılamadı", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sign := <-quit
	config.Log.Info("Sunucu kapatılıyor", zap.String("signal", sign.String()))

	services.StopCronJobs()

	ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		config.Log.Fatal("Sunucu kapatılamadı", zap.Error(err))
	}
	config.Log.Info("Sunucu başarıyla kapatıldı")
}
