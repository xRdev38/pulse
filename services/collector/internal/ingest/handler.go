package ingest

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yourname/pulse/collector/internal/pipeline"
	"go.uber.org/zap"
)

// Handler — reçoit POST /ingest et exécute le pipeline
type Handler struct {
	pipeline *pipeline.Pipeline
	logger   *zap.Logger
}

func NewHandler(p *pipeline.Pipeline, logger *zap.Logger) *Handler {
	return &Handler{pipeline: p, logger: logger}
}

func (h *Handler) Handle(c *gin.Context) {
	var payload pipeline.MetricPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON", "detail": err.Error()})
		return
	}

	// TenantID injecté par le Gateway via header (jamais par le client)
	payload.TenantID = c.GetHeader("X-Tenant-ID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.pipeline.Execute(ctx, &payload); err != nil {
		h.logger.Error("pipeline failed", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "processing failed", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
		"metric": payload.NormalizedName,
		"ts":     payload.TimestampNs,
	})
}

// NewPostgresPool — pool de connexions pgx avec retry backoff exponentiel (Ch.4)
func NewPostgresPool(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	for i := 1; i <= 5; i++ {
		pool, err = pgxpool.NewWithConfig(context.Background(), cfg)
		if err == nil {
			if err = pool.Ping(context.Background()); err == nil {
				return pool, nil
			}
			pool.Close()
		}
		time.Sleep(time.Duration(math.Pow(2, float64(i-1))) * time.Second)
	}
	return nil, err
}

// NewRabbitMQ — connexion AMQP avec retry backoff exponentiel (Ch.4)
func NewRabbitMQ(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for i := 1; i <= 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		time.Sleep(time.Duration(math.Pow(2, float64(i-1))) * time.Second)
	}
	return nil, err
}
