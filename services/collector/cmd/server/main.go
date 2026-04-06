package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourname/pulse/collector/internal/circuit"
	"github.com/yourname/pulse/collector/internal/health"
	"github.com/yourname/pulse/collector/internal/ingest"
	"github.com/yourname/pulse/collector/internal/pipeline"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	port := getEnv("PORT", "8081")

	// Connexions avec retry (Ch.4 Availability)
	db, err := ingest.NewPostgresPool(getEnv("POSTGRES_DSN", ""))
	if err != nil {
		logger.Fatal("postgres failed", zap.Error(err))
	}
	defer db.Close()

	mq, err := ingest.NewRabbitMQ(getEnv("RABBITMQ_URL", ""))
	if err != nil {
		logger.Fatal("rabbitmq failed", zap.Error(err))
	}
	defer mq.Close()

	// Circuit Breaker PostgreSQL (Ch.4)
	cb := circuit.NewPostgresBreaker(logger)

	// Pipeline Pipe-and-Filter (Ch.13)
	p := pipeline.New(logger).
		Add(pipeline.ValidateFilter()).
		Add(pipeline.NormalizeFilter()).
		Add(pipeline.EnrichFilter()).
		Add(pipeline.PersistFilter(db, cb)).
		Add(pipeline.PublishFilter(mq))

	// Handlers
	ingestH := ingest.NewHandler(p, logger)
	healthH := health.NewChecker(db, mq, logger)

	// Router
	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(logger))
	router.POST("/ingest", ingestH.Handle)
	router.GET("/health", healthH.Handle)
	router.GET("/metrics", prometheusHandler())

	// Serveur HTTP avec graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("collector started", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
