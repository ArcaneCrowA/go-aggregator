package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/data"
	"github.com/ArcaneCrowA/go-aggregator/internal/handlers"
	"github.com/ArcaneCrowA/go-aggregator/internal/middleware"
	"github.com/ArcaneCrowA/go-aggregator/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"go.uber.org/zap"
)

type app struct {
	log  *zap.Logger
	cfg  config
	pool *pgxpool.Pool
}

type config struct {
	address  string
	dbString string
}

func Start(log *zap.Logger) {
	app := &app{log: log}

	app.loadConfig()
	app.runMigrations()
	close := app.connectDB()
	defer close()

	app.startServer()
}

func (a *app) loadConfig() {
	addr := os.Getenv("ADDRESS")
	if addr == "" {
		a.log.Warn("ADDRESS env not set, use default", zap.String("default", ":8080"))
		a.cfg.address = ":8080"
	} else {
		a.cfg.address = addr
	}

	dbString := os.Getenv("GOOSE_DBSTRING")
	if dbString == "" {
		a.log.Warn("GOOSE_DBSTRING env not set, use default", zap.String("default", "postgres://postgres:secret@localhost:5432/aggregator?sslmode=disable"))
		a.cfg.dbString = "postgres://postgres:secret@localhost:5432/aggregator?sslmode=disable"
	} else {
		a.cfg.dbString = dbString
	}

	a.log.Info("config loaded", zap.String("address", a.cfg.address))
}

func (a *app) runMigrations() {
	sqlDB, err := sql.Open("pgx", a.cfg.dbString)
	if err != nil {
		a.log.Fatal("unable to open database for migrations", zap.Error(err))
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		a.log.Fatal("failed to set goose dialect", zap.Error(err))
	}

	migrationDir := os.Getenv("GOOSE_MIGRATION_DIR")
	if migrationDir == "" {
		migrationDir = "./migrations"
	}

	if err := goose.Up(sqlDB, migrationDir); err != nil {
		a.log.Fatal("migration failed", zap.Error(err))
	}

	a.log.Info("migrations applied successfully")
}

func (a *app) connectDB() func() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, a.cfg.dbString)
	if err != nil {
		a.log.Fatal("unable to connect to database", zap.Error(err))
	}

	a.pool = pool

	return pool.Close
}

func (a *app) startServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.SetTrustedProxies(nil)

	router.Use(middleware.ZapLogger(a.log))
	router.Use(gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		a.log.Info("ping success")
		c.JSON(http.StatusOK, "pong")
	})

	{
		repo := data.New(a.pool)
		svc := service.New(repo)
		handler := handlers.New(a.log, svc)

		g := router.Group("/service")

		g.POST("/", handler.AddSubscription)
		g.GET("/", handler.GetSubscriptionsFilter)
	}

	server := &http.Server{
		Addr:    a.cfg.address,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		a.log.Info("server shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			a.log.Fatal("failed to close server", zap.Error(err))
		}
	}()

	a.log.Info("server starts", zap.String("address", a.cfg.address))
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			a.log.Info("server closed")
		} else {
			a.log.Error("failed to serve", zap.Error(err))
		}
	}

}
